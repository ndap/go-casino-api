package controllers

import (
	"casino_api_go/config"
	"casino_api_go/models"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// GetProfile returns user profile information
func GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

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
		Message: "Profile retrieved successfully",
		Data: gin.H{
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"role":     user.Role,
				"status":   user.Status,
			},
			"wallet": gin.H{
				"balance":  user.Wallet.Balance,
				"currency": user.Wallet.Currency,
			},
		},
	})
}

// UpdateProfileRequest structure for profile updates
type UpdateProfileRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
}

// UpdateProfile updates user profile information
func UpdateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req UpdateProfileRequest
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

	// Check if username or email already exists (if being updated)
	if req.Username != "" && req.Username != user.Username {
		var existingUser models.User
		if err := config.DB.Where("username = ? AND id != ?", req.Username, userID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, AuthResponse{
				Success: false,
				Message: "Username already exists",
			})
			return
		}
		user.Username = req.Username
	}

	if req.Email != "" && req.Email != user.Email {
		var existingUser models.User
		if err := config.DB.Where("email = ? AND id != ?", req.Email, userID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, AuthResponse{
				Success: false,
				Message: "Email already exists",
			})
			return
		}
		user.Email = req.Email
	}

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to update profile",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Profile updated successfully",
		Data: gin.H{
			"user": gin.H{
				"id":       user.ID,
				"username": user.Username,
				"email":    user.Email,
				"role":     user.Role,
				"status":   user.Status,
			},
		},
	})
}

// GetWallet returns user wallet information
func GetWallet(c *gin.Context) {
	userID := c.GetUint("user_id")

	var wallet models.Wallet
	if err := config.DB.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		c.JSON(http.StatusNotFound, AuthResponse{
			Success: false,
			Message: "Wallet not found",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Wallet information retrieved successfully",
		Data: gin.H{
			"wallet": gin.H{
				"id":       wallet.ID,
				"balance":  wallet.Balance,
				"currency": wallet.Currency,
			},
		},
	})
}

// ChangePasswordRequest structure for password change
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// ChangePassword allows users to change their password
func ChangePassword(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req ChangePasswordRequest
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

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Current password is incorrect",
		})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to process new password",
		})
		return
	}

	user.Password = string(hashedPassword)
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to update password",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Password changed successfully",
	})
}

// DepositRequest structure for deposit request
type DepositRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description,omitempty"`
}

// WithdrawRequest structure for withdraw request
type WithdrawRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description,omitempty"`
}

// Deposit allows users to deposit money to their wallet
func Deposit(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Validate minimum deposit amount
	if req.Amount < 10000 {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Minimum deposit amount is 10,000 IDR",
		})
		return
	}

	// Validate maximum deposit amount
	if req.Amount > 10000000 {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Maximum deposit amount is 10,000,000 IDR",
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

	// Check if user is banned
	if user.Status == "banned" {
		c.JSON(http.StatusForbidden, AuthResponse{
			Success: false,
			Message: "Account is banned, cannot deposit",
		})
		return
	}

	// Use transaction for wallet update and transaction record
	tx := config.DB.Begin()

	// Update wallet balance
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

	// Create transaction record
	transaction := models.Transaction{
		UserID:      userID,
		Type:        "deposit",
		Amount:      req.Amount,
		Balance:     user.Wallet.Balance,
		Description: req.Description,
		Status:      "completed",
		Reference:   generateReferenceNumber(),
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to create transaction record",
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Deposit successful",
		Data: gin.H{
			"transaction": gin.H{
				"id":          transaction.ID,
				"type":        transaction.Type,
				"amount":      transaction.Amount,
				"balance":     transaction.Balance,
				"description": transaction.Description,
				"status":      transaction.Status,
				"reference":   transaction.Reference,
				"created_at":  transaction.CreatedAt,
			},
			"wallet": gin.H{
				"old_balance": oldBalance,
				"new_balance": user.Wallet.Balance,
				"currency":    user.Wallet.Currency,
			},
		},
	})
}

// Withdraw allows users to withdraw money from their wallet
func Withdraw(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Validate minimum withdraw amount
	if req.Amount < 50000 {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Minimum withdraw amount is 50,000 IDR",
		})
		return
	}

	// Validate maximum withdraw amount
	if req.Amount > 5000000 {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Maximum withdraw amount is 5,000,000 IDR",
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

	// Check if user is banned
	if user.Status == "banned" {
		c.JSON(http.StatusForbidden, AuthResponse{
			Success: false,
			Message: "Account is banned, cannot withdraw",
		})
		return
	}

	// Check if user has sufficient balance
	if user.Wallet.Balance < req.Amount {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Insufficient wallet balance",
		})
		return
	}

	// Use transaction for wallet update and transaction record
	tx := config.DB.Begin()

	// Update wallet balance
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

	// Create transaction record
	transaction := models.Transaction{
		UserID:      userID,
		Type:        "withdraw",
		Amount:      req.Amount,
		Balance:     user.Wallet.Balance,
		Description: req.Description,
		Status:      "completed",
		Reference:   generateReferenceNumber(),
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to create transaction record",
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Withdraw successful",
		Data: gin.H{
			"transaction": gin.H{
				"id":          transaction.ID,
				"type":        transaction.Type,
				"amount":      transaction.Amount,
				"balance":     transaction.Balance,
				"description": transaction.Description,
				"status":      transaction.Status,
				"reference":   transaction.Reference,
				"created_at":  transaction.CreatedAt,
			},
			"wallet": gin.H{
				"old_balance": oldBalance,
				"new_balance": user.Wallet.Balance,
				"currency":    user.Wallet.Currency,
			},
		},
	})
}

// GetTransactionHistory returns user's transaction history
func GetTransactionHistory(c *gin.Context) {
	userID := c.GetUint("user_id")

	// Get pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	transactionType := c.Query("type") // Filter by transaction type

	offset := (page - 1) * limit

	var transactions []models.Transaction
	var total int64

	query := config.DB.Model(&models.Transaction{}).Where("user_id = ?", userID)

	// Apply type filter if provided
	if transactionType != "" {
		query = query.Where("type = ?", transactionType)
	}

	// Get total count
	query.Count(&total)

	// Get transactions with pagination
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&transactions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to retrieve transaction history",
		})
		return
	}

	// Format response data
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
		Message: "Transaction history retrieved successfully",
		Data: gin.H{
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

// generateReferenceNumber generates a unique reference number for transactions
func generateReferenceNumber() string {
	return fmt.Sprintf("REF-%d-%d", time.Now().Unix(), rand.Intn(9999))
}
