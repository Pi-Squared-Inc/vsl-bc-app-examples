package main

import (
	"backend/api"
	"backend/clients"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) { // Ignore if .env file does not exist
		log.Fatalf("Error loading environment variables")
	}

	// Get environment variables
	dbPath := os.Getenv("DATABASE_PATH")

	// Initialize app
	app, err := clients.CreateApp(dbPath)
	if err != nil {
		log.Fatalf("Error initializing API server: %v", err)
	}
	api.RegisterClaimAPI(app)

	// Start the API server
	log.Fatal(app.API.Listen(":3001"))
}
