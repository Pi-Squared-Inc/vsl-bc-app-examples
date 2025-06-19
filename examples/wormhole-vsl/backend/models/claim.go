package models

import "gorm.io/gorm"

type ClaimRecord struct {
	gorm.Model
	Address                    string `json:"address" gorm:"column:address"`
	ClaimId                    string `json:"claim_id" gorm:"column:claim_id"`
	Claim                      string `json:"claim" gorm:"column:claim"`
	ClaimJSON                  string `json:"claim_json" gorm:"column:claim_json"`
	SourceTransactionHash      string `json:"source_transaction_hash" gorm:"column:source_transaction_hash"`
	DestinationTransactionHash string `json:"destination_transaction_hash" gorm:"column:destination_transaction_hash"`
}
