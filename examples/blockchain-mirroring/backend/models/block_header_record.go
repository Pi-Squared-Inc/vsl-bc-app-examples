package models

import (
	"time"
)

type BlockHeaderRecord struct {
	ClaimID             string    `gorm:"primaryKey;column:claim_id"`
	Chain               string    `gorm:"column:chain"`
	BlockNumber         uint64    `gorm:"column:block_number"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	Claim               *string   `gorm:"column:claim"`
	VerificationContext *string   `gorm:"column:verification_context"`
	VerificationTime    *uint64   `gorm:"column:verification_time"`
	Error               *string   `gorm:"column:error"`
}
