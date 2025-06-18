package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/gorilla/websocket"
)

func startHTTPServer(rpcServer *rpc.Server) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Reject WebSocket upgrade requests on HTTP port
		if strings.ToLower(r.Header.Get("Upgrade")) == "websocket" {
			http.Error(w, "WebSocket connections only accepted on port 8546", http.StatusMethodNotAllowed)
			return
		}
		// Basic CORS for testing
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodPost {
			rpcServer.ServeHTTP(w, r)
		} else {
			http.Error(w, "Method not allowed on port 8545 (only POST/OPTIONS)", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/notify", func(w http.ResponseWriter, r *http.Request) {
		// Basic CORS for testing
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodPost {
			// Setup WebSocket client outside the handler
			conn, _, err := websocket.DefaultDialer.Dial("ws://[::1]:8546", nil)
			if err != nil {
				log.Fatalf("Failed to connect to websocket: %v", err)
			}
			defer conn.Close()
			log.Println("Connected to websocket")
			err = conn.WriteMessage(websocket.TextMessage, []byte("{\"jsonrpc\": \"2.0\", \"id\": 3, \"method\": \"notify_new_head\", \"params\": {}}"))
			if err != nil {
				log.Printf("Failed to send message: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Failed to send message"))
				return
			}

			log.Println("Message sent successfully to websocket")

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Message sent successfully"))
		} else {
			http.Error(w, "Method not allowed on /notify (only POST/OPTIONS)", http.StatusMethodNotAllowed)
		}
	})
	server := &http.Server{Addr: ":8545", Handler: mux}
	log.Printf("Starting HTTP RPC server on :8545")
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP RPC server failed: %v", err)
		}
	}()
}
