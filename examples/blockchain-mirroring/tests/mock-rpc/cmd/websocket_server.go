package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type jsonrpcMessage struct {
	Version string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *jsonError      `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

type jsonError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type subscriptionNotificationParams struct {
	Subscription string      `json:"subscription"`
	Result       interface{} `json:"result"`
}

type activeSubscription struct {
	id   string
	conn *websocket.Conn
}

var (
	subscriptions      = make(map[string]*activeSubscription)
	subscriptionsMutex sync.Mutex
	nextSubscriptionID int64 = 0
)

func generateSubID() string {
	subscriptionsMutex.Lock()
	defer subscriptionsMutex.Unlock()
	nextSubscriptionID++
	return fmt.Sprintf("0x%x", nextSubscriptionID)
}

func addSubscription(sub *activeSubscription) {
	subscriptionsMutex.Lock()
	defer subscriptionsMutex.Unlock()
	subscriptions[sub.id] = sub
	log.Printf("WS Sub Add: ID=%s, Conn=%p", sub.id, sub.conn)
}

func removeSubscription(id string) {
	subscriptionsMutex.Lock()
	defer subscriptionsMutex.Unlock()
	delete(subscriptions, id)
	log.Printf("WS Sub Remove: ID=%s", id)
}

func startWebSocketServer(mockService *MockRPCService, blockHeader types.Header) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WS Upgrade failed: %v", err)
			return
		}
		defer conn.Close()
		log.Printf("WS Connection established: %s", conn.RemoteAddr())
		handleWebSocketConnection(conn, mockService, blockHeader)
		log.Printf("WS Connection closed: %s", conn.RemoteAddr())
	})
	server := &http.Server{Addr: ":8546", Handler: mux}
	log.Printf("Starting WebSocket RPC server on :8546")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("WebSocket RPC server failed: %v", err)
	}
}

func handleWebSocketConnection(conn *websocket.Conn, mockService *MockRPCService, blockHeader types.Header) {
	var currentSubID string // Track subscription ID for this connection
	defer func() {
		if currentSubID != "" {
			removeSubscription(currentSubID)
		}
	}()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WS Read Error: %v", err)
			} else {
				log.Printf("WS Client Closed: %v", err)
			}
			break
		}

		if messageType != websocket.TextMessage && messageType != websocket.BinaryMessage {
			continue
		}

		log.Printf("WS Received raw: %s", string(p))
		var req jsonrpcMessage
		if err := json.Unmarshal(p, &req); err != nil {
			log.Printf("WS JSON Unmarshal Error: %v", err)
			continue
		}

		log.Printf("WS Parsed Request: ID=%s, Method=%s", string(req.ID), req.Method)

		resp := processWebSocketRequest(req, conn, mockService, blockHeader, &currentSubID)

		// Send response if one was generated (and not handled internally, like subscribe)
		if resp != nil {
			responseBytes, _ := json.Marshal(resp)
			log.Printf("WS Sending Response: %s", string(responseBytes))
			if err := conn.WriteMessage(websocket.TextMessage, responseBytes); err != nil {
				log.Printf("WS Write Error: %v", err)
				break // Stop processing on write error
			}
		}
	}
}

// processWebSocketRequest handles a single JSON-RPC request received over WebSocket.
// It modifies currentSubID if a subscription is created.
// It returns a response message to send back, or nil if the response was handled internally (e.g., subscribe).
func processWebSocketRequest(req jsonrpcMessage, conn *websocket.Conn, mockService *MockRPCService, blockHeader types.Header, currentSubID *string) *jsonrpcMessage {
	resp := jsonrpcMessage{Version: "2.0", ID: req.ID}

	switch req.Method {
	case "eth_chainId":
		chainId, err := mockService.ChainId()
		if err != nil {
			log.Printf("WS eth_chainId call failed: %v", err)
			resp.Error = &jsonError{Code: -32000, Message: "Internal error calling ChainId"}
		} else {
			resultBytes, _ := json.Marshal(chainId)
			resp.Result = resultBytes
			log.Printf("WS eth_chainId result: %s", string(resultBytes))
		}

	case "eth_subscribe":
		var params []string
		if err := json.Unmarshal(req.Params, &params); err != nil || len(params) == 0 {
			log.Printf("WS Invalid eth_subscribe params: %v", err)
			resp.Error = &jsonError{Code: -32602, Message: "Invalid params"}
		} else if params[0] == "newHeads" {
			subID := generateSubID()
			*currentSubID = subID // Update the caller's tracked subscription ID
			sub := &activeSubscription{id: subID, conn: conn}
			addSubscription(sub)

			resultBytes, _ := json.Marshal(subID)
			resp.Result = resultBytes
			log.Printf("WS Sending subscription ID: %s", subID)

			// Send the subscription confirmation response immediately
			responseBytes, _ := json.Marshal(resp)
			if err := conn.WriteMessage(websocket.TextMessage, responseBytes); err != nil {
				log.Printf("WS Error writing subscription response: %v", err)
				removeSubscription(subID) // Clean up if sending fails
				*currentSubID = ""        // Untrack subID
				return nil                // Don't attempt to send another response
			}

			return nil // Indicate response already handled

		} else {
			log.Printf("WS Unsupported subscription type: %s", params[0])
			resp.Error = &jsonError{Code: -32601, Message: "Unsupported subscription type"}
		}

	case "eth_unsubscribe":
		log.Printf("WS Received eth_unsubscribe (Not Implemented)")
		resp.Error = &jsonError{Code: -32601, Message: "Method not found (unsubscribe not implemented)"}

	case "notify_new_head":
		for _, v := range subscriptions {
			go sendNewHeadsNotification(v, blockHeader)
		}
		log.Printf("WS Received notify_new_head call")
		return nil

	default:
		log.Printf("WS Received unhandled method '%s'", req.Method)
		resp.Error = &jsonError{Code: -32601, Message: fmt.Sprintf("Method '%s' not supported over WebSocket", req.Method)}
	}

	return &resp
}

func sendNewHeadsNotification(sub *activeSubscription, blockHeader types.Header) {
	// Construct and send the notification
	notificationParams := subscriptionNotificationParams{Subscription: sub.id, Result: blockHeader}
	paramsBytes, _ := json.Marshal(notificationParams)
	notificationMsg := jsonrpcMessage{
		Version: "2.0",
		Method:  "eth_subscription",
		Params:  paramsBytes,
	}
	notificationBytes, _ := json.Marshal(notificationMsg)

	log.Printf("WS Notifier (Sub %s): Sending newHeads notification: %s", sub.id, string(notificationBytes))

	// Use WriteMessage for thread-safety (Gorilla handles mutex internally per connection)
	err := sub.conn.WriteMessage(websocket.TextMessage, notificationBytes)
	if err != nil {
		log.Printf("WS Notifier (Sub %s): Error sending notification: %v. Removing subscription.", sub.id, err)
		removeSubscription(sub.id) // Remove on failure
	} else {
		log.Printf("WS Notifier (Sub %s): Successfully sent notification.", sub.id)
	}
}
