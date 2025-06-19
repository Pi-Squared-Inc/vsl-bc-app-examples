package api

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"
	"vsl-rpc-demo/cmd/backend-server/models"
	"vsl-rpc-demo/cmd/client"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const CHAR_LIMIT int = 100

func RegisterVerificationRecordAPI(app *models.App, clientApp *client.App) {
	supported_compute_types := []string{"img_class", "plain_text"}
	app.API.Post("/compute", func(c fiber.Ctx) error {
		compute_type := c.FormValue("type")
		if compute_type == "" {
			return c.Status(400).SendString("compute type is required")
		}

		if !contains(supported_compute_types, compute_type) {
			return c.Status(400).SendString("unsupported compute type")
		}

		attester, err := app.LoadBalancer.GetNextAttester()
		sender_address := c.FormValue("sender_address")
		if sender_address == "" {
			return c.Status(400).SendString("sender address is required")
		}
		payment_claim_id := c.FormValue("payment_claim_id")
		if payment_claim_id == "" {
			return c.Status(400).SendString("payment claim id is required")
		}

		if err != nil {
			log.Println(err.Error())
			return c.Status(400).SendString(err.Error())
		}
		// Check if user has other pending requests for the same computation type
		can_submit, err := checkCanSubmit(app, sender_address, compute_type)
		if err != nil {
			log.Println(err.Error())
			return c.Status(500).SendString("database query failed")
		}
		if !can_submit {
			return fmt.Errorf("please wait for your other %s request to finish", compute_type)
		}

		switch compute_type {
		case "img_class":
			return handleImageClassification(c, app, clientApp, attester, sender_address, payment_claim_id)
		case "plain_text":
			return handlePlainTextComputation(c, app, clientApp, attester, sender_address, payment_claim_id)
		default:
			return c.Status(400).SendString("unsupported compute type")
		}
	})
	app.API.Get("/get_backend_address", func(c fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"address": app.ClientAddress,
		})
	})
	app.API.Get("/can_submit/:compute_type/:address", func(c fiber.Ctx) error {
		compute_type := c.Params("compute_type")
		if !contains(supported_compute_types, compute_type) {
			return c.Status(400).SendString("unsupported compute type")
		}
		sender_address := c.Params("address")
		can_submit, err := checkCanSubmit(app, sender_address, compute_type)
		if err != nil {
			log.Println(err.Error())
			return c.Status(500).SendString("database query failed")
		}
		return c.Status(200).JSON(fiber.Map{
			"allowed": can_submit,
		})
	})
	app.API.Get("/verification_records/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		if id == "" {
			return c.Status(400).SendString("id is required")
		}
		var record models.VerificationRecord
		err := app.DB.Where("id = ?", id).First(&record).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return c.Status(404).SendString("verification record not found")
			}
			log.Println("error fetching verification record: ", err)
			return c.Status(500).SendString("internal server error")
		}
		return c.JSON(record.ToResponse())
	})
	app.API.Get("/verification_records", func(c fiber.Ctx) error {
		page := c.Query("page", "0")
		pageSize := c.Query("pageSize", "25")
		pageInt, err := strconv.Atoi(page)
		if err != nil || pageInt < 0 {
			return c.Status(400).SendString("invalid page")
		}
		pageSizeInt, err := strconv.Atoi(pageSize)
		if err != nil || pageSizeInt <= 0 {
			return c.Status(400).SendString("invalid page size")
		}
		offset := pageInt * pageSizeInt

		var records []models.VerificationRecord
		err = app.DB.Model(&models.VerificationRecord{}).
			Order("created_at DESC").
			Limit(pageSizeInt).
			Offset(offset).
			Find(&records).Error

		if err != nil {
			return c.Status(500).SendString("internal server error")
		}

		var total int64
		err = app.DB.Model(&models.VerificationRecord{}).
			Count(&total).Error

		if err != nil {
			return c.Status(500).SendString("internal server error")
		}

		verificationRecordsResponse := make([]models.VerificationResponse, 0)
		for _, verificationRecord := range records {
			verificationRecordsResponse = append(verificationRecordsResponse, verificationRecord.ToResponse())
		}

		return c.JSON(fiber.Map{
			"records": verificationRecordsResponse,
			"total":   total,
		})
	})
}

