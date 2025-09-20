package main

import (
	"casino_api_go/config"
	"casino_api_go/config/seeders"
	"casino_api_go/controllers"
	"casino_api_go/routes"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDB()
	seeders.SeedAllData()

	router := gin.Default()

	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	routes.SetupAuthRoutes(router)
	routes.SetupProtectedRoutes(router)
	routes.SetupAdminRoutes(router)
	routes.SetupCasinoRoutes(router)

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		controllers.CleanupExpiredBlacklistedTokens()

		for range ticker.C {
			controllers.CleanupExpiredBlacklistedTokens()
		}
	}()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "Casino API is running",
		})
	})

	log.Println("Server starting on port 8080...")
	if err := router.Run("0.0.0.0:8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
