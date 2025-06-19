package clients

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/client"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"gorm.io/gorm"
)

type App struct {
	DB        *gorm.DB
	API       *fiber.App
	APIClient *client.Client
}

type structValidator struct {
	validate *validator.Validate
}

// Validator needs to implement the Validate method
func (v *structValidator) Validate(out any) error {
	return v.validate.Struct(out)
}

func CreateApp(dbPath string) (*App, error) {
	fiberApp := fiber.New(fiber.Config{
		StructValidator: &structValidator{validate: validator.New()},
	})
	fiberApp.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
	}))
	db, err := NewDatabase(dbPath)
	if err != nil {
		return nil, err
	}
	return &App{
		DB:        db,
		API:       fiberApp,
		APIClient: client.New(),
	}, nil
}
