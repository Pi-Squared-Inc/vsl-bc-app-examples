package client

import (
	types "base-tee/pkg/abstract_types"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	_, currentFile, _, _  = runtime.Caller(0)
	src_dir               = filepath.Dir(currentFile) + "/src/"
	program               = src_dir + "mirroring-reth-kevm"
	blocks_dir            = "blocks/"
	stop_at_first_failure = false
)

var blockProcessingKRethCmd = &cobra.Command{
	Use:   "block_processing_kreth",
	Short: "Fetches Ethereum blocks using Reth, generates and requests verification of claim-context pair on a TEE.",
	Run: func(cmd *cobra.Command, args []string) {
		rpc_host, rpc_port, verifier_addr, client_addr, client_priv, exp_seconds, loop_interval := get_env()
		bigIntFee := new(big.Int).SetUint64(Fee)
		APP = NewApp(
			rpc_host,
			rpc_port,
			verifier_addr,
			client_addr,
			client_priv,
			exp_seconds,
			loop_interval,
			bigIntFee,
			Zero_Nonce,
		)
		ctx, cancel := context.WithCancel(context.Background())
		go catchSigInt(ctx, cancel)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			runBlockProcessingReth(ctx)
		}()
		blockProcessingRequests(ctx, cancel)
		wg.Wait()
	},
}

func catchSigInt(ctx context.Context, cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	select {
	case <-c:
		log.Println("Received CTRL-C, will gracefully exit...")
		cancel()
	case <-ctx.Done():
	}
	signal.Stop(c)
	cancel()
}

// runBlockProcessingReth fetches Ethereum blocks using Reth, generates and saves the claim-context pair
// in the blocks directory
func runBlockProcessingReth(ctx context.Context) {

	// Path of block_processing_reth binary executable
	if _, err := os.Stat(program); os.IsNotExist(err) {
		log.Printf("cannot find block_processing_reth executable: %v", program)
		return
	}
	log.Printf("Running block processing using Reth...")

	cmd := exec.Command(program)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start process: %v", err)
		return
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-ctx.Done():
		log.Println("Context canceled, killing process group...")
		if cmd.Process != nil {
			// Kill the process group (-pid)
			if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
				log.Printf("Failed to kill process group: %v", err)
			} else {
				log.Println("Process group killed.")
			}
		}
		<-done // Wait for process to exit
		return
	case err := <-done:
		if err != nil {
			log.Printf("Process exited with error: %v", err)
		} else {
			log.Printf("Process exited successfully.")
		}
		return
	}
}

// blockProcessingRequests processes the JSON files in the blocks directory and sends the claim-context pairs
// to the PerformRelyingParty function for verification
func blockProcessingRequests(ctx context.Context, cancel context.CancelFunc) {
	for {
		// if context is canceled, exit the loop
		select {
		case <-ctx.Done():
			log.Println("Context canceled, stopping blockProcessingRequests.")
			return
		default:
		}

		log.Printf("Checking for JSON files in %s...", blocks_dir)
		files, err := os.ReadDir(blocks_dir)
		if err != nil {
			log.Printf("Failed to read blocks directory: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		found := false
		for _, file := range files {
			// Check if context is canceled before processing each file
			select {
			case <-ctx.Done():
				log.Println("Context canceled, stopping blockProcessingRequests.")
				return
			default:
			}
			if filepath.Ext(file.Name()) == ".json" {
				found = true
				filePath := filepath.Join(blocks_dir, file.Name())
				log.Printf("Found JSON file: %s", filePath)
				data, err := os.ReadFile(filePath)
				if err != nil {
					log.Printf("Failed to read file %s: %v", filePath, err)
					continue
				}

				var parsed map[string]interface{}
				if err := json.Unmarshal(data, &parsed); err != nil {
					log.Printf("Failed to parse JSON in %s: %v", filePath, err)
					continue
				}

				byteClaim, _ := json.Marshal(parsed["claim"])
				contextData, _ := json.Marshal(parsed["context"])

				log.Printf("Processing and requesting verification for claim and context from %s", filePath)
				computationInput := make([]string, 2)
				computationInput[0] = string(byteClaim)
				computationInput[1] = string(contextData)
				claim, proof, err := AttestAndGenClaim(APP, ATTESTER_ENDPOINT, "block_processing_kreth", computationInput)
				if err != nil {
					log.Printf("Claim generation failed: %v", err)
					if stop_at_first_failure {
						log.Println("Stopping block processing due to failure.")
						cancel() // Cancel context to stop runBlockProcessingReth
						return
					}
				}
				if err = submitToBackend(APP, claim, proof); err != nil {
					log.Printf("Submitting block to backend failed: %v", err)
					if stop_at_first_failure {
						log.Println("Stopping block processing due to failure.")
						cancel() // Cancel context to stop runBlockProcessingReth
						return
					}
				}

				// Remove the file after processing
				if err := os.Remove(filePath); err != nil {
					log.Printf("Failed to remove processed file %s: %v", filePath, err)
				} else {
					log.Printf("Processed and removed file: %s", filePath)
				}
			}
		}

		if !found {
			log.Println("No JSON files found, waiting for next check...")
			time.Sleep(5 * time.Second)
		}
	}
}

func submitToBackend(app *App, claim *types.TEEComputationClaim, proof *types.TEEComputationClaimVerificationContext) error {
	// First: Submit claim to VSL
	claimId, _, err := SubmitClaim(app, claim, proof)
	if err != nil {
		return err
	}

	// Next: submit claim to block mirroring backend
	// Retrieve the block number
	log.Println("Submitting claim to backend for block mirroring...")
	var inputMap map[string]interface{}
	err = json.Unmarshal([]byte(claim.Input[0]), &inputMap)
	if err != nil {
		return fmt.Errorf("failed to unmarshal input: %w", err)
	}
	assumptions, ok := inputMap["assumptions"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("assumptions not found or invalid in input: %v", claim.Input[0])
	}
	blockNumberHex, ok := assumptions["number"].(string)
	if !ok {
		return fmt.Errorf("block number not found in assumptions: %v", claim.Input[0])
	}
	blockNumber, err := strconv.ParseUint(blockNumberHex, 0, 64)
	if err != nil {
		return fmt.Errorf("failed to parse block number from assumptions: %w", err)
	}
	// Get block mirroring backend URL from .env file
	backendURL := os.Getenv("BLOCK_MIRRORING_BACKEND_URL")
	if backendURL == "" {
		return fmt.Errorf("BLOCK_MIRRORING_BACKEND_URL not set in environment variables")
	}
	// Submit the claim to the backend for block mirroring
	success, err := submitClaimToBackend(backendURL, blockNumber, claimId, "MirroringKRethTEE")
	if err != nil {
		return fmt.Errorf("failed to submit claim to backend: %w", err)
	}
	if !success {
		return fmt.Errorf("failed to submit claim to backend for block mirroring")
	}
	log.Printf("Claim successfully submitted to backend for block mirroring for block number %d", blockNumber)
	return nil
}

func init() {
	blockProcessingKRethCmd.Flags().BoolVar(&Zero_Nonce, "zero-nonce", false, "Sets nonce to zero when requesting attestation report. Useful for testing.")

	blockProcessingKRethCmd.Flags().Uint64Var(&Fee, "fee", 1*(1e18), "Fee promised for claim verification (in atto-VSL).")
	blockProcessingKRethCmd.Flags().BoolVar(&stop_at_first_failure, "fail-stop", false, "Whether to stop the client on the first error encountered, or keep polling for new blocks. By default, false (will not stop).")
	clientCmd.AddCommand(blockProcessingKRethCmd)
}
