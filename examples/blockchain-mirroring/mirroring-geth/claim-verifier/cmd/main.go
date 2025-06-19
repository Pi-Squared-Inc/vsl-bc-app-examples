package main

import (
	"base/pkg/abstract_types"
	"encoding/json"
	"fmt"
	"log"
	"mirroring-geth-claim-verifier/models"
	"mirroring-geth-claim-verifier/utils"
	"os"
	"time"
	"verification-block-processing-evm/pkg/verification"

	"github.com/joho/godotenv"

	"base/pkg/vsl"

	generationModels "generation-block-processing-evm/pkg/models"
)

func main() {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) { // Ignore if .env file does not exist
		log.Fatalf("Error loading environment variables")
	}

	app := models.NewApp()

	fmt.Println("Start verifier for VSL(", app.VSLRPC, ") with verifier address: ", app.VerifierAddress)

	since := abstract_types.Timestamp{
		Seconds: uint64(time.Now().Unix()),
		Nanos:   0,
	}

	for {
		fmt.Println("Since: seconds: ", since.Seconds, "nanos: ", since.Nanos)

		claims, err := app.VSLRPCClient.ListSubmittedClaimsForReceiver(vsl.ListSubmittedClaimsForReceiverParams{
			Since:   since,
			Address: app.VerifierAddress,
		})
		if err != nil {
			log.Printf("Error getting request claims for address: %v", err)
			continue
		}

		for _, claim := range claims {
			claimId := claim.Get("id").String()
			claimInformations := claim.Get("data")

			claimTimestampSeconds := claim.Get("timestamp").Get("seconds").Uint()
			claimTimestampNanos := claim.Get("timestamp").Get("nanos").Uint()

			// Unmarshal claim
			claimBytesString := claimInformations.Get("claim").String()
			var claim generationModels.EVMBlockProcessingClaim
			err := json.Unmarshal([]byte(claimBytesString), &claim)
			if err != nil {
				errString := fmt.Sprintf("Error unmarshalling claim: %v", err)
				log.Println(errString)
				err = utils.SubmitClaimToBackend(app, &claimId, nil, &errString)
				if err != nil {
					log.Printf("Error submitting claim to backend: %v", err)
					continue
				}
				continue
			}

			// Unmarshal proof
			proofBytesString := claimInformations.Get("proof").String()
			var proof generationModels.EVMBlockProcessingClaimVerificationContext
			err = json.Unmarshal([]byte(proofBytesString), &proof)
			if err != nil {
				errString := fmt.Sprintf("Error unmarshalling proof: %v", err)
				log.Println(errString)
				err = utils.SubmitClaimToBackend(app, &claimId, nil, &errString)
				if err != nil {
					log.Printf("Error submitting claim to backend: %v", err)
					continue
				}
				continue
			}

			// Verify claim and record verification time
			var verificationTime uint64
			verificationTimeStart := time.Now()
			err = verification.Verify(&claim, &proof)
			if err != nil {
				errString := fmt.Sprintf("Error verifying claim: %v", err)
				log.Println(errString)
				err = utils.SubmitClaimToBackend(app, &claimId, nil, &errString)
				if err != nil {
					log.Printf("Error submitting claim to backend: %v", err)
					continue
				}
				continue
			}
			verificationTime = uint64(time.Since(verificationTimeStart).Microseconds())

			// Settle claim to VSL
			settledClaimId, err := utils.SettleClaimToVSL(app, claimId)
			if err != nil {
				errString := fmt.Sprintf("Error settling claim: %v", err)
				log.Println(errString)
				err = utils.SubmitClaimToBackend(app, &claimId, nil, &errString)
				if err != nil {
					log.Printf("Error submitting claim to backend: %v", err)
					continue
				}
				continue
			}
			fmt.Println("Settled claim: ", *settledClaimId)

			// Submit claim to backend
			err = utils.SubmitClaimToBackend(app, settledClaimId, &verificationTime, nil)
			if err != nil {
				log.Printf("Error submitting claim to backend: %v", err)
				continue
			}

			if claimTimestampSeconds >= since.Seconds {
				since.Seconds = claimTimestampSeconds
				if uint32(claimTimestampNanos) > since.Nanos {
					since.Nanos = uint32(claimTimestampNanos)
				}
				since.Tick()
			}

		}

		time.Sleep(10 * time.Second)
	}
}
