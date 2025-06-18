package api

import (
	"backend/models"
	"encoding/json"
	"github.com/gofiber/fiber/v3"
	"strconv"
	"time"
)

func RegisterBlockHeaderAPI(app *models.App) {

	type BlockHeaderRecordDetails struct {
		ClaimID          string     `json:"claim_id"`
		CreatedAt        *time.Time `json:"created_at"`
		VerificationTime *uint64    `json:"verification_time"`
		Error            *string    `json:"error"`
		Chain            string     `json:"chain"`
	}

	type BlockHeaderRecordResponse struct {
		BlockNumber  uint64                   `json:"block_number"`
		ClaimDetails BlockHeaderRecordDetails `json:"claim_details"`
	}

	app.API.Get("/block_header_records", func(c fiber.Ctx) error {
		page := c.Query("page", "0")
		pageSize := c.Query("page_size", "25")
		chain := c.Query("chain", "")

		pageInt, err := strconv.Atoi(page)
		if err != nil || pageInt < 0 {
			return c.Status(400).SendString("invalid page")
		}
		pageSizeInt, err := strconv.Atoi(pageSize)
		if err != nil || pageSizeInt <= 0 {
			return c.Status(400).SendString("invalid page size")
		}
		offset := pageInt * pageSizeInt

		query := app.DB.Model(&models.BlockHeaderRecord{}).Select("DISTINCT block_number")
		if chain != "" {
			query = query.Where("chain = ?", chain)
		}

		var blockNumbers []uint64
		err = query.Order("block_number DESC").
			Limit(pageSizeInt).
			Offset(offset).
			Pluck("block_number", &blockNumbers).Error

		if err != nil {
			return c.Status(500).SendString("internal server error")
		}

		finalResponse := make([]BlockHeaderRecordResponse, 0, len(blockNumbers))
		if len(blockNumbers) == 0 {
			var total int64
			countQuery := app.DB.Model(&models.BlockHeaderRecord{})
			if chain != "" {
				countQuery = countQuery.Where("chain = ?", chain)
			}
			if countErr := countQuery.Select("COUNT(DISTINCT block_number)").Count(&total).Error; countErr != nil {
				return c.Status(500).SendString("internal server error")
			}
			return c.JSON(fiber.Map{
				"block_header_records": []BlockHeaderRecordResponse{},
				"total":                total,
			})
		}

		var headerInfos []models.BlockHeaderRecord
		detailQuery := app.DB.Model(&models.BlockHeaderRecord{}).
			Select("block_number, claim_id, chain, verification_time, error, created_at, claim").
			Where("block_number IN ?", blockNumbers)
		if chain != "" {
			detailQuery = detailQuery.Where("chain = ?", chain)
		}
		err = detailQuery.Order("block_number DESC, claim_id ASC").
			Find(&headerInfos).Error

		if err != nil {
			return c.Status(500).SendString("internal server error")
		}

		processedDetails := make(map[uint64]BlockHeaderRecordDetails)
		for _, info := range headerInfos {
			processedDetails[info.BlockNumber] = BlockHeaderRecordDetails{
				ClaimID:          info.ClaimID,
				VerificationTime: info.VerificationTime,
				Error:            info.Error,
				CreatedAt:        &info.CreatedAt,
				Chain:            info.Chain,
			}
		}

		for _, blockNum := range blockNumbers {
			claimDetail, ok := processedDetails[blockNum]
			if !ok {
				claimDetail = BlockHeaderRecordDetails{}
			}
			finalResponse = append(finalResponse, BlockHeaderRecordResponse{
				BlockNumber:  blockNum,
				ClaimDetails: claimDetail,
			})
		}

		var total int64
		totalCountQuery := app.DB.Model(&models.BlockHeaderRecord{})
		if chain != "" {
			totalCountQuery = totalCountQuery.Where("chain = ?", chain)
		}
		if countErr := totalCountQuery.Select("COUNT(DISTINCT block_number)").Count(&total).Error; countErr != nil {
			return c.Status(500).SendString("internal server error")
		}

		return c.JSON(fiber.Map{
			"block_header_records": finalResponse,
			"total":                total,
		})
	})

	app.API.Post("/block_header_record", func(c fiber.Ctx) error {
		type requestBody struct {
			ClaimID          *string `json:"claim_id"`
			Chain            *string `json:"chain"`
			BlockNumber      *uint64 `json:"block_number"`
			Claim            *string `json:"claim"`
			Error            *string `json:"error"`
			VerificationTime *uint64 `json:"verification_time"`
		}
		req := new(requestBody)
		if err := c.Bind().JSON(req); err != nil {
			return c.Status(400).SendString("invalid request body")
		}

		if req.ClaimID == nil {
			return c.Status(400).SendString("claim_id is required")
		}

		if req.Chain == nil {
			return c.Status(400).SendString("chain is required")
		}

		if req.BlockNumber == nil {
			return c.Status(400).SendString("block_number is required")
		}

		dbHeaderRecord := models.BlockHeaderRecord{
			ClaimID:          *req.ClaimID,
			Chain:            *req.Chain,
			BlockNumber:      *req.BlockNumber,
			Claim:            req.Claim,
			VerificationTime: req.VerificationTime,
			Error:            req.Error,
		}

		err := app.DB.Create(&dbHeaderRecord).Error
		if err != nil {
			return c.Status(500).SendString("internal server error")
		}

		return c.JSON(dbHeaderRecord)
	})

	app.API.Get("/block_header_records/:claim_id/claim", func(c fiber.Ctx) error {
		claimID := c.Params("claim_id")
		if claimID == "" {
			return c.Status(400).SendString("claim_id is required")
		}

		dbHeaderRecord := models.BlockHeaderRecord{}
		err := app.DB.Where("claim_id = ?", claimID).First(&dbHeaderRecord).Error
		if err != nil {
			return c.Status(404).SendString("block header record not found")
		}

		if dbHeaderRecord.Claim == nil {
			return c.Status(404).SendString("claim not found")
		}

		var claimData any
		err = json.Unmarshal([]byte(*dbHeaderRecord.Claim), &claimData)
		if err != nil {
			return c.Status(400).SendString("invalid claim")
		}

		return c.JSON(claimData)
	})

	app.API.Put("/block_header_records/:claim_id/verification_time", func(c fiber.Ctx) error {
		type requestBody struct {
			VerificationTime *uint64 `json:"verification_time"`
		}
		req := new(requestBody)
		if err := c.Bind().JSON(req); err != nil {
			return c.Status(400).SendString("invalid request body")
		}

		claimID := c.Params("claim_id")
		if claimID == "" {
			return c.Status(400).SendString("claim_id is required")
		}

		if req.VerificationTime == nil {
			return c.Status(400).SendString("verification_time is required")
		}

		var record models.BlockHeaderRecord
		err := app.DB.Where("claim_id = ?", claimID).First(&record).Error
		if err != nil {
			return c.Status(404).SendString("block header record not found")
		}

		result := app.DB.Model(&models.BlockHeaderRecord{}).
			Where("claim_id = ?", claimID).
			Update("verification_time", *req.VerificationTime)
		if result.Error != nil {
			return c.Status(500).SendString("internal server error")
		}

		return c.JSON(fiber.Map{
			"claim_id":          claimID,
			"verification_time": *req.VerificationTime,
		})
	})
}
