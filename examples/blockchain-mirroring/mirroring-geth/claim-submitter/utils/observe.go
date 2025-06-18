package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"generation-block-processing-evm/pkg/generation"
	"log"
	"mirroring-geth-claim-submitter/models"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

func ObserveBlocks(app *models.App) error {
	ctx := context.Background()

	// Get chain ID
	chainIdBig, err := app.EthRPCClient.ChainID(ctx)
	if err != nil {
		log.Fatalf("Failed to get chain ID: %+v", err)
	}
	chainId := hexutil.EncodeBig(chainIdBig)
	log.Printf("Chain ID: %s\n", chainId)

	// Subscribe to new logs
	headerChannel := make(chan *types.Header)
	headerSubscribe, err := app.EthWSClient.SubscribeNewHead(ctx, headerChannel)
	if err != nil {
		log.Printf("Error subscribing to new headers: %+v", err)
		return err
	}

	for {
		select {
		case err := <-headerSubscribe.Err():
			log.Printf("Error subscribing to new headers: %+v", err)
			return err
		case header := <-headerChannel:
			log.Printf("New header detected: %s", header.Number.String())

			// Generate block processing claim
			claim, verCtx, err := generation.Generate(app.EthRPCClient, header.Number)
			if err != nil {
				errString := fmt.Sprintf("Error generating block claim: %+v", err)
				log.Print(errString)
				err = SubmitClaimToBackend(app, header.Number.Uint64(), nil, &errString)
				if err != nil {
					log.Printf("Error submitting claim to backend: %+v", err)
				}
				continue
			}

			// Marshall claim and verification context for validation
			_, err = json.Marshal(claim)
			if err != nil {
				errString := fmt.Sprintf("Error marshalling claim: %+v", err)
				log.Print(errString)
				err = SubmitClaimToBackend(app, header.Number.Uint64(), nil, &errString)
				if err != nil {
					log.Printf("Error submitting claim to backend: %+v", err)
				}
				continue
			}

			_, err = json.Marshal(verCtx)
			if err != nil {
				errString := fmt.Sprintf("Error marshalling verification context: %+v", err)
				log.Print(errString)
				err = SubmitClaimToBackend(app, header.Number.Uint64(), nil, &errString)
				if err != nil {
					log.Printf("Error submitting claim to backend: %+v", err)
				}
				continue
			}

			claimId, err := SubmitClaimToVSL(app, header.Number.Uint64(), claim, verCtx, nil)
			if err != nil {
				errString := fmt.Sprintf("Error submitting claim to VSL: %+v", err)
				err = SubmitClaimToBackend(app, header.Number.Uint64(), nil, &errString)
				if err != nil {
					log.Printf("Error submitting claim to backend: %+v", err)
				}
				continue
			}
			log.Printf("Successfully submitted claim for block %s to VSL with ID %s", header.Number.String(), *claimId)

			// Submit block processing claim to remote RPC (for verifier to fetch)
			err = SubmitClaimToBackend(app, header.Number.Uint64(), claimId, nil)
			if err != nil {
				log.Printf("Error submitting claim to backend: %+v", err)
			} else {
				log.Printf("Successfully submitted claim for block %s to backend", header.Number.String())
			}
		}
	}
}
