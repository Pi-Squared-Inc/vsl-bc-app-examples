package clients

import (
	"backend/models"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewDatabase(dbPath string) (*gorm.DB, error) {
	// Create the database directory if it doesn't exist
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto migrate the schema
	db.AutoMigrate(&models.ClaimRecord{})

	return db, nil
}
