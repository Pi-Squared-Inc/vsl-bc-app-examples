package main

import (
	"backend/api"
	"backend/models"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) { // Ignore if .env file does not exist
		log.Fatalf("Error loading environment variables")
	}
	port := os.Getenv("PORT")

	app := models.NewApp()
	if app.DB == nil {
		log.Fatal("Database connection is not initialized")
	}

	// Start the API server
	api.RegisterBlockHeaderAPI(app)
	api.RegisterBlockMirroringAPI(app)
	api.RegisterBlockMirroringBTCAPI(app)
	log.Fatal(app.API.Listen(":" + port))
}
