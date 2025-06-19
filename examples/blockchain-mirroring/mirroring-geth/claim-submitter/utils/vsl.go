package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	generationModels "generation-block-processing-evm/pkg/models"
	"log"
	"math/big"
	"mirroring-geth-claim-submitter/models"
	"time"

	"base/pkg/abstract_types"
	"base/pkg/vsl"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/v3/client"
)

func SubmitClaimToVSL(app *models.App, blockNumber uint64, claim *generationModels.EVMBlockProcessingClaim, verificationContext *generationModels.EVMBlockProcessingClaimVerificationContext, errString *string) (*string, error) {
	claimJSON, err := json.Marshal(claim)
	if err != nil {
		errString := fmt.Sprintf("Error marshalling claim: %+v", err)
		log.Println(errString)
		return nil, errors.New(errString)
	}
	verificationContextJSON, err := json.Marshal(verificationContext)
	if err != nil {
		errString := fmt.Sprintf("Error marshalling verification context: %+v", err)
		log.Println(errString)
		return nil, errors.New(errString)
	}

	nonce, err := app.VSLClient.GetAccountNonce(vsl.GetAccountNonceParams{
		AccountId: app.VSLSubmitterAddress,
	})
	if err != nil {
		errString := fmt.Sprintf("Error getting account nonce: %+v", err)
		log.Println(errString)
		return nil, errors.New(errString)
	}
	claimId, err := app.VSLClient.SubmitClaim(vsl.SubmitClaimParams{
		Claim:     string(claimJSON),
		ClaimType: "MirroringGeth",
		Proof:     string(verificationContextJSON),
		Nonce:     fmt.Sprintf("%d", *nonce),
		To:        []string{app.VSLVerifierAddress},
		Quorum:    1,
		From:      app.VSLSubmitterAddress,
		Expires: abstract_types.Timestamp{
			Seconds: uint64(time.Now().Unix() + 600), // 10 minutes,
			Nanos:   0,
		},
		Fee: hexutil.EncodeBig(big.NewInt(1)),
	})
	if err != nil {
		errString := fmt.Sprintf("Error submitting claim to VSL: %+v", err)
		log.Println(errString)
		return nil, errors.New(errString)
	}

	return claimId, nil
}

const (
	ExecutionClient = "MirroringGeth"
)

func SubmitClaimToBackend(app *models.App, blockNumber uint64, claimId *string, errString *string) error {
	requestBody := fiber.Map{
		"block_number":     blockNumber,
		"execution_client": ExecutionClient,
		"claim_id":         claimId,
		"error":            errString,
	}

	remoteClient := client.New()
	resp, err := remoteClient.Post(app.BackendEndpoint+"/block_mirroring_record", client.Config{
		Body: requestBody,
	})
	if err != nil {
		errorString := fmt.Sprintf("Error sending claim to remote RPC: %+v", err)
		log.Println(errorString)
		return errors.New(errorString)
	}
	if resp.StatusCode() != 200 {
		errorString := fmt.Sprintf("Error sending claim to remote RPC: %s", string(resp.Body()))
		log.Println(errorString)
		return errors.New(errorString)
	}

	return nil
}
