package main

import (
	"fmt"
	"log"
	"observer/api"
	"observer/models"
	"observer/utils"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) { // Ignore if .env file does not exist
		log.Fatalf("Error loading environment variables")
	}

	app := models.NewApp()

	mode := os.Getenv("MODE")

	if mode == "auto" {
		utils.ObserveViewFnAutoMode(app)
	} else if mode == "manual" {
		api.RegisterGenerateClaimAPI(app)
		log.Fatal(app.API.Listen(fmt.Sprintf(":%s", app.Port)))
	} else {
		log.Fatal("Invalid mode")
	}
}
