package utils

import (
	"base/pkg/abstract_types"
	"base/pkg/ethrpc"
	"base/pkg/evm"
	"base/pkg/vsl"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"observer/models"
	"strings"
	"time"

	generationModels "generation-view-fn-evm/pkg/models"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/v3/client"
	"github.com/pkg/errors"
)

type SubmitClaimParams struct {
	Claim         string                   `json:"claim"`
	ClaimType     string                   `json:"claim_type"`
	Proof         string                   `json:"proof"`
	Nonce         string                   `json:"nonce"`
	Verifiers     []string                 `json:"verifiers"`
	MinSignatures uint16                   `json:"min_signatures"`
	Client        string                   `json:"client"`
	Expires       abstract_types.Timestamp `json:"expires"`
	Fee           uint64                   `json:"fee"`
}

type SignedSubmitClaimParams struct {
	SubmitClaimParams
	evm.SignedComponents
}

// SubmitClaimToPod will sends the claim to the USL API
func SubmitClaimToVSL(app *models.App, claim interface{}, verificationContext interface{}) (*string, *string, *string, error) {
	log.Printf("Submitting claim to VSL to url %s", app.VSLRPC)

	claimBytes, err := claim.(*generationModels.EVMViewFnClaim).AbiEncode()
	if err != nil {
		return nil, nil, nil, errors.WithStack(err)
	}

	proofBytes, err := verificationContext.(*generationModels.EVMViewFnClaimVerificationContext).AbiEncode()
	if err != nil {
		return nil, nil, nil, errors.WithStack(err)
	}

	clientNonce, err := app.VSLRPCClient.GetAccountNonce(vsl.GetAccountNonceParams{
		AccountId: app.VSLClientAddress,
	})
	if err != nil {
		return nil, nil, nil, errors.WithStack(err)
	}

	claimHex := hexutil.Encode(claimBytes)
	proofHex := hexutil.Encode(proofBytes)

	claimId, err := app.VSLRPCClient.SubmitClaim(vsl.SubmitClaimParams{
		Claim:     claimHex,
		ClaimType: "EVMViewFn",
		Proof:     proofHex,
		To:        []string{app.VSLVerifierAddress},
		Quorum:    1,
		From:      app.VSLClientAddress,
		Nonce:     fmt.Sprintf("%d", *clientNonce),
		Expires: abstract_types.Timestamp{
			Seconds: uint64(time.Now().Unix() + 600), // 10 minutes
			Nanos:   0,
		},
		Fee: hexutil.EncodeBig(big.NewInt(1)),
	})

	if err != nil {
		log.Printf("Failed to submit claim: %s", err)
		return nil, nil, nil, errors.WithStack(err)
	}

	log.Printf("Claim submitted to VSL with id %s", *claimId)

	return claimId, &claimHex, &proofHex, nil
}

func SubmitClaimToBackend(app *models.App, sourceChainTransactionHex string, claimId string, claim any, claimHex string) error {
	srcChainRPCClient, err := rpc.Dial(app.SourceChainRPCEndpoint)
	if err != nil {
		return errors.WithStack(err)
	}
	srcChainTx, err := ethrpc.GetTransactionByHash(srcChainRPCClient, context.Background(), common.HexToHash(sourceChainTransactionHex).String())
	if err != nil {
		return errors.WithStack(err)
	}

	// Submit claim to backend
	apiClient := client.New()

	claimJSON, err := json.Marshal(claim)
	if err != nil {
		return errors.WithStack(err)
	}
	reqBody := fiber.Map{
		"claim_id":                claimId,
		"address":                 strings.ToLower(srcChainTx.From),
		"source_transaction_hash": sourceChainTransactionHex,
		"claim_json":              string(claimJSON),
		"claim":                   claimHex,
	}
	resp, err := apiClient.Put(app.BackendAPIEndpoint+"/claim-by-source-tx/"+sourceChainTransactionHex, client.Config{
		Body: reqBody,
	})
	if err != nil {
		return errors.WithStack(err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("failed to submit claim\nError: %s", resp.Body())
	}

	return nil
}
