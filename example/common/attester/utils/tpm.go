package utils

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/google/go-tpm/tpmutil"
	"google.golang.org/protobuf/proto"
)

const (
	PCRIndex = 23
)

// Resets the contents of the PCR
// Note: Only PCRs 23 and 16 can be reset with default userland permissions.
func ResetPCR(tpm io.ReadWriteCloser) error {
	pcr := tpmutil.Handle(PCRIndex)
	err := tpm2.PCRReset(tpm, pcr)
	if err != nil {
		return fmt.Errorf("failed resetting PCR %v: %v", PCRIndex, err)
	}
	return nil
}

// Extends the PCR with the digests of the byte array given as argument
func ExtendPCR(tpm io.ReadWriteCloser, extendData []byte) error {
	pcr := tpmutil.Handle(PCRIndex)

	sha384 := sha512.Sum384(extendData)
	sha256 := sha256.Sum256(extendData)
	sha1 := sha1.Sum(extendData)

	err := tpm2.PCRExtend(tpm, pcr, tpm2.AlgSHA384, sha384[:], "")
	if err != nil {
		return fmt.Errorf("failed extending PCR %v: %v", PCRIndex, err)
	}
	err = tpm2.PCRExtend(tpm, pcr, tpm2.AlgSHA256, sha256[:], "")
	if err != nil {
		return fmt.Errorf("failed extending PCR %v: %v", PCRIndex, err)
	}
	err = tpm2.PCRExtend(tpm, pcr, tpm2.AlgSHA1, sha1[:], "")
	if err != nil {
		return fmt.Errorf("failed extending PCR %v: %v", PCRIndex, err)
	}

	log.Printf("Extended SHA384 bank with digest: %x\n", sha384)
	log.Printf("Extended SHA256 bank with digest: %x\n", sha256)
	log.Printf("Extended SHA1 bank with digest: %x\n", sha1)

	return nil
}

// Extends the PCR with the digests of the file whose path is given as argument
func ExtendPCRFileHash(tpm io.ReadWriteCloser, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	hasherSHA1 := sha1.New()
	hasherSHA256 := sha256.New()
	hasherSHA384 := sha512.New384()

	multiWriter := io.MultiWriter(hasherSHA1, hasherSHA256, hasherSHA384)

	if _, err := io.Copy(multiWriter, file); err != nil {
		return fmt.Errorf("failed to copy file content to hashers: %w", err)
	}

	// Get the hash sums
	sha1 := hasherSHA1.Sum(nil)
	sha256 := hasherSHA256.Sum(nil)
	sha384 := hasherSHA384.Sum(nil)

	// Extend the PCR
	pcr := tpmutil.Handle(PCRIndex)
	err = tpm2.PCRExtend(tpm, pcr, tpm2.AlgSHA384, sha384[:], "")
	if err != nil {
		return fmt.Errorf("failed extending PCR %v: %v", PCRIndex, err)
	}
	err = tpm2.PCRExtend(tpm, pcr, tpm2.AlgSHA256, sha256[:], "")
	if err != nil {
		return fmt.Errorf("failed extending PCR %v: %v", PCRIndex, err)
	}
	err = tpm2.PCRExtend(tpm, pcr, tpm2.AlgSHA1, sha1[:], "")
	if err != nil {
		return fmt.Errorf("failed extending PCR %v: %v", PCRIndex, err)
	}

	log.Printf("Extended SHA384 bank with digest: %x\n", sha384)
	log.Printf("Extended SHA256 bank with digest: %x\n", sha256)
	log.Printf("Extended SHA1 bank with digest: %x\n", sha1)
	return nil
}

// Fetches an attestation report for the TPM and TEE devices
func FetchAttestationReport(tpm io.ReadWriteCloser, nonce []byte) (string, error) {
	// Attestation report generation: experimental, this is just a WIP compiled from various code examples
	ak, err := client.AttestationKeyECC(tpm)
	if err != nil {
		return "", fmt.Errorf("failed to create attestation key: %v", err)
	}
	defer ak.Close()

	sevqp, err := client.CreateSevSnpQuoteProvider()
	if err != nil {
		return "", fmt.Errorf("expected SEV TEE device... %v", err)
	}

	eventLog, err := client.GetEventLog(tpm)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve TCG Event Log: %v", err)
	}

	attestation, err := ak.Attest(client.AttestOpts{
		Nonce:       nonce,
		TEEDevice:   sevqp,
		TCGEventLog: eventLog,
	})
	if err != nil {
		return "", fmt.Errorf("failed to attest: %v", err)
	}

	tpm_out, err := proto.Marshal(attestation)
	if err != nil {
		return "", fmt.Errorf("failed to marshal attestation proto: %v", attestation)
	}
	b64Report := base64.StdEncoding.EncodeToString(tpm_out)
	return b64Report, nil
}

func GetAKB64(tpm io.ReadWriteCloser) (string, error) {
	ak, err := client.AttestationKeyECC(tpm)
	if err != nil {
		return "", fmt.Errorf("failed to create attestation key: %v", err)
	}
	defer ak.Close()
	derBytes, err := x509.MarshalPKIXPublicKey(ak.PublicKey())
	if err != nil {
		log.Fatalf("failed to marshal public key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(derBytes), nil
}

// Generates the SHA384, SHA256 and SHA1 digests of the file whose path is given as argument
func GenerateSHAHashes(path string) error {
	// Read the file data
	fileData, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed reading file: %v", err)
	}

	// Generate the SHA384, SHA256 and SHA1 hashes
	sha384 := sha512.Sum384(fileData)
	sha256 := sha256.Sum256(fileData)
	sha1 := sha1.Sum(fileData)

	// Print the hashes
	log.Printf("SHA384: %x\n", sha384)
	log.Printf("SHA256: %x\n", sha256)
	log.Printf("SHA1: %x\n", sha1)

	return nil
}
