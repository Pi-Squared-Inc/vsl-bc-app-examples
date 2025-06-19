package cmd

import (
	types "base-tee/pkg/abstract_types"
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"vsl-rpc-demo/cmd"

	verification "verification/pkg/verification"

	vsl "vsl-rpc-demo/vsl-wrapper"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	VSL_HOST      string
	VSL_PORT      string
	VERIFIER_ADDR string
	VERIFIER_PRIV *ecdsa.PrivateKey
	LOOP_INTERVAL time.Duration
	NUM_LOOPS     int
)

// verifierCmd represents the verifier command
var verifierCmd = &cobra.Command{
	Use:   "verifier",
	Short: "Polls the VSL for new TEE verification requests addressed to it, and attempts to settle them.",
	Run: func(cmd *cobra.Command, args []string) {
		err := PerformVerifier()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	verifierCmd.Flags().IntVar(&NUM_LOOPS, "num-claims", -1, "Maximum number of claims to verify. If -1 (default), will poll forever. Useful for testing.")
	cmd.RootCmd.AddCommand(verifierCmd)
}

func PerformVerifier() error {
	get_env()

	vsl, err := vsl.DialVSL(VSL_HOST, VSL_PORT)
	if err != nil {
		return fmt.Errorf("failed VSL connection: %v", err)
	}

	var teeClaim types.TEEComputationClaim
	var teeProof types.TEEComputationClaimVerificationContext
	since := types.Timestamp{
		Seconds:     uint64(time.Now().Unix()),
		Nanoseconds: 0,
	}
	claims_left := NUM_LOOPS
	log.Println("Verifier: Polling...")
	for claims_left != 0 {
		newClaims, err := vsl.ListSubmittedByVerifier(
			VERIFIER_ADDR,
			since,
		)
		if err != nil {
			log.Printf("VSL error: %v\n", err)
			continue
		}
		for _, vslClaim := range newClaims {
			since = types.MaxT(since, vslClaim.Timestamp.Tick())
			err = json.Unmarshal([]byte(vslClaim.Data.Claim), &teeClaim)
			if err != nil {
				return fmt.Errorf("could not gather claim")
			}
			err = json.Unmarshal([]byte(vslClaim.Data.Proof), &teeProof)
			if err != nil {
				return fmt.Errorf("could not gather proof")
			}
			startTime := time.Now() // start timer for verification
			err = verification.VerifyTEEComputationClaim(&teeClaim, &teeProof)
			elapsedTime := time.Since(startTime) // measure verification time
			if err != nil {
				log.Printf("Claim not verified! %v", err)
				continue
			}
			// Settle on VSL:
			settledId, err := vsl.Settle(
				VERIFIER_PRIV,
				VERIFIER_ADDR,
				vslClaim.ID,
			)
			if err != nil {
				log.Printf("VSL error: %v\n", err)
				continue
			}
			log.Println("Settled claim ID: ", settledId)

			if teeClaim.Computation == types.BlockProcessingKReth {
				// Set verification time for claim on backend:
				verificationTime := uint64(elapsedTime.Microseconds())
				backendURL := os.Getenv("BLOCK_MIRRORING_BACKEND_URL")
				if backendURL == "" {
					return fmt.Errorf("BLOCK_MIRRORING_BACKEND_URL not set in environment variables")
				}
				success, err := setVerificationTimeForClaim(
					backendURL,
					settledId,
					"MirroringKRethTEE",
					verificationTime,
				)
				if err != nil {
					log.Printf("Failed to set verification time for claim %s: %v", settledId, err)
					continue
				}
				if !success {
					log.Printf("Failed to set verification time for claim %s on backend", settledId)
					continue
				}
				log.Printf("Verification time for claim %s set successfully on backend", settledId)
			}

			log.Printf("Claim verified in %.2f milliseconds", elapsedTime.Seconds()*1000)
			claims_left = min(claims_left-1, 0)
		}
		time.Sleep(LOOP_INTERVAL * time.Second)
	}
	return nil
}

func get_env() {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) { // Ignore if .env file does not exist
		log.Fatalf("Error loading environment variables")
	}
	VSL_HOST = os.Getenv("VSL_HOST")
	VSL_PORT = os.Getenv("VSL_PORT")
	VERIFIER_ADDR = os.Getenv("VERIFIER_ADDR")
	privStr := os.Getenv("VERIFIER_PRIV")
	seconds, err := strconv.Atoi(os.Getenv("VERIFIER_LOOP_INTERVAL"))

	if VSL_HOST == "" || VSL_PORT == "" || VERIFIER_ADDR == "" || err != nil {
		log.Fatalf("Error loading environment variables")
	}
	LOOP_INTERVAL = time.Duration(seconds)
	privBytes, err := hexutil.Decode(privStr)
	if err != nil {
		log.Fatalf("Error loading private key")
	}
	VERIFIER_PRIV, err = crypto.ToECDSA(privBytes)
	if err != nil {
		log.Fatalf("Error loading private key")
	}
}

// setVerificationTimeForClaim sends verification time for a claim to the backend and returns true if successful.
func setVerificationTimeForClaim(backendURL, claimID, client string, verificationTime uint64) (bool, error) {
	payload := map[string]interface{}{
		"execution_client":  client,
		"claim_id":          claimID,
		"verification_time": verificationTime,
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
