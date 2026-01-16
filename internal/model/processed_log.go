package model

import "gorm.io/gorm"

// ProcessedLog represents the last processed block in the blockchain.
type ProcessedLog struct {
	gorm.Model
	BlockNumber uint64 `gorm:"uniqueIndex"` // Ensure unique block numbers
}
