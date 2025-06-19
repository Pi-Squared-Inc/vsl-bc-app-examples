package verification

import (
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// getTrustedAKs retrieves the trusted attestation keys from the environment variable GCP_AK_B64_LIST.
func getTrustedAKs() ([]crypto.PublicKey, error) {
	// Load .env file (only needed once, can be moved to init or main)
	_ = godotenv.Load()

	akListStr := os.Getenv("GCP_AK_B64_LIST")
	if akListStr == "" {
		return nil, fmt.Errorf("GCP_AK_B64_LIST not set in environment")
	}
	akList := strings.Split(akListStr, ",")

	trustedAKs := make([]crypto.PublicKey, 0, len(akList))
	for _, b64Key := range akList {
		b64Key = strings.TrimSpace(b64Key)
		if b64Key == "" {
			continue
		}
		pk, err := makePK(b64Key)
		if err != nil {
			return nil, err
		}
		trustedAKs = append(trustedAKs, pk)
	}
	return trustedAKs, nil
}

func makePK(b64Key string) (crypto.PublicKey, error) {
	derBytes, err := base64.StdEncoding.DecodeString(b64Key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode b64 public key: %w", err)
	}
	pub, err := x509.ParsePKIXPublicKey(derBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}
	return pub, nil
}
