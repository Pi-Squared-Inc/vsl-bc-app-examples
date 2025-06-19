package models

import (
	"time"
)

type ValidationRecord struct {
	ID                  uint      `gorm:"primaryKey,autoIncrement"`
	Address             string    `gorm:"column:address"`
	CreatedAt           time.Time `gorm:"column:created_at"`
	Computation         *string   `gorm:"column:computation"`
	Input               *string   `gorm:"column:input"`
	Nonce               *string   `gorm:"column:nonce"`
	Result              *string   `gorm:"column:result"`
	Attestation         *string   `gorm:"column:attestation"`
	Claim               *string   `gorm:"column:claim"`
	VerificationContext *string   `gorm:"column:verification_context"`
	ValidationTime      *uint64   `gorm:"column:validation_time"`
	ValidationError     *string   `gorm:"column:validation_error"`
}

func (c *ValidationRecord) GetClaimSize() int {
	if c.Claim == nil {
		return 0
	}
	return len(*c.Claim)
}

func (c *ValidationRecord) GetVerificationContextSize() int {
	if c.VerificationContext == nil {
		return 0
	}
	return len(*c.VerificationContext)
}

func (c *ValidationRecord) ToResponse() ClaimResponse {
	return ClaimResponse{
		ID:                      c.ID,
		Address:                 c.Address,
		ValidationError:         c.ValidationError,
		ValidationTime:          c.ValidationTime,
		CreatedAt:               c.CreatedAt,
		ClaimSize:               c.GetClaimSize(),
		VerificationContextSize: c.GetVerificationContextSize(),
		Result:                  c.Result,
	}
}

type ClaimResponse struct {
	ID                      uint      `json:"id"`
	Address                 string    `json:"address"`
	ValidationError         *string   `json:"validation_error"`
	ValidationTime          *uint64   `json:"validation_time"`
	CreatedAt               time.Time `json:"created_at"`
	ClaimSize               int       `json:"claim_size"`
	VerificationContextSize int       `json:"verification_context_size"`
	Result                  *string   `json:"result"`
}
