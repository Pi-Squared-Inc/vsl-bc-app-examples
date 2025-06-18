package models

import (
	"time"
)

type BlockMirroringBTCRecord struct {
	BlockNumber      uint64    `gorm:"primaryKey"`
	ExecutionClient  string    `gorm:"primaryKey;column:execution_client"`
	ClaimID          string    `gorm:"uniqueIndex:idx_btc_claim;column:claim_id"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	VerificationTime *uint64   `gorm:"column:verification_time"`
	Error            *string   `gorm:"column:error"`
}

type BlockMirroringBTCRecordResponse struct {
	BlockNumber      uint64    `json:"block_number"`
	ExecutionClient  string    `json:"execution_client"`
	ClaimID          string    `json:"claim_id"`
	Error            *string   `json:"error"`
	VerificationTime *uint64   `json:"verification_time"`
	CreatedAt        time.Time `json:"created_at"`
}

func (c *BlockMirroringBTCRecord) ToResponse() BlockMirroringBTCRecordResponse {
	return BlockMirroringBTCRecordResponse{
		BlockNumber:      c.BlockNumber,
		ExecutionClient:  c.ExecutionClient,
		ClaimID:          c.ClaimID,
		Error:            c.Error,
		VerificationTime: c.VerificationTime,
		CreatedAt:        c.CreatedAt,
	}
}
