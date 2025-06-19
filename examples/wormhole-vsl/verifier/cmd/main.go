package main

import (
	"base/pkg/abstract_types"
	"base/pkg/evm"
	"fmt"
	"log"
	"os"
	"time"
	"verification-view-fn-evm/pkg/verification"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/joho/godotenv"

	"base/pkg/vsl"

	generationModels "generation-view-fn-evm/pkg/models"
)

func main() {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) { // Ignore if .env file does not exist
		log.Fatalf("Error loading environment variables")
	}

	vslRPC := os.Getenv("VSL_RPC")

	vslVerifierPrivateKey := os.Getenv("VSL_VERIFIER_PRIVATE_KEY")
	verifierAddress, err := evm.AddressFromPrivateKey(vslVerifierPrivateKey)
	if err != nil {
		log.Fatalf("Error getting verifier address: %v", err)
	}

	vslRPCClient := vsl.NewVSLRPCClient(vslRPC, vslVerifierPrivateKey)

	fmt.Println("Start observing VSL(", vslRPC, ") for verifier address: ", verifierAddress)

	since := abstract_types.Timestamp{
		Seconds: uint64(time.Now().Unix()),
		Nanos:   0,
	}

	for {
		fmt.Println("Since: seconds: ", since.Seconds, "nanos: ", since.Nanos)

		claims, err := vslRPCClient.ListSubmittedClaimsForReceiver(vsl.ListSubmittedClaimsForReceiverParams{
			Since:   since,
			Address: verifierAddress.Hex(),
		})
		if err != nil {
			log.Fatalf("Error getting request claims for address: %v", err)
			continue
		}

		for _, claim := range claims {
			claimId := claim.Get("id").String()
			claimInformations := claim.Get("data")

			claimTimestampSeconds := claim.Get("timestamp").Get("seconds").Uint()
			claimTimestampNanos := claim.Get("timestamp").Get("nanos").Uint()

			claimBytesString := claimInformations.Get("claim").String()
			claimBytes := hexutil.MustDecode(claimBytesString)
			claim, err := generationModels.AbiDecodeEVMViewFnClaim(claimBytes)
			if err != nil {
				log.Printf("Error decoding claim: %v", err)
				continue
			}

			proofBytesString := claimInformations.Get("proof").String()
			proofBytes := hexutil.MustDecode(proofBytesString)
			proof, err := generationModels.AbiDecodeEVMViewFnClaimVerificationContext(proofBytes)
			if err != nil {
				log.Printf("Error decoding proof: %v", err)
				continue
			}

			err = verification.Verify(claim, proof)
			if err != nil {
				log.Printf("Error verifying claim: %v", err)
				continue
			}

			nonce, err := vslRPCClient.GetAccountNonce(vsl.GetAccountNonceParams{
				AccountId: verifierAddress.Hex(),
			})
			if err != nil {
				log.Printf("Error getting nonce: %v", err)
				continue
			}

			settledClaimId, err := vslRPCClient.SettleClaim(vsl.SettleClaimParams{
				From:          verifierAddress.Hex(),
				Nonce:         fmt.Sprintf("%d", *nonce),
				TargetClaimId: claimId,
			})

			if err != nil {
				log.Printf("Error settling claim: %v", err)
				continue
			}

			fmt.Println("Settled claim: ", *settledClaimId)

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
