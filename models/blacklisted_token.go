package models

import (
	"time"

	"gorm.io/gorm"
)

type BlacklistedToken struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Token     string         `json:"token" gorm:"uniqueIndex:idx_blacklisted_tokens_token,length:255;not null"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}
