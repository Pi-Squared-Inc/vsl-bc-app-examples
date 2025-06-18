package api

import (
	"context"
	"generation-view-fn-evm/pkg/generation"
	"log"
	"observer/models"
	"observer/utils"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gofiber/fiber/v3"

	"github.com/tidwall/gjson"
)

func RegisterGenerateClaimAPI(app *models.App) {
	app.API.Post("/generate_claim", func(c fiber.Ctx) error {
		// Request body
		body := c.Body()
		ctx := context.Background()

		// Type
		transactionHash := gjson.GetBytes(body, "transaction_hash").String()
		if transactionHash == "" {
			return c.Status(400).SendString("transaction_hash is required")
		}

		tx, err := app.EthRPCClient.TransactionReceipt(ctx, common.HexToHash(transactionHash))
		if err != nil {
			return c.Status(400).SendString("failed to get transaction")
		}
		for _, l := range tx.Logs {
			for _, t := range l.Topics {
				if t.Cmp(crypto.Keccak256Hash([]byte(app.SourceVSLContractFunction))) == 0 {
					claim, verCtx, err := generation.Generate(app.EthRPCClient, *l, *app.SourceVSLContractAddress, app.SourceVSLContractABIJSON)
					if err != nil {
						log.Printf("Failed to generate claim\nError: %+v", err)
						return c.Status(400).SendString("failed to generate claim")
					}
					claimId, claimHex, _, err := utils.SubmitClaimToVSL(app, claim, verCtx)
					if err != nil {
						log.Printf("Failed to submit claim to VSL\nError: %+v", err)
						return c.Status(400).SendString("failed to submit claim")
					}
					err = utils.SubmitClaimToBackend(app, l.TxHash.Hex(), *claimId, claim, *claimHex)
					if err != nil {
						log.Printf("Failed to submit claim to backend\nError: %+v", err)
						return c.Status(400).SendString("failed to submit claim")
					}
					return c.SendStatus(200)
				}
			}
		}

		return c.SendStatus(400)
	})
}
