package seeders

import (
	"casino_api_go/config"
	"casino_api_go/models"
	"log"

	"golang.org/x/crypto/bcrypt"
)

// SeedUsers creates default users for testing
func SeedUsers() {
	log.Println("Seeding users...")

	// Check if users already exist
	var count int64
	config.DB.Model(&models.User{}).Count(&count)

	if count > 0 {
		log.Println("Users already exist, skipping...")
		return
	}

	// Create admin user
	adminPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	admin := models.User{
		Username: "admin",
		Email:    "admin@casino.com",
		Password: string(adminPassword),
		Role:     "admin",
		Status:   "active",
	}

	if err := config.DB.Create(&admin).Error; err != nil {
		log.Printf("Error creating admin user: %v", err)
		return
	}

	// Create admin wallet
	adminWallet := models.Wallet{
		UserID:   admin.ID,
		Balance:  1000000.0, // 1,000,000 IDR
		Currency: "IDR",
	}

	if err := config.DB.Create(&adminWallet).Error; err != nil {
		log.Printf("Error creating admin wallet: %v", err)
		return
	}

	// Create test user
	userPassword, _ := bcrypt.GenerateFromPassword([]byte("user123"), bcrypt.DefaultCost)
	user := models.User{
		Username: "testuser",
		Email:    "user@casino.com",
		Password: string(userPassword),
		Role:     "user",
		Status:   "active",
	}

	if err := config.DB.Create(&user).Error; err != nil {
		log.Printf("Error creating test user: %v", err)
		return
	}

	// Create user wallet
	userWallet := models.Wallet{
		UserID:   user.ID,
		Balance:  100000.0, // 100,000 IDR
		Currency: "IDR",
	}

	if err := config.DB.Create(&userWallet).Error; err != nil {
		log.Printf("Error creating user wallet: %v", err)
		return
	}

	log.Println("Users seeded successfully!")
	log.Printf("Admin user created: %s (admin@casino.com)", admin.Username)
	log.Printf("Test user created: %s (user@casino.com)", user.Username)
	log.Println("Default passwords: admin123 / user123")
}
