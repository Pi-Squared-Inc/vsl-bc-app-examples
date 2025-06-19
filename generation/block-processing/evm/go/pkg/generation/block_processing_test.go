package generation

import (
	basemodels "base/pkg/models"
	"context"
	"encoding/json"
	"log"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	// The block number to generate the claim for, you can get one from the [Ethereum Explorer](https://etherscan.io/blocks),
	// please pick the one that in recent 128 blocks as the rpc node will prune the blocks.
	blockNumber = 22296380

	// The node RPC endpoint to use
	rpcEndpoint = "http://127.0.0.1:8545"
)

// Note: This test is used to generate the block processing claim related mock files
//
// After running this test, you can find the mock files in the current directory:
// - block_processing_test_mock_block.json: The block in json format
// - block_processing_test_mock_witness.json: The witness in json format
// - block_processing_test_mock_claim.json: The claim
// - block_processing_test_mock_verification_context.json: The verification context
func TestGenerateToMockFiles(t *testing.T) {
	ctx := context.Background()

	rpcClient, err := rpc.Dial(rpcEndpoint)
	if err != nil {
		log.Fatalf("Failed to create RPC client: %+v", err)
	}

	ethRPCClient := ethclient.NewClient(rpcClient)

	blockNumberBig := big.NewInt(blockNumber)

	claim, verificationContext, err := Generate(ethRPCClient, blockNumberBig)
	if err != nil {
		t.Fatalf("error generating block claim: %+v", err)
	}

	var blockRaw interface{}
	err = ethRPCClient.Client().CallContext(ctx, &blockRaw, "eth_getBlockByNumber", hexutil.EncodeBig(blockNumberBig), true)
	if err != nil {
		t.Fatalf("error getting block: %+v", err)
	}

	blockJSON, err := json.Marshal(blockRaw)
	if err != nil {
		t.Fatalf("error encoding block: %+v", err)
	}

	err = os.WriteFile("block_processing_test_mock_block.json", blockJSON, 0644)
	if err != nil {
		t.Fatalf("error writing block to file: %+v", err)
	}

	block, err := ethRPCClient.BlockByNumber(ctx, blockNumberBig)
	if err != nil {
		t.Fatalf("error getting block: %+v", err)
	}

	// Write witness RLP to file
	var gethWitness *basemodels.GethWitness
	err = ethRPCClient.Client().CallContext(ctx, &gethWitness, "debug_executionWitness", block.Hash())
	if err != nil {
		t.Fatalf("error getting geth witness: %+v", err)
	}

	gethWitnessJSON, err := json.Marshal(gethWitness)
	if err != nil {
		t.Fatalf("error encoding witness: %+v", err)
	}

	err = os.WriteFile("block_processing_test_mock_witness.json", gethWitnessJSON, 0644)
	if err != nil {
		t.Fatalf("error writing witness to file: %+v", err)
	}

	// Write claim json to file
	claimJSON, err := json.Marshal(claim)
	if err != nil {
		t.Fatalf("error encoding claim: %+v", err)
	}

	err = os.WriteFile("block_processing_test_mock_claim.json", claimJSON, 0644)
	if err != nil {
		t.Fatalf("error writing claim to file: %+v", err)
	}

	// Write verification context json to file
	verificationContextJSON, err := json.Marshal(verificationContext)
	if err != nil {
		t.Fatalf("error encoding verification context: %+v", err)
	}

	err = os.WriteFile("block_processing_test_mock_verification_context.json", verificationContextJSON, 0644)
	if err != nil {
		t.Fatalf("error writing verification context to file: %+v", err)
	}
}
