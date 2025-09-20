package routes

import (
	"casino_api_go/controllers"

	"github.com/gin-gonic/gin"
)

// SetupAuthRoutes configures authentication routes
func SetupAuthRoutes(router *gin.Engine) {
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
	}
}

// SetupProtectedRoutes configures protected routes that require authentication
func SetupProtectedRoutes(router *gin.Engine) {
	protected := router.Group("/api")
	protected.Use(controllers.AuthMiddleware())
	{
		// User profile routes
		protected.GET("/profile", controllers.GetProfile)
		protected.PUT("/profile", controllers.UpdateProfile)
		protected.PUT("/change-password", controllers.ChangePassword)
		protected.POST("/logout", controllers.Logout)

		// Wallet routes
		protected.GET("/wallet", controllers.GetWallet)

		// Deposit and Withdraw routes
		protected.POST("/deposit", controllers.Deposit)
		protected.POST("/withdraw", controllers.Withdraw)
		protected.GET("/transactions", controllers.GetTransactionHistory)
	}
}
