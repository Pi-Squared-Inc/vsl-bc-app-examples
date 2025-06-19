package models

import (
	"time"

	"github.com/google/uuid"
)

type VerificationRecord struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserAddress string    `gorm:"column:user_address"`
	Type        string    `gorm:"column:type"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	Status      string    `gorm:"column:status"`
	ClaimID     *string   `gorm:"column:claim_id"`
	Result      *string   `gorm:"column:result"`
}

type VerificationResponse struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
	ClaimID   *string   `json:"claim_id"`
	Result    *string   `json:"result"`
}

func (v *VerificationRecord) ToResponse() VerificationResponse {
	return VerificationResponse{
		ID:        v.ID.String(),
		Type:      v.Type,
		CreatedAt: v.CreatedAt,
		Status:    v.Status,
		ClaimID:   v.ClaimID,
		Result:    v.Result,
	}
}
