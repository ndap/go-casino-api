package seeders

import (
	"casino_api_go/config"
	"casino_api_go/models"
	"log"
)

// SeedGameSettings creates default game settings
func SeedGameSettings() {
	log.Println("Seeding game settings...")

	// Check if game settings already exist
	var count int64
	config.DB.Model(&models.GameSettings{}).Count(&count)

	if count > 0 {
		log.Println("Game settings already exist, skipping...")
		return
	}

	// Create default game settings
	gameSettings := models.GameSettings{
		MaxMultiplier:   100.0,     // Maximum multiplier 100x
		MinBetAmount:    1000.0,    // Minimum bet 1,000 IDR
		MaxBetAmount:    1000000.0, // Maximum bet 1,000,000 IDR
		MultiplierSpeed: 0.1,       // Multiplier increases by 0.1 every second
		IsActive:        true,      // Game settings is active
	}

	if err := config.DB.Create(&gameSettings).Error; err != nil {
		log.Printf("Error creating game settings: %v", err)
		return
	}

	log.Println("Game settings seeded successfully!")
	log.Printf("Max Multiplier: %.2fx", gameSettings.MaxMultiplier)
	log.Printf("Min Bet Amount: %.2f IDR", gameSettings.MinBetAmount)
	log.Printf("Max Bet Amount: %.2f IDR", gameSettings.MaxBetAmount)
	log.Printf("Multiplier Speed: %.2f per second", gameSettings.MultiplierSpeed)
}

// SeedAllGameData seeds all game-related data
func SeedAllGameData() {
	SeedGameSettings()
}
