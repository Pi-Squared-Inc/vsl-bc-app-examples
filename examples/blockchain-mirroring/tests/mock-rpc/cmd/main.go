package main

import (
	"encoding/json"
	"log"
	"math/big"
	"os"

	basemodels "base/pkg/models"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/joho/godotenv"
)

const mockChainID = 1

func loadMockData() (int64, map[string]interface{}, types.Header, map[string]interface{}, hexutil.Bytes, *basemodels.GethWitness, map[string]interface{}) {
	blockJSON, err := os.ReadFile("mock_block.json")
	if err != nil {
		log.Fatalf("Failed to read mock_block.json: %v", err)
	}
	var blockHeader types.Header
	if err := json.Unmarshal(blockJSON, &blockHeader); err != nil {
		log.Fatalf("Failed to decode block header JSON: %v", err)
	}
	var block map[string]interface{}
	if err := json.Unmarshal(blockJSON, &block); err != nil {
		log.Fatalf("Failed to decode block JSON: %v", err)
	}

	gethWitnessJSON, err := os.ReadFile("mock_geth_witness.json")
	if err != nil {
		log.Fatalf("Failed to read mock_geth_witness.json: %v", err)
	}
	var gethWitness basemodels.GethWitness
	if err := json.Unmarshal(gethWitnessJSON, &gethWitness); err != nil {
		log.Fatalf("Failed to decode witness JSON: %v", err)
	}

	chainConfigJSON, err := os.ReadFile("mock_chain_config.json")
	if err != nil {
		log.Fatalf("Failed to read mock_chain_config.json: %v", err)
	}

	var chainConfig map[string]interface{}
	if err := json.Unmarshal([]byte(chainConfigJSON), &chainConfig); err != nil {
		log.Fatalf("Failed to decode chain config JSON: %v", err)
	}

	rawBlock, err := os.ReadFile("mock_raw_block.txt")
	if err != nil {
		log.Fatalf("Failed to read mock_raw_block.txt: %v", err)
	}

	rethWitnessJSON, err := os.ReadFile("mock_reth_witness.json")
	if err != nil {
		log.Fatalf("Failed to read mock_reth_witness.json: %v", err)
	}
	var rethWitness map[string]interface{}
	if err := json.Unmarshal(rethWitnessJSON, &rethWitness); err != nil {
		log.Fatalf("Failed to decode reth witness JSON: %v", err)
	}

	log.Printf("Successfully loaded mock data (ChainID: %d)", mockChainID)
	return mockChainID, block, blockHeader, chainConfig, rawBlock, &gethWitness, rethWitness
}

func setupRPCServer(mockService *MockRPCService) *rpc.Server {
	server := rpc.NewServer()
	if err := server.RegisterName("eth", mockService); err != nil {
		log.Fatalf("Failed to register mock service (eth): %v", err)
	}
	if err := server.RegisterName("debug", mockService); err != nil {
		log.Fatalf("Failed to register mock service (debug): %v", err)
	}
	return server
}

func main() {
	err := godotenv.Load(".env") // Load .env file if present
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Error loading .env file: %v", err)
	}

	mockChainID, mockBlock, mockBlockHeader, chainConfig, rawBlock, gethWitness, rethWitness := loadMockData()

	mockService := NewMockRPCService(
		(*hexutil.Big)(big.NewInt(mockChainID)),
		mockBlock,
		mockBlockHeader,
		rawBlock,
		gethWitness,
		rethWitness,
		chainConfig,
	)

	rpcServer := setupRPCServer(mockService)

	startHTTPServer(rpcServer) // non-blocking

	// Start the WebSocket server (blocking)
	// This will keep the main goroutine alive
	startWebSocketServer(mockService, mockBlockHeader)
}
