package seeders

import (
	"log"
)

func SeedAllData() {
	log.Println("=== Starting Database Seeding ===")
	SeedUsers()
	SeedGameSettings()
	log.Println("=== Database Seeding Completed ===")
}
