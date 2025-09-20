package models

import (
	"gorm.io/gorm"
)

type Wallet struct {
	gorm.Model
	UserID    uint    `gorm:"not null;unique"`
	Balance   float64 `gorm:"not null"`
	Currency  string  `gorm:"not null"`
	User      *User   `gorm:"belongsTo:User"`
}
