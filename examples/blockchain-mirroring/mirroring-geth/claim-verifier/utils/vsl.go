package utils

import (
	"errors"
	"fmt"
	"log"
	"mirroring-geth-claim-verifier/models"

	"base/pkg/vsl"

	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/v3/client"
)

const (
	ExecutionClient = "MirroringGeth"
)

func SettleClaimToVSL(app *models.App, claimId string) (*string, error) {
	nonce, err := app.VSLRPCClient.GetAccountNonce(vsl.GetAccountNonceParams{
		AccountId: app.VerifierAddress,
	})
	if err != nil {
		errString := fmt.Sprintf("Error getting account nonce: %+v", err)
		log.Println(errString)
		return nil, errors.New(errString)
	}
	settledClaimId, err := app.VSLRPCClient.SettleClaim(vsl.SettleClaimParams{
		From:          app.VerifierAddress,
		Nonce:         fmt.Sprintf("%d", *nonce),
		TargetClaimId: claimId,
	})
	if err != nil {
		errString := fmt.Sprintf("Error submitting claim to VSL: %+v", err)
		log.Println(errString)
		return nil, errors.New(errString)
	}

	return settledClaimId, nil
}

func SubmitClaimToBackend(app *models.App, claimId *string, verificationTime *uint64, errString *string) error {
	requestBody := fiber.Map{
		"claim_id":          claimId,
		"execution_client":  ExecutionClient,
		"verification_time": verificationTime,
		"error":             errString,
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