// handleImageClassification handles the image classification computation request.
func handleImageClassification(c fiber.Ctx, app *models.App, clientApp *client.App, attester *models.AttesterEndpoint, sender_address string, payment_claim_id string) error {
	image, err := c.FormFile("image")
	if err != nil {
		log.Println("error getting image file: ", err)
		return c.Status(400).SendString("image file is required")
	}

	if image == nil {
		return c.Status(400).SendString("image file is required")
	}

	if image.Size == 0 {
		return c.Status(400).SendString("image file cannot be empty")
	}
	hash := c.FormValue("hash")
	if hash == "" {
		return c.
			Status(400).SendString("request must be signed")
	}
	r := c.FormValue("r")
	if r == "" {
		return c.Status(400).SendString("request must be signed")
	}
	s := c.FormValue("s")
	if s == "" {
		return c.Status(400).SendString("request must be signed")
	}
	vStr := c.FormValue("v")
	if vStr == "" {
		return c.Status(400).SendString("request must be signed")
	}
	vInt, err := strconv.Atoi(vStr)
	if err != nil || (vInt != 0 && vInt != 1 && vInt != 27 && vInt != 28) {
		return c.Status(400).SendString("invalid v value")
	}
	v := hexutil.Encode([]byte{byte(vInt)})

	// Open the image file
	file, err := image.Open()
	if err != nil {
		log.Println("error opening image file: ", err)
		return c.Status(500).SendString("internal server error")
	}
	defer file.Close()

	imageData, err := io.ReadAll(file)
	if err != nil {
		log.Println("error reading image file: ", err)
		return c.Status(500).SendString("internal server error")
	}

	base64Image := base64.StdEncoding.EncodeToString(imageData)
	computationInput := []string{base64Image}

	record := models.VerificationRecord{
		ID:          uuid.New(),
		UserAddress: sender_address,
		Status:      "queued",
		Type:        "img_class",
		CreatedAt:   time.Now(),
	}

	err = app.DB.Create(&record).Error

	if err != nil {
		log.Println("error creating verification record: ", err)
		return c.Status(500).SendString("internal server error")
	}

	sha256 := sha256.Sum256(imageData)
	go func() {
		err := models.SubmitWithRetry(app.WorkerPool1, models.SendToAttesterPayload{
			ID:        record.ID.String(),
			App:       app,
			ClientApp: clientApp,
			Attester:  attester,
			Request: models.SignedComputeRequest{
				Computation:    "img_class",
				Input:          computationInput,
				SenderAddress:  sender_address,
				PaymentClaimId: payment_claim_id,
				Hash:           hash,
				R:              r,
				S:              s,
				V:              v,
			},
			InputSHA256: hexutil.Encode(sha256[:])[2:], // get rid of 0x prefix
			Retried:     0,
			MaxRetries:  3,
		}, 30, 10*time.Second)

		if err != nil {
			log.Printf("Failed to submit task1Fn to worker pool: %v", err)
		}
	}()

	return c.JSON(record.ToResponse())
}

func handlePlainTextComputation(c fiber.Ctx, app *models.App, clientApp *client.App, attester *models.AttesterEndpoint, sender_address string, payment_claim_id string) error {
	prompt := c.FormValue("prompt")
	if prompt == "" {
		return c.Status(400).SendString("prompt is required")
	}
	if len(prompt) > CHAR_LIMIT {
		return c.Status(400).SendString("prompt should not exceed " + strconv.Itoa(CHAR_LIMIT) + " characters")
	}
	hash := c.FormValue("hash")
	if hash == "" {
		return c.
			Status(400).SendString("request must be signed")
	}
	r := c.FormValue("r")
	if r == "" {
		return c.Status(400).SendString("request must be signed")
	}
	s := c.FormValue("s")
	if s == "" {
		return c.Status(400).SendString("request must be signed")
	}
	vStr := c.FormValue("v")
	if vStr == "" {
		return c.Status(400).SendString("request must be signed")
	}
	vInt, err := strconv.Atoi(vStr)
	if err != nil || (vInt != 0 && vInt != 1 && vInt != 27 && vInt != 28) {
		return c.Status(400).SendString("invalid v value")
	}
	v := hexutil.Encode([]byte{byte(vInt)})

	record := models.VerificationRecord{
		ID:          uuid.New(),
		UserAddress: sender_address,
		Status:      "queued",
		Type:        "plain_text",
		CreatedAt:   time.Now(),
	}

	err = app.DB.Create(&record).Error

	if err != nil {
		log.Println("error creating verification record: ", err)
		return c.Status(500).SendString("internal server error")
	}

	sha256 := sha256.Sum256([]byte(prompt))
	go func() {
		err := models.SubmitWithRetry(app.WorkerPool1, models.SendToAttesterPayload{
			ID:        record.ID.String(),
			App:       app,
			ClientApp: clientApp,
			Attester:  attester,
			Request: models.SignedComputeRequest{
				Computation:    "text_gen",
				Input:          []string{base64.StdEncoding.EncodeToString([]byte(prompt))},
				SenderAddress:  sender_address,
				PaymentClaimId: payment_claim_id,
				Hash:           hash,
				R:              r,
				S:              s,
				V:              v,
			},
			InputSHA256: hexutil.Encode(sha256[:])[2:], // get rid of 0x prefix
			Retried:     0,
			MaxRetries:  3,
		}, 30, 10*time.Second)

		if err != nil {
			log.Printf("Failed to submit task1Fn to worker pool: %v", err)
		}
	}()

	return c.JSON(record.ToResponse())
}

func checkCanSubmit(app *models.App, sender_address string, compute_type string) (bool, error) {
	// Check if user has other pending requests for the same computation type
	var num_pending_requests int64
	err := app.DB.Model(&models.VerificationRecord{}).
		Where("user_address = ?", sender_address).
		Where("type = ?", compute_type).
		Where("status IN (?)", []string{"queued", "awaiting_settlement"}).
		Count(&num_pending_requests).Error
	if err != nil {
		return false, err
	}
	return num_pending_requests == 0, nil
}
