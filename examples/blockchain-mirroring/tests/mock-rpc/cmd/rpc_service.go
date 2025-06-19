package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	basemodels "base/pkg/models"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rpc"
)

// MockRPCService holds the mock data and configuration.
// MockBlock now holds the raw JSON data (map[string]interface{}).
type MockRPCService struct {
	MockBlock       map[string]interface{} // Stores the raw JSON map[string]interface{}
	MockBlockHeader types.Header
	ChainID         *hexutil.Big
	GethWitness     *basemodels.GethWitness
	RawBlock        hexutil.Bytes // RLP encoded block data
	chainConfig     map[string]interface{}
	RethWitness     map[string]interface{}

	mu sync.Mutex
}

// NewMockRPCService creates a new mock RPC service.
// initialBlockData should be the direct JSON response (map[string]interface{}) from an RPC call.
func NewMockRPCService(chainID *hexutil.Big, mockBlock map[string]interface{}, mockBlockHeader types.Header, rawBlock hexutil.Bytes, gethWitness *basemodels.GethWitness, rethWitness map[string]interface{}, chainConfig map[string]interface{}) *MockRPCService {
	if mockBlock == nil {
		log.Println("Error: Initial mock block data provided to NewMockRPCService is nil.")
	}

	service := &MockRPCService{
		ChainID:         chainID,
		RawBlock:        rawBlock,
		MockBlock:       mockBlock,
		MockBlockHeader: mockBlockHeader,
		GethWitness:     gethWitness,
		RethWitness:     rethWitness,
		chainConfig:     chainConfig,
	}

	return service
}

// GetBlockByNumber is the RPC method handler for eth_getBlockByNumber.
// It now returns the stored raw block data (map[string]interface{}).
func (s *MockRPCService) GetBlockByNumber(ctx context.Context, blockNumber rpc.BlockNumber, fullTx bool) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.MockBlock, nil
}

// ChainId is the RPC method handler for eth_chainId.
func (s *MockRPCService) ChainId() (*hexutil.Big, error) {
	// No lock needed as ChainID is read-only after init
	log.Printf("Received eth_chainId request, returning chainId %s", s.ChainID.String())
	return s.ChainID, nil
}

// ExecutionWitness is the RPC method handler for debug_executionWitness.
// It returns the witness data formatted exactly as basemodels.GethWitness.
func (s *MockRPCService) ExecutionWitness(ctx context.Context, txHash common.Hash) (*basemodels.GethWitness, error) {
	log.Printf("Received debug_executionWitness request for tx %s", txHash.Hex())
	if s.GethWitness == nil {
		log.Println("Error: Mock witness is not loaded for debug_executionWitness.")
		return nil, fmt.Errorf("mock witness not loaded")
	}

	log.Printf("Returning mock witness formatted as basemodels.GethWitness.")
	return s.GethWitness, nil
}

// GetRawBlock is the RPC method handler for debug_getRawBlock.
// It returns the raw RLP-encoded block data.
func (s *MockRPCService) GetRawBlock(ctx context.Context, blockNrOrHash rpc.BlockNumberOrHash) (hexutil.Bytes, error) {
	log.Printf("Received debug_getRawBlock request for block number %s", blockNrOrHash.String())

	return s.RawBlock, nil
}

// ChainConfig is the RPC method handler for eth_chainConfig.
// It returns the chain configuration as a map[string]interface{}.
func (s *MockRPCService) ChainConfig(ctx context.Context) (map[string]interface{}, error) {
	log.Printf("Received eth_chainConfig request")
	return s.chainConfig, nil
}

// ExecutionWitnessByBlockHash is the RPC method handler for debug_executionWitnessByBlockHash.
// It returns the witness data for a given block hash formatted as a map[string]interface{}.
func (s *MockRPCService) ExecutionWitnessByBlockHash(ctx context.Context, blockHash common.Hash) (map[string]interface{}, error) {
	log.Printf("Received debug_executionWitnessByBlockHash request for block hash %s", blockHash.Hex())

	return s.RethWitness, nil
}
