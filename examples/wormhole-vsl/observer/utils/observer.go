package utils

import (
	"context"
	"generation-view-fn-evm/pkg/generation"
	"log"
	"observer/models"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// ObserveViewFnAutoMode observes the source chain for USL contract function calls and generates claims for state queries
//
// Parameters:
// - app: The application context
func ObserveViewFnAutoMode(app *models.App) error {
	ctx := context.Background()

	chainIdBig, err := app.EthRPCClient.ChainID(ctx)
	if err != nil {
		return err
	}
	chainId := hexutil.EncodeBig(chainIdBig)
	log.Printf("Chain ID: %s\n", chainId)

	// Print source bridge contract details
	log.Printf("Source contract function: %s\n", app.SourceVSLContractFunction)
	log.Printf("Source contract address: %s\n", app.SourceVSLContractAddress)

	// Subscribe to new logs
	newLogsChannel := make(chan types.Log)
	newLogsSubscribe, err := app.EthWSClient.SubscribeFilterLogs(ctx, ethereum.FilterQuery{
		Addresses: []common.Address{*app.SourceVSLContractAddress},
		Topics:    [][]common.Hash{{crypto.Keccak256Hash([]byte(app.SourceVSLContractFunction))}},
	}, newLogsChannel)
	if err != nil {
		return err
	}

	for {
		select {
		case err := <-newLogsSubscribe.Err():
			return err
		case newLog := <-newLogsChannel:
			log.Printf("New log detected")

			// Generate claim
			claim, verCtx, err := generation.Generate(app.EthRPCClient, newLog, *app.SourceVSLContractAddress, app.SourceVSLContractABIJSON)
			if err != nil {
				log.Printf("Failed to generate state query claim\nError: %+v", err)
				continue
			}

			// Submit claim to VSL
			claimId, claimHex, _, err := SubmitClaimToVSL(app, claim, verCtx)
			if err != nil {
				log.Printf("%v", err)
				continue
			}

			// Submit claim to Backend
			err = SubmitClaimToBackend(app, newLog.TxHash.Hex(), *claimId, claim, *claimHex)
			if err != nil {
				log.Printf("%v", err)
				continue
			}
		}
	}
}
