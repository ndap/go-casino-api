package controllers

import (
	"casino_api_go/config"
	"casino_api_go/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("user_role")
		if userRole != "admin" {
			c.JSON(http.StatusForbidden, AuthResponse{
				Success: false,
				Message: "Access denied. Admin privileges required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

func GetAllUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	status := c.Query("status")

	offset := (page - 1) * limit

	var users []models.User
	var total int64

	query := config.DB.Model(&models.User{})

	if search != "" {
		query = query.Where("username LIKE ? OR email LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	if err := query.Preload("Wallet").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to retrieve users",
		})
		return
	}

	var userData []gin.H
	for _, user := range users {
		userData = append(userData, gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
			"status":   user.Status,
			"wallet": gin.H{
				"balance":  user.Wallet.Balance,
				"currency": user.Wallet.Currency,
			},
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Users retrieved successfully",
		Data: gin.H{
			"users": userData,
			"pagination": gin.H{
				"page":       page,
				"limit":      limit,
				"total":      total,
				"total_page": (int(total) + limit - 1) / limit,
			},
		},
	})
}

func GetUserByID(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := config.DB.Preload("Wallet").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, AuthResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "User retrieved successfully",
		Data: gin.H{
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"role":     user.Role,
				"status":   user.Status,
				"wallet": gin.H{
					"balance":  user.Wallet.Balance,
					"currency": user.Wallet.Currency,
				},
				"created_at": user.CreatedAt,
				"updated_at": user.UpdatedAt,
			},
		},
	})
}

type BanUserRequest struct {
	Reason string `json:"reason" binding:"required"`
}

func BanUser(c *gin.Context) {
	userID := c.Param("id")

	var req BanUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, AuthResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	if user.Status == "banned" {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "User is already banned",
		})
		return
	}

	if user.Role == "admin" {
		c.JSON(http.StatusForbidden, AuthResponse{
			Success: false,
			Message: "Cannot ban admin users",
		})
		return
	}

	user.Status = "banned"
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to ban user",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "User banned successfully",
		Data: gin.H{
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"status":   user.Status,
			},
			"ban_reason": req.Reason,
		},
	})
}

func UnbanUser(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, AuthResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	if user.Status != "banned" {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "User is not banned",
		})
		return
	}

	user.Status = "active"
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to unban user",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "User unbanned successfully",
		Data: gin.H{
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"status":   user.Status,
			},
		},
	})
}

type TopUpWalletRequest struct {
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	Currency string  `json:"currency" binding:"required"`
	Note     string  `json:"note"`
}

func TopUpWallet(c *gin.Context) {
	userID := c.Param("id")

	var req TopUpWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	var user models.User
	if err := config.DB.Preload("Wallet").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, AuthResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	if user.Status == "banned" {
		c.JSON(http.StatusForbidden, AuthResponse{
			Success: false,
			Message: "Cannot top up wallet for banned user",
		})
		return
	}

	tx := config.DB.Begin()
	oldBalance := user.Wallet.Balance
	user.Wallet.Balance += req.Amount

	if err := tx.Save(&user.Wallet).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to update wallet",
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Wallet topped up successfully",
		Data: gin.H{
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
			},
			"wallet": gin.H{
				"old_balance":   oldBalance,
				"new_balance":   user.Wallet.Balance,
				"currency":      user.Wallet.Currency,
				"top_up_amount": req.Amount,
			},
			"note": req.Note,
		},
	})
}

type DeductWalletRequest struct {
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	Currency string  `json:"currency" binding:"required"`
	Note     string  `json:"note"`
}

func DeductWallet(c *gin.Context) {
	userID := c.Param("id")

	var req DeductWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	var user models.User
	if err := config.DB.Preload("Wallet").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, AuthResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	if user.Status == "banned" {
		c.JSON(http.StatusForbidden, AuthResponse{
			Success: false,
			Message: "Cannot deduct from banned user's wallet",
		})
		return
	}

	if user.Wallet.Balance < req.Amount {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Insufficient wallet balance",
		})
		return
	}

	tx := config.DB.Begin()
	oldBalance := user.Wallet.Balance
	user.Wallet.Balance -= req.Amount

	if err := tx.Save(&user.Wallet).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to update wallet",
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Wallet deducted successfully",
		Data: gin.H{
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
			},
			"wallet": gin.H{
				"old_balance":     oldBalance,
				"new_balance":     user.Wallet.Balance,
				"currency":        user.Wallet.Currency,
				"deducted_amount": req.Amount,
			},
			"note": req.Note,
		},
	})
}

func GetWalletHistory(c *gin.Context) {
	userID := c.Param("id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	transactionType := c.Query("type")

	offset := (page - 1) * limit

	var user models.User
	if err := config.DB.Preload("Wallet").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, AuthResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	var transactions []models.Transaction
	var total int64

	query := config.DB.Model(&models.Transaction{}).Where("user_id = ?", userID)

	if transactionType != "" {
		query = query.Where("type = ?", transactionType)
	}

	query.Count(&total)
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to retrieve transaction history",
		})
		return
	}

	var transactionData []gin.H
	for _, transaction := range transactions {
		transactionData = append(transactionData, gin.H{
			"id":          transaction.ID,
			"type":        transaction.Type,
			"amount":      transaction.Amount,
			"balance":     transaction.Balance,
			"description": transaction.Description,
			"status":      transaction.Status,
			"reference":   transaction.Reference,
			"created_at":  transaction.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Wallet history retrieved successfully",
		Data: gin.H{
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
			},
			"wallet": gin.H{
				"balance":  user.Wallet.Balance,
				"currency": user.Wallet.Currency,
			},
			"transactions": transactionData,
			"pagination": gin.H{
				"page":       page,
				"limit":      limit,
				"total":      total,
				"total_page": (int(total) + limit - 1) / limit,
			},
		},
	})
}

