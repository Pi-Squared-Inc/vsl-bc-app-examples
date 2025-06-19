package models

import (
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type App struct {
	API *fiber.App
	DB  *gorm.DB
}

func NewApp() *App {
	// Create data folder if it doesn't exist
	os.MkdirAll("data", 0755)

	db, err := gorm.Open(sqlite.Open("data/db.sqlite"), &gorm.Config{TranslateError: true})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&BlockHeaderRecord{}, &BlockMirroringRecord{}, &BlockMirroringBTCRecord{})

	apiServer := fiber.New(fiber.Config{
		BodyLimit: 100 * 1024 * 1024, // 100MB
	})

	// CORS
	apiServer.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	return &App{
		API: apiServer,
		DB:  db,
	}
}
