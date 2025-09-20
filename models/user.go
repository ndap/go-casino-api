package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username     string        `gorm:"not null;unique"`
	Password     string        `gorm:"not null"`
	Email        string        `gorm:"not null;unique"`
	Role         string        `gorm:"type:enum('admin', 'user');default:'user'"`
	Status       string        `gorm:"type:enum('active', 'banned');default:'active'"`
	Wallet       *Wallet       `gorm:"hasOne:Wallet"`
	Games        []Game        `gorm:"hasMany:Game"`
	Transactions []Transaction `gorm:"hasMany:Transaction"`
}
