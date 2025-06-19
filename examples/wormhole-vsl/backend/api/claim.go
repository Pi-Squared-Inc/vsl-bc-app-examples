package api

import (
	"backend/clients"
	"backend/models"
	"strings"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

type QueryClaimsParameter struct {
	Address string `json:"address" form:"address" query:"address"`
	Limit   int    `json:"limit" form:"limit" query:"limit"`
}

type UpsertClaimParameter struct {
	Address                    string `json:"address" form:"address" query:"address"`
	SourceTransactionHash      string `json:"source_transaction_hash" form:"source_transaction_hash" query:"source_transaction_hash"`
	DestinationTransactionHash string `json:"destination_transaction_hash" form:"destination_transaction_hash" query:"destination_transaction_hash"`
	ClaimId                    string `json:"claim_id" form:"claim_id" query:"claim_id"`
	Claim                      string `json:"claim" form:"claim" query:"claim"`
	ClaimJSON                  string `json:"claim_json" form:"claim_json" query:"claim_json"`
}

func RegisterClaimAPI(app *clients.App) {
	app.API.Get("/claims", func(c fiber.Ctx) error {
		queryParams := new(QueryClaimsParameter)
		if err := c.Bind().Query(queryParams); err != nil {
			return err
		}
		var claims []models.ClaimRecord
		limit := queryParams.Limit
		if limit == 0 {
			limit = 10
		}
		dbQuery := app.DB.Limit(limit)
		if queryParams.Address != "" {
			dbQuery = dbQuery.Where("address = ?", strings.ToLower(queryParams.Address))
		}
		err := dbQuery.Order("created_at DESC").Find(&claims).Error
		if err != nil {
			return c.Status(500).SendString("internal server error")
		}
		return c.JSON(claims)
	})

	app.API.Get("/claim-by-source-tx/:source_tx_hash", func(c fiber.Ctx) error {
		sourceTxHash := c.Params("source_tx_hash", "")
		if sourceTxHash == "" {
			return c.Status(400).SendString("source_tx_hash is required")
		}

		var claim models.ClaimRecord
		err := app.DB.Where("source_transaction_hash = ?", sourceTxHash).First(&claim).Error
		if err != nil {
			return c.Status(500).SendString("internal server error")
		}
		return c.JSON(claim)
	})

	app.API.Get("/claim/:id", func(c fiber.Ctx) error {
		id := c.Params("id", "")

		if id == "" {
			return c.Status(400).SendString("claim id is required")
		}

		var claim models.ClaimRecord
		err := app.DB.Where("claim_id = ?", id).First(&claim).Error
		if err != nil {
			return c.Status(500).SendString("internal server error")
		}
		return c.JSON(claim)
	})

	app.API.Put("/claim-by-source-tx/:source_tx_hash", func(c fiber.Ctx) error {
		upsertClaimParams := new(UpsertClaimParameter)
		if err := c.Bind().Body(upsertClaimParams); err != nil {
			return err
		}

		sourceTxHash := c.Params("source_tx_hash", "")
		if sourceTxHash == "" {
			return c.Status(400).SendString("source_tx_hash is required")
		}

		var claim models.ClaimRecord
		err := app.DB.Where("source_transaction_hash = ?", sourceTxHash).First(&claim).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				claim = models.ClaimRecord{
					SourceTransactionHash: sourceTxHash,
				}
				err = app.DB.Create(&claim).Error
				if err != nil {
					return c.Status(500).SendString("internal server error")
				}
			} else {
				return c.Status(500).SendString("internal server error")
			}
		}

		upsertClaim(&claim, upsertClaimParams)

		err = app.DB.Save(&claim).Error
		if err != nil {
			return c.Status(500).SendString("internal server error")
		}
		return c.JSON(claim)
	})

	app.API.Put("/claim/:id", func(c fiber.Ctx) error {
		upsertClaimParams := new(UpsertClaimParameter)
		if err := c.Bind().Body(upsertClaimParams); err != nil {
			return err
		}

		claimId := c.Params("id", "")
		if claimId == "" {
			return c.Status(400).SendString("claim_id is required")
		}

		var claim models.ClaimRecord
		err := app.DB.Where("claim_id = ?", claimId).First(&claim).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				claim = models.ClaimRecord{
					ClaimId: claimId,
				}
				err = app.DB.Create(&claim).Error
				if err != nil {
					return c.Status(500).SendString("internal server error")
				}
			} else {
				return c.Status(500).SendString("internal server error")
			}
		}

		upsertClaim(&claim, upsertClaimParams)
		err = app.DB.Save(&claim).Error
		if err != nil {
			return c.Status(500).SendString("internal server error")
		}
		return c.JSON(claim)
	})
}

func upsertClaim(claim *models.ClaimRecord, upsertClaimParams *UpsertClaimParameter) {
	if upsertClaimParams.Address != "" {
		claim.Address = upsertClaimParams.Address
	}
	if upsertClaimParams.SourceTransactionHash != "" {
		claim.SourceTransactionHash = upsertClaimParams.SourceTransactionHash
	}
	if upsertClaimParams.DestinationTransactionHash != "" {
		claim.DestinationTransactionHash = upsertClaimParams.DestinationTransactionHash
	}
	if upsertClaimParams.ClaimId != "" {
		claim.ClaimId = upsertClaimParams.ClaimId
	}
	if upsertClaimParams.Claim != "" {
		claim.Claim = upsertClaimParams.Claim
	}
	if upsertClaimParams.ClaimJSON != "" {
		claim.ClaimJSON = upsertClaimParams.ClaimJSON
	}
}
