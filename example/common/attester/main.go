package main

import (
	utils "attester/utils"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"attester/models"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const TPMPath = "/dev/tpm0"

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  start")
		fmt.Println("  machine-ak")
		fmt.Println("  file-hashes --file <file-name>")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "start":
		runServer()
	case "machine-ak":
		runPrintAK()
	case "file-hashes":
		runFileHashes(os.Args[2:])
	default:
		fmt.Println("Unknown command:", os.Args[1])
		os.Exit(1)
	}
}

func runServer() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("failed loading environment variables")
		return
	}
	port := os.Getenv("ATTESTER_PORT")

	tpm, err := tpm2.OpenTPM(TPMPath)
	if err != nil {
		log.Fatalf("failed opening TPM device")
		return
	}
	defer tpm.Close()
	printKey(tpm)

	app := models.NewApp(tpm, &sync.Mutex{})

	app.API.Post("/", app.HandleTEEQuery, app.QueryParserMiddleware, models.MetricsMiddleware)
	app.API.Head("/health_check", func(c fiber.Ctx) error {
		return c.Status(200).SendString("OK")
	})
	app.API.Get("/metrics", adaptor.HTTPHandler(promhttp.HandlerFor(models.MakePromRegistry(), promhttp.HandlerOpts{})))

	log.Println("Attester: starting to listen...")
	err = app.API.Listen(":" + port)
	if err != nil {
		log.Printf("Error on HTTP listening!")
		os.Exit(1)
	}
}

func runPrintAK() {
	tpm, err := tpm2.OpenTPM(TPMPath)
	if err != nil {
		log.Fatalf("failed opening TPM device")
		return
	}
	defer tpm.Close()
	printKey(tpm)
}

func runFileHashes(args []string) {
	fs := flag.NewFlagSet("file-hashes", flag.ExitOnError)
	fileName := fs.String("file", "", "File to hash")
	fs.Parse(args)

	if *fileName == "" {
		fmt.Println("Please provide a file name with --file")
		os.Exit(1)
	}

	err := utils.GenerateSHAHashes(*fileName)
	if err != nil {
		log.Fatalf("Failed to generate hashes: %v", err)
	}
}

func printKey(tpm io.ReadWriteCloser) {
	b64Key, err := utils.GetAKB64(tpm)
	if err != nil {
		log.Fatalf("Failed to get AK: %v", err)
		return
	}
	log.Println("AK of this machine:", b64Key)
}
