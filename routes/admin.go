package routes

import (
	"casino_api_go/controllers"

	"github.com/gin-gonic/gin"
)

// SetupAdminRoutes configures admin routes that require admin privileges
func SetupAdminRoutes(router *gin.Engine) {
	admin := router.Group("/api/admin")
	admin.Use(controllers.AuthMiddleware())  // Require authentication
	admin.Use(controllers.AdminMiddleware()) // Require admin role
	{
		// Dashboard
		admin.GET("/dashboard", controllers.GetDashboardStats)

		// User management
		admin.GET("/users", controllers.GetAllUsers)
		admin.GET("/users/:id", controllers.GetUserByID)
		admin.POST("/users/:id/ban", controllers.BanUser)
		admin.POST("/users/:id/unban", controllers.UnbanUser)

		// Wallet management
		admin.POST("/users/:id/wallet/topup", controllers.TopUpWallet)
		admin.POST("/users/:id/wallet/deduct", controllers.DeductWallet)
		admin.GET("/users/:id/wallet/history", controllers.GetWalletHistory)

		// Game management
		admin.GET("/games", controllers.GetAllGames)
		admin.GET("/game-settings", controllers.GetAdminGameSettings)
		admin.PUT("/game-settings", controllers.UpdateGameSettings)
	}
}
