package seeders

import (
	"log"
)

// SeedAllData runs all seeders
func SeedAllData() {
	log.Println("=== Starting Database Seeding ===")

	// Seed users first (admin and test user)
	SeedUsers()

	// Seed game settings
	SeedGameSettings()

	log.Println("=== Database Seeding Completed ===")
}
