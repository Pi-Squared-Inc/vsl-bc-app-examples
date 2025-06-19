package block_processing

import (
	utils "attester/utils"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"runtime"
	"path/filepath"
)

var (
	_, currentFile, _, _ = runtime.Caller(0)
	kreth_dir = filepath.Dir(currentFile) + "/kreth/"
	program = kreth_dir + "block_processing_kreth"
)

func BlockProcessingKReth(tpm io.ReadWriteCloser, claim string, context string) (string, error) {

	// Path of block_processing_kreth binary executable
	if _, err := os.Stat(program); os.IsNotExist(err) {
		return "", fmt.Errorf("cannot find block_processing_kreth executable: %v", program)
	}

	log.Println("Block processing using KReth: Saving claim and context to temporary files...")
	// Create temp files for claim
	tmp_claim_file, err := os.CreateTemp(".", "json*")
	if err != nil {
		return "", fmt.Errorf("cannot create temp file for claim: %v", err)
	}
	defer os.Remove(tmp_claim_file.Name())
	bytesForClaim, err := tmp_claim_file.WriteString(claim)
	if err != nil || len(claim) != bytesForClaim {
		return "", fmt.Errorf("couldn't write claim to temp file: %v", err)
	}
	tmp_claim_path := tmp_claim_file.Name()
	// Create temp files for context
	tmp_context_file, err := os.CreateTemp(".", "json*")
	if err != nil {
		return "", fmt.Errorf("cannot create temp file for context: %v", err)
	}
	defer os.Remove(tmp_context_file.Name())
	bytesForContext, err := tmp_context_file.WriteString(context)
	if err != nil || len(context) != bytesForContext {
		return "", fmt.Errorf("couldn't write context to temp file: %v", err)
	}
	tmp_context_path := tmp_context_file.Name()

	// Invoke the block processing (using kreth) binary executable 
	log.Println("Block processing using KReth: Executing program...")
	out, err := exec.Command(program, tmp_claim_path, tmp_context_path).Output()
	if err != nil || !strings.Contains(string(out), "verification_time") {
		return "", fmt.Errorf("couldn't run the block processing (using kreth) binary executable successfully: out: '%v', err: %v", string(out), err)
	}
	log.Printf("Block processing using KReth: %s", string(out))

	// Log all this interaction to PCR 23:
	log.Println("Block processing using KReth: Resetting PCR 23 old value...")
	utils.ResetPCR(tpm)
	log.Println("Block processing using KReth: Extending PCR 23 with current measurements...")

	// Hash of the block processing (using kreth) binary executable 
	log.Println("Logging executed program...")
	utils.ExtendPCRFileHash(tpm, program)
	// Hash of the claim
	log.Println("Logging claim...")
	utils.ExtendPCRFileHash(tpm, tmp_claim_path)
	// Hash of the output
	log.Println("Logging verification time...")
	utils.ExtendPCR(tpm, out)

	log.Println("Block processing using KReth: Done.")

	return string(out), nil
}