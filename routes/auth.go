package routes

import (
	"casino_api_go/controllers"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(router *gin.Engine) {
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
	}
}

func SetupProtectedRoutes(router *gin.Engine) {
	protected := router.Group("/api")
	protected.Use(controllers.AuthMiddleware())
	{
		protected.GET("/profile", controllers.GetProfile)
		protected.PUT("/profile", controllers.UpdateProfile)
		protected.PUT("/change-password", controllers.ChangePassword)
		protected.POST("/logout", controllers.Logout)

		protected.GET("/wallet", controllers.GetWallet)

		protected.POST("/deposit", controllers.Deposit)
		protected.POST("/withdraw", controllers.Withdraw)
		protected.GET("/transactions", controllers.GetTransactionHistory)
	}
}
