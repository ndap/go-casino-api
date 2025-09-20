package seeders

import (
	"casino_api_go/config"
	"casino_api_go/models"
	"log"
)

func SeedGameSettings() {
	log.Println("Seeding game settings...")

	var count int64
	config.DB.Model(&models.GameSettings{}).Count(&count)

	if count > 0 {
		log.Println("Game settings already exist, skipping...")
		return
	}

	gameSettings := models.GameSettings{
		MaxMultiplier:   100.0,
		MinBetAmount:    1000.0,
		MaxBetAmount:    1000000.0,
		MultiplierSpeed: 0.1,
		IsActive:        true,
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

func SeedAllGameData() {
	SeedGameSettings()
}
