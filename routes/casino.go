package routes

import (
	"casino_api_go/controllers"

	"github.com/gin-gonic/gin"
)

// SetupCasinoRoutes configures casino game routes
func SetupCasinoRoutes(router *gin.Engine) {
	casino := router.Group("/api/casino")
	casino.Use(controllers.AuthMiddleware()) // Require authentication
	{
		// Game management
		casino.POST("/start", controllers.StartGame)
		casino.POST("/stop", controllers.StopGame)
		casino.GET("/games", controllers.GetUserGames)
		casino.GET("/settings", controllers.GetGameSettings)

		// Crash monitoring endpoints (harus sebelum /game/:id)
		casino.GET("/active-games", controllers.GetActiveGamesStatus)
		casino.GET("/game/:id/crash-info", controllers.GetGameCrashInfo)

		// Game status (harus terakhir karena lebih general)
		casino.GET("/game/:id", controllers.GetGameStatus)
	}
}
