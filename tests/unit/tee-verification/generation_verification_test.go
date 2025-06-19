package verification

import (
	claims "base-tee/pkg/abstract_types"
	"encoding/json"
	"generation/pkg/generation"
	"os"
	"path/filepath"
	"strings"
	"testing"
	verification "verification/pkg/verification"
)

type JSONContext struct {
	Report []byte `json:"report"`
	Nonce  []byte `json:"nonce"`
}

func TestGenerateVerifyClaims(t *testing.T) {
	sampleDir := "sample"
	files, err := os.ReadDir(sampleDir)
	if err != nil {
		t.Fatalf("Failed to read sample directory: %v", err)
	}

	// Map to store pairs: prefix -> {claim, proof}
	type pair struct {
		claim string
		proof string
	}
	pairs := make(map[string]*pair)

	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(name, "_mock_claim.json") {
			prefix := strings.TrimSuffix(name, "_mock_claim.json")
			if pairs[prefix] == nil {
				pairs[prefix] = &pair{}
			}
			pairs[prefix].claim = filepath.Join(sampleDir, name)
		} else if strings.HasSuffix(name, "_mock_proof.json") {
			prefix := strings.TrimSuffix(name, "_mock_proof.json")
			if pairs[prefix] == nil {
				pairs[prefix] = &pair{}
			}
			pairs[prefix].proof = filepath.Join(sampleDir, name)
		}
	}

	for prefix, p := range pairs {
		if p.claim == "" || p.proof == "" {
			t.Logf("Skipping incomplete pair for prefix %s", prefix)
			continue
		}
		t.Logf("Generating and verifying claim and proof from program - %s", prefix)
		t.Run(prefix, func(t *testing.T) {
			// Load and test as in your existing code, using p.claim and p.proof
			claimData, err := os.ReadFile(p.claim)
			if err != nil {
				t.Fatalf("Failed to read claim file: %v", err)
			}
			var parsedMockClaim claims.TEEComputationClaim
			if err := json.Unmarshal(claimData, &parsedMockClaim); err != nil {
				t.Fatalf("Failed to unmarshal claim: %v", err)
			}

			proofData, err := os.ReadFile(p.proof)
			if err != nil {
				t.Fatalf("Failed to read proof file: %v", err)
			}
			var mockVerificationContext JSONContext
			if err := json.Unmarshal(proofData, &mockVerificationContext); err != nil {
				t.Fatalf("Failed to unmarshal proof: %v", err)
			}

			mockClaim, mockContext, err := generation.GenerateTEEComputationClaim(parsedMockClaim.Computation, parsedMockClaim.Input, parsedMockClaim.Result, mockVerificationContext.Report, parsedMockClaim.Nonce)
			if err != nil {
				t.Fatalf("Failed to generate TEE computation claim: %v", err)
			}

			err = verification.VerifyTEEComputationClaim(mockClaim, mockContext)
			if err != nil {
				t.Fatalf("Failed to verify TEE computation claim: %v", err)
			}
			t.Logf("Successfully verified claim and proof for program - %s", prefix)
		})
	}
}
