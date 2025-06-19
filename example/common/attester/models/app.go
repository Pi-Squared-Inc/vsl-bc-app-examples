package models

import (
	"attester/block_processing"
	"attester/inference"
	"attester/utils"
	"io"
	"log"
	"sync"
	"time"

	types "base-tee/pkg/abstract_types"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

type RelyingPartyQuery struct {
	ClaimType   string            `json:"type"`
	Computation types.Computation `json:"computation"`
	Input       []string          `json:"input"`
	Nonce       []byte            `json:"nonce"`
}

type App struct {
	TPM      io.ReadWriteCloser
	TPMMutex *sync.Mutex
	API      *fiber.App
}

func NewApp(tpm io.ReadWriteCloser, tpmMutex *sync.Mutex) *App {
	api := fiber.New(fiber.Config{
		BodyLimit: 100 * 1024 * 1024, // 100MB
	})
	// CORS
	api.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))
	return &App{
		TPM:      tpm,
		TPMMutex: tpmMutex,
		API:      api,
	}
}

func (app *App) QueryParserMiddleware(c fiber.Ctx) error {
	if c.Get("Content-Type") != "application/json" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Content-Type must be application/json",
		})
	}
	var query RelyingPartyQuery
	if err := c.Bind().JSON(&query); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":  "Failed to parse JSON",
			"reason": err.Error(),
		})
	}
	if query.ClaimType != "TEEComputation" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Unknown claim type",
		})
	}
	// Validate computation type
	switch query.Computation {
	case types.InferImageClass:
	case types.InferTextGen:
	case types.BlockProcessingKReth:
	default:
		return c.Status(400).JSON(fiber.Map{
			"error": "Unknown computation type",
		})

	}
	c.Locals("query", query)
	return c.Next()
}

func (app *App) HandleTEEQuery(c fiber.Ctx) error {
	// Try to serve the requested computation
	query := c.Locals("query").(RelyingPartyQuery)
	var err error
	var answer string
	app.TPMMutex.Lock()
	defer app.TPMMutex.Unlock()
	// For logging purposes, keep track of duration:
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		c.Locals("duration", duration)
	}()
	switch query.Computation {
	case types.InferImageClass:
		if len(query.Input) != 1 {
			return c.Status(400).JSON(fiber.Map{
				"error": "Unexpected number of arguments for img_class",
			})
		}
		// Image classification will also extend the PCR registers
		answer, err = inference.Inference(app.TPM, types.ImageClass, query.Input[0])
	case types.InferTextGen:
		if len(query.Input) != 1 {
			return c.Status(400).JSON(fiber.Map{
				"error": "Unexpected number of arguments for img_class",
			})
		}
		// Text generation will also extend the PCR registers
		answer, err = inference.Inference(app.TPM, types.TextGen, query.Input[0])
	case types.BlockProcessingKReth:
		if len(query.Input) != 2 {
			return c.Status(400).JSON(fiber.Map{
				"error": "Unexpected number of arguments for block_processing_kreth",
			})
		}
		// Block processing using KReth will also extend the PCR registers
		answer, err = block_processing.BlockProcessingKReth(app.TPM, query.Input[0], query.Input[1])
	default:
	}

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  "Computation failed",
			"reason": err.Error(),
		})
	}

	// Get attestation report
	log.Println("Attester: generating attestation report...")
	b64Report, err := utils.FetchAttestationReport(app.TPM, query.Nonce)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":  "Fetching attestation report failed",
			"reason": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"result": answer,
		"report": b64Report,
	})
}
