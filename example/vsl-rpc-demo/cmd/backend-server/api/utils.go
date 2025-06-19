package api

import (
	types "base-tee/pkg/abstract_types"
	"encoding/json"
	"log"
	"time"
	verification "verification/pkg/verification"
	"vsl-rpc-demo/cmd/backend-server/models"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gofiber/fiber/v3"
)

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func validate(c fiber.Ctx, app *models.App, address *string, claim *types.TEEComputationClaim, verificationContext *types.TEEComputationClaimVerificationContext) error {
	claimBytes, err := json.Marshal(claim)
	if err != nil {
		log.Println("error marshalling claim: ", err)
		return c.Status(500).SendString("internal server error")
	}
	claimString := string(claimBytes)

	// Verification context
	verificationContextBytes, err := json.Marshal(verificationContext)
	if err != nil {
		log.Println("error marshalling verification context: ", err)
		return c.Status(500).SendString("internal server error")
	}
	verificationContextString := string(verificationContextBytes)

	computationInputJSON, err := json.Marshal(claim.Input)
	if err != nil {
		log.Println("error marshalling computation input: ", err)
		return c.Status(500).SendString("internal server error")
	}
	computationInputString := string(computationInputJSON)

	attestationString := hexutil.Encode(verificationContext.Attestation)
	nonceString := hexutil.Encode(claim.Nonce)

	// Validate the claim and get the validation time.
	var validationTime uint64
	start := time.Now()
	err = verification.VerifyTEEComputationClaim(claim, verificationContext)
	duration := time.Since(start)
	validationTime = uint64(duration.Microseconds())
	computation := string(claim.Computation)
	if err != nil {
		errMsg := err.Error()
		validationRecord, err := storeValidationRecord(app, *address, &errMsg, &claimString, &verificationContextString, nil, &computation, &computationInputString, &nonceString, &claim.Result, &attestationString)
		if err != nil {
			log.Println("error storing validation record: ", err)
			return c.Status(500).SendString("internal server error")
		}
		return c.JSON(validationRecord.ToResponse())
	}

	validationRecord, err := storeValidationRecord(app, *address, nil, &claimString, &verificationContextString, &validationTime, &computation, &computationInputString, &nonceString, &claim.Result, &attestationString)
	if err != nil {
		log.Println("error storing validation record: ", err)
		return c.Status(500).SendString("internal server error")
	}
	return c.JSON(validationRecord.ToResponse())
}

// storeValidationRecord creates and stores a validation record with the given parameters
// Returns the created record and any error that occurred during storage
func storeValidationRecord(
	app *models.App,
	address string,
	errorMessage *string,
	claim *string,
	verificationContext *string,
	validationTime *uint64,
	computation *string,
	input *string,
	nonce *string,
	result *string,
	attestation *string,
) (*models.ValidationRecord, error) {
	record := models.ValidationRecord{
		Address:             address,
		ValidationError:     errorMessage,
		Claim:               claim,
		VerificationContext: verificationContext,
		ValidationTime:      validationTime,
		Computation:         computation,
		Input:               input,
		Nonce:               nonce,
		Result:              result,
		Attestation:         attestation,
	}

	err := app.DB.Create(&record).Error
	return &record, err
}
