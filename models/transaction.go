package models

import (
	"gorm.io/gorm"
)

type Transaction struct {
	gorm.Model
	UserID      uint    `gorm:"not null"`
	GameID      *uint   `gorm:"null"` // Null if not related to game
	Type        string  `gorm:"type:enum('bet', 'win', 'loss', 'topup', 'deduct', 'deposit', 'withdraw');not null"`
	Amount      float64 `gorm:"not null"`
	Balance     float64 `gorm:"not null"` // Balance after transaction
	Description string  `gorm:"not null"`
	Status      string  `gorm:"type:enum('pending', 'completed', 'failed', 'cancelled');default:'completed'"` // Status for deposit/withdraw
	Reference   string  `gorm:"null"`                                                                         // Reference number for deposit/withdraw
	User        *User   `gorm:"belongsTo:User"`
	Game        *Game   `gorm:"belongsTo:Game"`
}
