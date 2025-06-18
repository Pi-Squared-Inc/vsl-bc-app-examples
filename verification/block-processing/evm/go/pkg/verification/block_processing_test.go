package verification

import (
	"encoding/json"
	"generation-block-processing-evm/pkg/models"
	"io"
	"os"
	"testing"
)

// Note: if you want to create the mock files, please check `generation/block-processing/evm/go/pkg/generation/block_processing_test.go`
//
// Necessary mock files:
// - block_processing_test_mock_claim.json: mock claim json file
// - block_processing_test_mock_verification_context.json: mock verification context json file
func TestVerify(t *testing.T) {
	// Load mock claim json from file
	mockClaimFile, err := os.Open("./block_processing_test_mock_claim.json")
	if err != nil {
		t.Fatalf("Failed to open mock claim file: %v", err)
	}
	defer mockClaimFile.Close()

	mockClaimBytes, err := io.ReadAll(mockClaimFile)
	if err != nil {
		t.Fatalf("Failed to read mock claim file: %v", err)
	}

	var mockClaim models.EVMBlockProcessingClaim
	err = json.Unmarshal(mockClaimBytes, &mockClaim)
	if err != nil {
		t.Fatalf("Failed to unmarshal mock claim: %v", err)
	}

	// Load mock verification context json from file
	mockVerificationContextFile, err := os.Open("./block_processing_test_mock_verification_context.json")
	if err != nil {
		t.Fatalf("Failed to open mock verification context file: %v", err)
	}
	defer mockVerificationContextFile.Close()

	mockVerificationContextBytes, err := io.ReadAll(mockVerificationContextFile)
	if err != nil {
		t.Fatalf("Failed to read mock verification context file: %v", err)
	}

	var mockVerificationContext models.EVMBlockProcessingClaimVerificationContext
	err = json.Unmarshal(mockVerificationContextBytes, &mockVerificationContext)
	if err != nil {
		t.Fatalf("Failed to unmarshal mock verification context: %v", err)
	}

	err = Verify(&mockClaim, &mockVerificationContext)
	if err != nil {
		t.Fatalf("Failed to validate block processing claim: %v", err)
	}

}
