package client

import (
	types "base-tee/pkg/abstract_types"
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"generation/pkg/generation"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type AttesterQuery struct {
	ClaimType   string            `json:"type"`
	Computation types.Computation `json:"computation"`
	Input       []string          `json:"input"`
	Nonce       []byte            `json:"nonce"`
}
type AttesterResponse struct {
	Result      string `json:"result"`
	Attestation []byte `json:"report"`
}

func PerformRelyingPartyCLI(app *App, endpoint string, computation types.Computation, computationInput []string) error {
	claim, proof, err := AttestAndGenClaim(app, endpoint, computation, computationInput)
	if err != nil {
		return fmt.Errorf("failed generating claim: %v", err)
	}
	_, _, err = VerifyClaim(app, claim, proof)
	if err != nil {
		return fmt.Errorf("claim not validated: %v", err)
	}
	printSuccess(claim)
	return nil
}

func AttestAndGenClaim(app *App, endpoint string, computation types.Computation, computationInput []string) (*types.TEEComputationClaim, *types.TEEComputationClaimVerificationContext, error) {
	nonce := generateNonce(app)
	resp, err := queryAttester(endpoint, computation, computationInput, nonce)
	if err != nil {
		return nil, nil, fmt.Errorf("error from attester: %v", err)
	}
	claim, proof, err := generation.GenerateTEEComputationClaim(computation, computationInput, resp.Result, resp.Attestation, nonce)
	if err != nil {
		return nil, nil, fmt.Errorf("failed generating claim: %v", err)
	}
	return claim, proof, nil
}

// Returns the attester's reponse.
func queryAttester(attesterURL string, computation types.Computation, input []string, nonce []byte) (AttesterResponse, error) {
	log.Println("Querying attester server...")
	var result AttesterResponse

	query := AttesterQuery{
		ClaimType:   "TEEComputation",
		Computation: computation,
		Input:       input,
		Nonce:       nonce,
	}
	reqBody, err := json.Marshal(query)
	if err != nil {
		return result, fmt.Errorf("error marshaling query to verifier: %v", err)
	}
	resp, err := http.Post(attesterURL, "application/json", bytes.NewBuffer(reqBody))

	if err != nil {
		return result, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		if body, err := io.ReadAll(resp.Body); err == nil {
			return result, fmt.Errorf("attester error: %v", string(body))
		}
		return result, fmt.Errorf("unknown attester error")
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("error unmarshling attester response: %v", err)
	}

	return result, nil
}

// First returned value is the claim ID
func VerifyClaim(app *App, claim *types.TEEComputationClaim, proof *types.TEEComputationClaimVerificationContext) (string, types.SignedSettledVerifiedClaim, error) {
	var vClaim types.SignedSettledVerifiedClaim
	// 1. Submit claim to VSL
	claimId, submitTime, err := SubmitClaim(app, claim, proof)
	if err != nil {
		return "", vClaim, err
	}
	// 2. Periodically query VSL until expiry time is over, to obtain validated claim
	log.Println("Polling VSL for validated claims...")
	vClaim, err = app.VSLClient.PollSettledByID(claimId, submitTime, app.ExpirySeconds, app.LoopInterval)
	if err != nil {
		return claimId, vClaim, fmt.Errorf("VSL error: %w", err)
	}
	return claimId, vClaim, nil
}

// Returns claim ID and submission time
func SubmitClaim(app *App, claim *types.TEEComputationClaim, proof *types.TEEComputationClaimVerificationContext) (string, time.Time, error) {
	log.Println("Submitting claim to VSL...")
	submitTime := time.Now()

	jsonClaim, err := json.Marshal(claim)
	if err != nil {
		return "", time.Now(), fmt.Errorf("error marshaling query to verifier: %v", err)
	}
	jsonProof, err := json.Marshal(proof)
	if err != nil {
		return "", time.Now(), fmt.Errorf("failed marshaling proof: %w", err)
	}

	verifiers := make([]string, 1)
	verifiers[0] = app.VerifierAddress
	claimId, err := app.VSLClient.SubmitClaim(
		app.ClientPrivateKey,
		string(jsonClaim),
		string(claim.ClaimType),
		string(jsonProof),
		verifiers,
		app.ClientAddress,
		app.ExpirySeconds,
		app.Fee,
	)
	if err != nil {
		return "", time.Now(), fmt.Errorf("VSL error: %w", err)
	}
	log.Printf("Claim ID: %s", claimId)
	return claimId, submitTime, nil
}

func generateNonce(app *App) []byte {
	nonce := make([]byte, 8)
	if !app.ZeroNonce {
		rand.Read(nonce)
	}
	return nonce
}

func printSuccess(claim *types.TEEComputationClaim) {
	// Open (or create) the file for appending
	f, err := os.OpenFile("success.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		return
	}
	defer f.Close()

	// Create a logger that writes to the file
	logger := log.New(f, "", log.LstdFlags)

	logger.Println("Successfully validated TEE claim:")
	logger.Println("\tComputation: " + claim.Computation)
	logger.Printf("\tInput: ")
	for _, arg := range claim.Input {
		logger.Println("\t\t" + arg)
	}
	logger.Println("\tResult: " + claim.Result)
	log.Println("Validated TEE claim successfully. Details logged to success.log.")
}

// submitClaimToBackend sends a claim record to the backend and returns true if successful.
func submitClaimToBackend(backendURL string, blockNumber uint64, claimID string, client string) (bool, error) {
	payload := map[string]interface{}{
		"block_number":     blockNumber,
		"execution_client": client,
		"claim_id":         claimID,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/block_mirroring_record", backendURL)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return false, fmt.Errorf("failed to send POST request: %w", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300, nil
}