func GetDashboardStats(c *gin.Context) {
	var totalUsers int64
	var activeUsers int64
	var bannedUsers int64
	var totalBalance float64
	var totalGames int64
	var totalBets float64
	var totalWins float64

	config.DB.Model(&models.User{}).Count(&totalUsers)
	config.DB.Model(&models.User{}).Where("status = ?", "active").Count(&activeUsers)
	config.DB.Model(&models.User{}).Where("status = ?", "banned").Count(&bannedUsers)
	config.DB.Model(&models.Wallet{}).Select("COALESCE(SUM(balance), 0)").Scan(&totalBalance)
	config.DB.Model(&models.Game{}).Count(&totalGames)
	config.DB.Model(&models.Game{}).Select("COALESCE(SUM(bet_amount), 0)").Scan(&totalBets)
	config.DB.Model(&models.Game{}).Select("COALESCE(SUM(win_amount), 0)").Scan(&totalWins)

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Dashboard statistics retrieved successfully",
		Data: gin.H{
			"statistics": gin.H{
				"total_users":   totalUsers,
				"active_users":  activeUsers,
				"banned_users":  bannedUsers,
				"total_balance": totalBalance,
				"total_games":   totalGames,
				"total_bets":    totalBets,
				"total_wins":    totalWins,
				"house_profit":  totalBets - totalWins,
				"currency":      "IDR",
			},
		},
	})
}

type UpdateGameSettingsRequest struct {
	MaxMultiplier   float64 `json:"max_multiplier" binding:"required,gt=1"`
	MinBetAmount    float64 `json:"min_bet_amount" binding:"required,gt=0"`
	MaxBetAmount    float64 `json:"max_bet_amount" binding:"required,gt=0"`
	MultiplierSpeed float64 `json:"multiplier_speed" binding:"required,gt=0"`
	IsActive        bool    `json:"is_active"`
}

func UpdateGameSettings(c *gin.Context) {
	var req UpdateGameSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	if req.MinBetAmount >= req.MaxBetAmount {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Min bet amount must be less than max bet amount",
		})
		return
	}

	if req.MultiplierSpeed > 10.0 {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Multiplier speed cannot exceed 10.0 (too fast)",
		})
		return
	}

	if req.MaxMultiplier < 1.1 || req.MaxMultiplier > 1000.0 {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Max multiplier must be between 1.1x and 1000.0x",
		})
		return
	}

	var settings models.GameSettings
	if err := config.DB.Where("is_active = ?", true).First(&settings).Error; err != nil {
		settings = models.GameSettings{
			MaxMultiplier:   req.MaxMultiplier,
			MinBetAmount:    req.MinBetAmount,
			MaxBetAmount:    req.MaxBetAmount,
			MultiplierSpeed: req.MultiplierSpeed,
			IsActive:        req.IsActive,
		}

		if err := config.DB.Create(&settings).Error; err != nil {
			c.JSON(http.StatusInternalServerError, AuthResponse{
				Success: false,
				Message: "Failed to create game settings",
			})
			return
		}
	} else {
		settings.MaxMultiplier = req.MaxMultiplier
		settings.MinBetAmount = req.MinBetAmount
		settings.MaxBetAmount = req.MaxBetAmount
		settings.MultiplierSpeed = req.MultiplierSpeed
		settings.IsActive = req.IsActive

		if err := config.DB.Save(&settings).Error; err != nil {
			c.JSON(http.StatusInternalServerError, AuthResponse{
				Success: false,
				Message: "Failed to update game settings",
			})
			return
		}
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Game settings updated successfully",
		Data: gin.H{
			"settings": gin.H{
				"id":               settings.ID,
				"max_multiplier":   settings.MaxMultiplier,
				"min_bet_amount":   settings.MinBetAmount,
				"max_bet_amount":   settings.MaxBetAmount,
				"multiplier_speed": settings.MultiplierSpeed,
				"is_active":        settings.IsActive,
			},
		},
	})
}

func GetAdminGameSettings(c *gin.Context) {
	var settings models.GameSettings
	if err := config.DB.Where("is_active = ?", true).First(&settings).Error; err != nil {
		c.JSON(http.StatusNotFound, AuthResponse{
			Success: false,
			Message: "Game settings not found",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Game settings retrieved successfully",
		Data: gin.H{
			"settings": gin.H{
				"id":               settings.ID,
				"max_multiplier":   settings.MaxMultiplier,
				"min_bet_amount":   settings.MinBetAmount,
				"max_bet_amount":   settings.MaxBetAmount,
				"multiplier_speed": settings.MultiplierSpeed,
				"is_active":        settings.IsActive,
			},
		},
	})
}

func GetAllGames(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")

	offset := (page - 1) * limit

	var games []models.Game
	var total int64

	query := config.DB.Model(&models.Game{}).Preload("User")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&games).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to retrieve games",
		})
		return
	}

	var gameData []gin.H
	for _, game := range games {
		gameData = append(gameData, gin.H{
			"id": game.ID,
			"user": gin.H{
				"id":       game.User.ID,
				"username": game.User.Username,
				"email":    game.User.Email,
			},
			"bet_amount":   game.BetAmount,
			"multiplier":   game.Multiplier,
			"win_amount":   game.WinAmount,
			"status":       game.Status,
			"is_completed": game.IsCompleted,
			"created_at":   game.CreatedAt,
			"updated_at":   game.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Games retrieved successfully",
		Data: gin.H{
			"games": gameData,
			"pagination": gin.H{
				"page":       page,
				"limit":      limit,
				"total":      total,
				"total_page": (int(total) + limit - 1) / limit,
			},
		},
	})
}
