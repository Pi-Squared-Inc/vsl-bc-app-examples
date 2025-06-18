package models

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"base/pkg/abstract_types"

	"github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/crypto/sha3"
)

/**
 * EVMBlockProcessing claim and its verification context
 */

// EVMBlockProcessingClaim is a type alias for BaseClaim
type EVMBlockProcessingClaim struct {
	ClaimType   string                     `json:"type"`
	Assumptions *types.Header              `json:"assumptions"`
	Result      []byte                     `json:"result"`
	Metadata    abstract_types.EVMMetadata `json:"metadata"`
}

// TODO: Properly implement this
func (c *EVMBlockProcessingClaim) GetId() (*string, error) {
	jsonBytes, err := json.Marshal(c)
	if err != nil {
		fmt.Printf("Error encoding to JSON: %v\n", err)
		return nil, errors.New("error encoding to JSON")
	}

	hash := sha3.NewLegacyKeccak256()
	hash.Write(jsonBytes)
	hashString := hex.EncodeToString(hash.Sum(nil))
	return &hashString, nil
}

// VerificationContext for EVMBlockProcessingClaim
type EVMBlockProcessingClaimVerificationContext struct {
	// Witness of the block
	Witness []byte `json:"witness"`
}
