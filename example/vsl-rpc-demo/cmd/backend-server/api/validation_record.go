package api

import (
	types "base-tee/pkg/abstract_types"
	"strconv"
	"time"
	"vsl-rpc-demo/cmd/backend-server/models"

	"github.com/gofiber/fiber/v3"
)

func RegisterClaimAPI(app *models.App) {

	// Define the structure for the client-specific claim details within a block for the /claims list
	type ClaimDetails struct {
		CreatedAt               *time.Time `json:"created_at"`
		ClaimSize               int        `json:"claim_size"`
		VerificationContextSize int        `json:"verification_context_size"`
		ValidationTime          uint64     `json:"validation_time"`
		Error                   *string    `json:"error"`
	}

	// Define the structure for the final response element for each block for the /claims list
	type BlockClaimResponse struct {
		BlockNumber  uint64                  `json:"block_number"`
		ClaimDetails map[string]ClaimDetails `json:"claim_details"`
	}

	app.API.Get("/validation_records/:address", func(c fiber.Ctx) error {
		address := c.Params("address")
		if address == "" {
			return c.Status(400).SendString("address is required")
		}

		page := c.Query("page", "0")
		pageSize := c.Query("page_size", "25")
		pageInt, err := strconv.Atoi(page)
		if err != nil || pageInt < 0 {
			return c.Status(400).SendString("invalid page")
		}
		pageSizeInt, err := strconv.Atoi(pageSize)
		if err != nil || pageSizeInt <= 0 {
			return c.Status(400).SendString("invalid page size")
		}
		offset := pageInt * pageSizeInt

		var validationRecords []models.ValidationRecord
		err = app.DB.Model(&models.ValidationRecord{}).
			Where("address = ?", address).
			Order("created_at DESC").
			Limit(pageSizeInt).
			Offset(offset).
			Find(&validationRecords).Error

		if err != nil {
			return c.Status(500).SendString("internal server error")
		}

		var total int64
		err = app.DB.Model(&models.ValidationRecord{}).
			Where("address = ?", address).
			Count(&total).Error

		if err != nil {
			return c.Status(500).SendString("internal server error")
		}

		validationRecordsResponse := make([]models.ClaimResponse, 0)
		for _, validationRecord := range validationRecords {
			validationRecordsResponse = append(validationRecordsResponse, validationRecord.ToResponse())
		}

		return c.JSON(fiber.Map{
			"validation_records": validationRecordsResponse,
			"total":              total,
		})
	})

	app.API.Post("/validate", func(c fiber.Ctx) error {
		type requestBody struct {
			Address *string                                       `json:"address"`
			Claim   *types.TEEComputationClaim                    `json:"claim"`
			VerCtx  *types.TEEComputationClaimVerificationContext `json:"proof"`
		}
		req := new(requestBody)
		if err := c.Bind().JSON(req); err != nil {
			return c.Status(400).SendString("invalid request body")
		}

		if req.Address == nil {
			return c.Status(400).SendString("address is required")
		}
		if req.Claim == nil {
			return c.Status(400).SendString("claim is required")
		}
		if req.VerCtx == nil {
			return c.Status(400).SendString("verificatoin context is required")
		}

		return validate(c, app, req.Address, req.Claim, req.VerCtx)
	})

	app.API.Get("/validation_records/:id/claim", func(c fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			return c.Status(400).SendString("id is required")
		}

		dbValidationRecord := models.ValidationRecord{}
		err := app.DB.Where("id = ?", id).First(&dbValidationRecord).Error
		if err != nil {
			return c.Status(404).SendString("validation record not found")
		}

		return c.JSON(fiber.Map{
			"claim": *dbValidationRecord.Claim,
			"size":  dbValidationRecord.GetClaimSize(),
		})
	})

	app.API.Get("/validation_records/:id/verification_context", func(c fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			return c.Status(400).SendString("id is required")
		}

		dbValidationRecord := models.ValidationRecord{}
		err := app.DB.Where("id = ?", id).First(&dbValidationRecord).Error
		if err != nil {
			return c.Status(404).SendString("validation record not found")
		}

		if dbValidationRecord.VerificationContext == nil {
			return c.JSON(fiber.Map{
				"verification_context": nil,
			})
		}

		return c.JSON(fiber.Map{
			"verification_context": *dbValidationRecord.VerificationContext,
			"size":                 dbValidationRecord.GetVerificationContextSize(),
		})
	})
}
