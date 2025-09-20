package routes

import (
	"casino_api_go/controllers"

	"github.com/gin-gonic/gin"
)

func SetupCasinoRoutes(router *gin.Engine) {
	casino := router.Group("/api/casino")
	casino.Use(controllers.AuthMiddleware())
	{
		casino.POST("/start", controllers.StartGame)
		casino.POST("/stop", controllers.StopGame)
		casino.GET("/games", controllers.GetUserGames)
		casino.GET("/settings", controllers.GetGameSettings)

		casino.GET("/active-games", controllers.GetActiveGamesStatus)
		casino.GET("/game/:id/crash-info", controllers.GetGameCrashInfo)

		casino.GET("/game/:id", controllers.GetGameStatus)
	}
}
