package api

import (
	"backend/models"
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

func RegisterBlockMirroringAPI(app *models.App) {
	// Define the structure for the client-specific claim details within a block for the /claims list
	type ClaimDetails struct {
		ClaimID          string     `json:"claim_id"`
		CreatedAt        *time.Time `json:"created_at"`
		VerificationTime *uint64    `json:"verification_time"`
		Error            *string    `json:"error"`
	}

	// Define the structure for the final response element for each block for the /claims list
	type BlockClaimResponse struct {
		BlockNumber  uint64                  `json:"block_number"`
		ClaimDetails map[string]ClaimDetails `json:"claim_details"`
	}

	// Endpoint to get all block mirroring records with pagination
	app.API.Get("/block_mirroring_records", func(c fiber.Ctx) error {
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

		// Get distinct block numbers for the current page.
		var blockNumbers []uint64
		err = app.DB.Model(&models.BlockMirroringRecord{}).
			Select("DISTINCT block_number").
			Order("block_number DESC").
			Limit(pageSizeInt).
			Offset(offset).
			Pluck("block_number", &blockNumbers).Error

		if err != nil {
			return c.Status(500).SendString("internal server error")
		}

		// Prepare the final response structure.
		finalResponse := make([]BlockClaimResponse, 0, len(blockNumbers))

		if len(blockNumbers) == 0 {
			// If no blocks on this page, return empty list but still get total count.
			var total int64
			countQuery := "SELECT COUNT(DISTINCT block_number) FROM block_mirroring_records"
			// Use a raw query to count distinct block numbers.
			if countErr := app.DB.Raw(countQuery).Scan(&total).Error; countErr != nil {
				return c.Status(500).SendString("internal server error")
			}
			return c.JSON(fiber.Map{
				"records": []BlockClaimResponse{},
				"total":   total,
			})
		}

		// Get client details for the blocks on this page.
		type ClientClaimInfo struct { // Local struct for query result
			BlockNumber      uint64
			ClaimID          string
			ExecutionClient  string
			VerificationTime *uint64
			Error            *string
			CreatedAt        *time.Time
		}

		var clientInfos []ClientClaimInfo
		err = app.DB.Model(&models.BlockMirroringRecord{}).
			Select("block_number, execution_client, claim_id, verification_time, error, created_at").
			Where("block_number IN ?", blockNumbers).
			Order("block_number DESC, execution_client ASC"). // Consistent order helps grouping
			Find(&clientInfos).Error

		if err != nil {
			return c.Status(500).SendString("internal server error")
		}

		// Group client details by block number.
		processedDetails := make(map[uint64]map[string]ClaimDetails)
		for _, info := range clientInfos {
			if _, ok := processedDetails[info.BlockNumber]; !ok {
				processedDetails[info.BlockNumber] = make(map[string]ClaimDetails)
			}
			processedDetails[info.BlockNumber][info.ExecutionClient] = ClaimDetails{
				ClaimID:          info.ClaimID,
				VerificationTime: info.VerificationTime,
				Error:            info.Error,
				CreatedAt:        info.CreatedAt,
			}
		}

		// Construct the final response array, preserving block number order.
		for _, blockNum := range blockNumbers {
			detailsMap, ok := processedDetails[blockNum]
			if !ok { // Ensure map exists even if block had no claims (edge case)
				detailsMap = make(map[string]ClaimDetails)
			}
			finalResponse = append(finalResponse, BlockClaimResponse{
				BlockNumber:  blockNum,
				ClaimDetails: detailsMap,
			})
		}

		// Get total count of distinct blocks for pagination.
		var total int64
		countQuery := "SELECT COUNT(DISTINCT block_number) FROM block_mirroring_records"
		if countErr := app.DB.Raw(countQuery).Scan(&total).Error; countErr != nil {
			return c.Status(500).SendString("internal server error")
		}

		return c.JSON(fiber.Map{
			"records": finalResponse,
			"total":   total,
		})
	})

	// Endpoint to create a new block mirroring record or update an existing one
	app.API.Post("/block_mirroring_record", func(c fiber.Ctx) error {
		type requestBody struct {
			BlockNumber      *uint64 `json:"block_number"`
			ExecutionClient  *string `json:"execution_client"`
			ClaimID          *string `json:"claim_id"`
			Error            *string `json:"error"`
			VerificationTime *uint64 `json:"verification_time"`
		}
		req := new(requestBody)
		if err := c.Bind().JSON(req); err != nil {
			return c.Status(400).SendString("invalid request body")
		}

		if req.ExecutionClient == nil {
			return c.Status(400).SendString("execution_client is required")
		}

		// check if the claim ID is provided
		if req.ClaimID == nil {
			return c.Status(400).SendString("claim_id is required")
		}

		// query for existing record
		var existingRecord models.BlockMirroringRecord
		err := app.DB.Where("execution_client = ? AND claim_id = ?", req.ExecutionClient, req.ClaimID).First(&existingRecord).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// create a new record if it doesn't exist
				if req.BlockNumber == nil {
					return c.Status(400).SendString("block_number is required")
				}

				if req.VerificationTime != nil || req.Error != nil {
					return c.Status(400).SendString("cannot set verification_time or error when creating a new record")
				}

				newRecord := models.BlockMirroringRecord{
					BlockNumber:     *req.BlockNumber,
					ExecutionClient: *req.ExecutionClient,
					ClaimID:         *req.ClaimID,
				}

				err = app.DB.Create(&newRecord).Error

				if err != nil {
					if errors.Is(err, gorm.ErrDuplicatedKey) {
						return c.Status(400).SendString("claim_id already exists for this execution client")
					}

					return c.Status(500).SendString("internal server error")
				}
				return c.JSON(newRecord.ToResponse())
			}

			return c.Status(500).SendString("internal server error")
		}

		// If a record already exists, update it
		if req.VerificationTime != nil && req.Error != nil {
			return c.Status(400).SendString("cannot set both verification_time and error for this record")
		}

		if req.VerificationTime != nil {
			if existingRecord.VerificationTime != nil {
				return c.Status(400).SendString("verification_time already set for this record")
			}
			existingRecord.VerificationTime = req.VerificationTime
		} else if req.Error != nil {
			if existingRecord.Error != nil {
				return c.Status(400).SendString("error already set for this record")
			}
			existingRecord.Error = req.Error
		} else {
			return c.Status(400).SendString("either verification_time or error must be provided")
		}

		err = app.DB.Save(&existingRecord).Error
		if err != nil {
			return c.Status(500).SendString("internal server error")
		}
		return c.JSON(existingRecord.ToResponse())
	})

	app.API.Get("/block_mirroring_records/:block_number/:execution_client/claim", func(c fiber.Ctx) error {
		// TODO: Implement this endpoint to return the claim details for a given block number and execution client.
		return c.Status(501).SendString("not implemented yet")
		// blockNumber := c.Params("block_number")
		// if blockNumber == "" {
		// 	return c.Status(400).SendString("block_number is required")
		// }

		// executionClient := c.Params("execution_client")
		// if executionClient == "" {
		// 	return c.Status(400).SendString("execution_client is required")
		// }

		// dbVerificationRecord := models.VerificationRecord{}
		// err := app.DB.Where("block_number = ? AND execution_client = ? AND verification_time IS NOT NULL", blockNumber, executionClient).First(&dbVerificationRecord).Error
		// if err != nil {
		// 	return c.Status(404).SendString("verification record not found")
		// }

		// // TODO: use VSL RPC pkg
		// type rpcRequest struct {
		// 	JSONRPC string      `json:"jsonrpc"`
		// 	Method  string      `json:"method"`
		// 	Params  interface{} `json:"params"`
		// 	ID      int         `json:"id"`
		// }
		// type rpcParams struct {
		// 	ClaimID string `json:"claim_id"`
		// }
		// type rpcResponse struct {
		// 	JSONRPC string          `json:"jsonrpc"`
		// 	Result  json.RawMessage `json:"result"`
		// 	Error   *struct {
		// 		Code    int    `json:"code"`
		// 		Message string `json:"message"`
		// 	} `json:"error,omitempty"`
		// 	ID int `json:"id"`
		// }

		// if dbVerificationRecord.ClaimID == "" {
		// 	return c.Status(400).SendString("claim_id is missing in verification record")
		// }

		// reqBody := rpcRequest{
		// 	JSONRPC: "2.0",
		// 	Method:  "vsl_getSettledClaimById",
		// 	Params:  rpcParams{ClaimID: dbVerificationRecord.ClaimID},
		// 	ID:      1,
		// }

		// jsonBytes, err := json.Marshal(reqBody)
		// if err != nil {
		// 	return c.Status(500).SendString("failed to marshal rpc request")
		// }

		// httpResp, err := http.Post(app.VSLEndpoint, "application/json", bytes.NewReader(jsonBytes))
		// if err != nil {
		// 	return c.Status(502).SendString("failed to contact claim RPC server")
		// }
		// defer httpResp.Body.Close()

		// respBytes, err := io.ReadAll(httpResp.Body)
		// if err != nil {
		// 	return c.Status(500).SendString("failed to read rpc response")
		// }

		// var rpcResp rpcResponse
		// if err := json.Unmarshal(respBytes, &rpcResp); err != nil {
		// 	return c.Status(500).SendString("failed to parse rpc response")
		// }
		// if rpcResp.Error != nil {
		// 	return c.Status(502).SendString("rpc error: " + rpcResp.Error.Message)
		// }

		// // Return the claim as-is (raw JSON)
		// var claimWithBlock map[string]interface{}
		// if err := json.Unmarshal(rpcResp.Result, &claimWithBlock); err != nil {
		// 	return c.Status(500).SendString("failed to parse claim result")
		// }

		// return c.JSON(claimWithBlock)
	})

	app.API.Get("/block_mirroring_records/:block_number/:execution_client/verification_time", func(c fiber.Ctx) error {
		blockNumber := c.Params("block_number")
		if blockNumber == "" {
			return c.Status(400).SendString("block_number is required")
		}

		executionClient := c.Params("execution_client")
		if executionClient == "" {
			return c.Status(400).SendString("execution_client is required")
		}

		dbValidationRecord := models.BlockMirroringRecord{}
		err := app.DB.Where("block_number = ? AND execution_client = ?", blockNumber, executionClient).First(&dbValidationRecord).Error
		if err != nil {
			return c.Status(404).SendString("block mirroring record not found")
		}

		if dbValidationRecord.VerificationTime == nil {
			return c.Status(400).SendString("verification time is not set")
		}

		return c.SendString(strconv.FormatUint(*dbValidationRecord.VerificationTime, 10))
	})
}
