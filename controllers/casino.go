package controllers

import (
	"casino_api_go/config"
	"casino_api_go/models"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// StartGameRequest structure for starting a game
type StartGameRequest struct {
	BetAmount float64 `json:"bet_amount" binding:"required,gt=0"`
}

// StopGameRequest structure for stopping a game
type StopGameRequest struct {
	GameID uint `json:"game_id" binding:"required"`
}

// GameResponse structure for game responses
type GameResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Global variables for game crash management
var (
	activeGames     = make(map[uint]*models.Game)
	activeGamesMux  sync.RWMutex
	crashWorkerOnce sync.Once
)

// StartGame starts a new casino game
func StartGame(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req StartGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, GameResponse{
			Success: false,
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Get user and wallet
	var user models.User
	if err := config.DB.Preload("Wallet").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, GameResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	// Check if user is banned
	if user.Status == "banned" {
		c.JSON(http.StatusForbidden, GameResponse{
			Success: false,
			Message: "Account is banned",
		})
		return
	}

	// Get game settings
	var settings models.GameSettings
	if err := config.DB.Where("is_active = ?", true).First(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GameResponse{
			Success: false,
			Message: "Game settings not found",
		})
		return
	}

	// Validate bet amount
	if req.BetAmount < settings.MinBetAmount || req.BetAmount > settings.MaxBetAmount {
		c.JSON(http.StatusBadRequest, GameResponse{
			Success: false,
			Message: fmt.Sprintf("Bet amount must be between %.2f and %.2f", settings.MinBetAmount, settings.MaxBetAmount),
		})
		return
	}

	// Check if user has sufficient balance
	if user.Wallet.Balance < req.BetAmount {
		c.JSON(http.StatusBadRequest, GameResponse{
			Success: false,
			Message: "Insufficient wallet balance",
		})
		return
	}

	// Use transaction for game creation and wallet deduction
	tx := config.DB.Begin()

	// Deduct bet amount from wallet
	oldBalance := user.Wallet.Balance
	user.Wallet.Balance -= req.BetAmount

	if err := tx.Save(&user.Wallet).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, GameResponse{
			Success: false,
			Message: "Failed to deduct bet amount",
		})
		return
	}

	// Calculate crash point once at game start
	crashPoint := simulateGameCrash(settings)

	// Create game
	game := models.Game{
		UserID:      userID,
		BetAmount:   req.BetAmount,
		Multiplier:  1.0,
		WinAmount:   0,
		CrashPoint:  crashPoint,
		Status:      "active",
		IsCompleted: false,
	}

	if err := tx.Create(&game).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, GameResponse{
			Success: false,
			Message: "Failed to create game",
		})
		return
	}

	// Create transaction record
	transaction := models.Transaction{
		UserID:      userID,
		GameID:      &game.ID,
		Type:        "bet",
		Amount:      -req.BetAmount,
		Balance:     user.Wallet.Balance,
		Description: "Bet placed for casino game",
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, GameResponse{
			Success: false,
			Message: "Failed to create transaction record",
		})
		return
	}

	tx.Commit()

	// Add game to active games for crash monitoring
	activeGamesMux.Lock()
	activeGames[game.ID] = &game
	activeGamesMux.Unlock()

	// Start crash worker if not already running
	crashWorkerOnce.Do(func() {
		go startCrashWorker()
	})

	c.JSON(http.StatusCreated, GameResponse{
		Success: true,
		Message: "Game started successfully",
		Data: gin.H{
			"game": gin.H{
				"id":         game.ID,
				"bet_amount": game.BetAmount,
				"multiplier": game.Multiplier,
				"status":     game.Status,
			},
			"wallet": gin.H{
				"old_balance": oldBalance,
				"new_balance": user.Wallet.Balance,
				"currency":    user.Wallet.Currency,
			},
		},
	})
}

// StopGame stops an active game and calculates winnings
func StopGame(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req StopGameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, GameResponse{
			Success: false,
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	// Get game
	var game models.Game
	if err := config.DB.First(&game, req.GameID).Error; err != nil {
		c.JSON(http.StatusNotFound, GameResponse{
			Success: false,
			Message: "Game not found",
		})
		return
	}

	// Check if game belongs to user
	if game.UserID != userID {
		c.JSON(http.StatusForbidden, GameResponse{
			Success: false,
			Message: "Access denied",
		})
		return
	}

	// Check if game is already completed using atomic operation
	if game.IsCompletedAtomically() {
		c.JSON(http.StatusBadRequest, GameResponse{
			Success: false,
			Message: "Game is already completed",
		})
		return
	}

	// Remove from active games
	activeGamesMux.Lock()
	delete(activeGames, game.ID)
	activeGamesMux.Unlock()

	// Complete the game
	_, response := completeGame(&game, "manual_stop")
	if response != nil {
		c.JSON(http.StatusOK, *response)
	} else {
		c.JSON(http.StatusInternalServerError, GameResponse{
			Success: false,
			Message: "Failed to complete game",
		})
	}
}

// GetGameStatus returns current game status
func GetGameStatus(c *gin.Context) {
	userID := c.GetUint("user_id")
	gameID := c.Param("id")

	var game models.Game
	if err := config.DB.First(&game, gameID).Error; err != nil {
		c.JSON(http.StatusNotFound, GameResponse{
			Success: false,
			Message: "Game not found",
		})
		return
	}

	// Check if game belongs to user
	if game.UserID != userID {
		c.JSON(http.StatusForbidden, GameResponse{
			Success: false,
			Message: "Access denied",
		})
		return
	}

	// Get game settings
	var settings models.GameSettings
	if err := config.DB.Where("is_active = ?", true).First(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GameResponse{
			Success: false,
			Message: "Game settings not found",
		})
		return
	}

	// Calculate current multiplier if game is still active
	currentMultiplier := game.Multiplier
	if !game.IsCompleted {
		currentMultiplier = calculateCurrentMultiplier(game.CreatedAt, settings.MultiplierSpeed)
	}

	c.JSON(http.StatusOK, GameResponse{
		Success: true,
		Message: "Game status retrieved successfully",
		Data: gin.H{
			"game": gin.H{
				"id":           game.ID,
				"bet_amount":   game.BetAmount,
				"multiplier":   currentMultiplier,
				"win_amount":   game.WinAmount,
				"status":       game.Status,
				"is_completed": game.IsCompleted,
				"created_at":   game.CreatedAt,
			},
		},
	})
}

// GetUserGames returns user's game history
func GetUserGames(c *gin.Context) {
	userID := c.GetUint("user_id")

	var games []models.Game
	if err := config.DB.Where("user_id = ?", userID).Order("created_at DESC").Limit(20).Find(&games).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GameResponse{
			Success: false,
			Message: "Failed to retrieve games",
		})
		return
	}

	var gameData []gin.H
	for _, game := range games {
		gameData = append(gameData, gin.H{
			"id":           game.ID,
			"bet_amount":   game.BetAmount,
			"multiplier":   game.Multiplier,
			"win_amount":   game.WinAmount,
			"status":       game.Status,
			"is_completed": game.IsCompleted,
			"created_at":   game.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, GameResponse{
		Success: true,
		Message: "Games retrieved successfully",
		Data: gin.H{
			"games": gameData,
		},
	})
}

// GetGameSettings returns current game settings
func GetGameSettings(c *gin.Context) {
	var settings models.GameSettings
	if err := config.DB.Where("is_active = ?", true).First(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GameResponse{
			Success: false,
			Message: "Game settings not found",
		})
		return
	}

	c.JSON(http.StatusOK, GameResponse{
		Success: true,
		Message: "Game settings retrieved successfully",
		Data: gin.H{
			"settings": gin.H{
				"min_bet_amount": settings.MinBetAmount,
				"max_bet_amount": settings.MaxBetAmount,
				"max_multiplier": settings.MaxMultiplier,
			},
		},
	})
}

// calculateCurrentMultiplier calculates the current multiplier based on time elapsed
func calculateCurrentMultiplier(createdAt time.Time, speed float64) float64 {
	elapsed := time.Since(createdAt).Seconds()
	multiplier := 1.0 + (elapsed * speed)

	// Ensure multiplier doesn't exceed reasonable limits
	if multiplier > 100.0 {
		multiplier = 100.0
	}

	return multiplier
}

// simulateGameCrash simulates the game crash mechanism
// Game will always crash at the MaxMultiplier set by admin
func simulateGameCrash(settings models.GameSettings) float64 {
	// Game always crashes at the maximum multiplier set by admin
	// This ensures predictable and fair gameplay
	crashPoint := settings.MaxMultiplier

	// Ensure minimum crash point of 1.01x
	if crashPoint < 1.01 {
		crashPoint = 1.01
	}

	return crashPoint
}

// completeGame handles game completion logic
func completeGame(game *models.Game, stopReason string) (*gin.Context, *GameResponse) {
	// Check if game is already completed using atomic operation
	if !game.TryCompleteGame() {
		return nil, &GameResponse{
			Success: false,
			Message: "Game already completed",
		}
	}

	// Get user and wallet
	var user models.User
	if err := config.DB.Preload("Wallet").First(&user, game.UserID).Error; err != nil {
		return nil, &GameResponse{
			Success: false,
			Message: "User not found",
		}
	}

	// Get game settings
	var settings models.GameSettings
	if err := config.DB.Where("is_active = ?", true).First(&settings).Error; err != nil {
		return nil, &GameResponse{
			Success: false,
			Message: "Game settings not found",
		}
	}

	// Calculate current multiplier
	currentMultiplier := calculateCurrentMultiplier(game.CreatedAt, settings.MultiplierSpeed)

	// Use the crash point that was determined at game start
	crashPoint := game.CrashPoint

	// Use transaction for game completion and wallet update
	tx := config.DB.Begin()

	// Update game
	game.Multiplier = currentMultiplier
	game.IsCompleted = true

	var winAmount float64
	var gameStatus string
	var transactionType string
	var description string

	// Check if player stopped before crash
	if currentMultiplier < crashPoint {
		// Player won - stopped before crash
		winAmount = game.BetAmount * currentMultiplier
		gameStatus = "won"
		transactionType = "win"
		description = fmt.Sprintf("Game won with multiplier %.2fx", currentMultiplier)

		// Add winnings to wallet
		user.Wallet.Balance += winAmount
	} else {
		// Game crashed - player lost
		winAmount = 0
		gameStatus = "lost"
		transactionType = "loss"
		description = fmt.Sprintf("Game lost - crashed at multiplier %.2fx", crashPoint)
	}

	game.WinAmount = winAmount
	game.Status = gameStatus

	// Update wallet
	oldBalance := user.Wallet.Balance
	if err := tx.Save(&user.Wallet).Error; err != nil {
		tx.Rollback()
		return nil, &GameResponse{
			Success: false,
			Message: "Failed to update wallet",
		}
	}

	if err := tx.Save(&game).Error; err != nil {
		tx.Rollback()
		return nil, &GameResponse{
			Success: false,
			Message: "Failed to update game",
		}
	}

	// Create transaction record
	transaction := models.Transaction{
		UserID:      game.UserID,
		GameID:      &game.ID,
		Type:        transactionType,
		Amount:      winAmount,
		Balance:     user.Wallet.Balance,
		Description: description,
	}

	if err := tx.Create(&transaction).Error; err != nil {
		tx.Rollback()
		return nil, &GameResponse{
			Success: false,
			Message: "Failed to create transaction record",
		}
	}

	tx.Commit()

	message := "Game won!"
	if gameStatus == "lost" {
		message = "Game lost - crashed!"
	}

	return nil, &GameResponse{
		Success: true,
		Message: message,
		Data: gin.H{
			"game": gin.H{
				"id":          game.ID,
				"bet_amount":  game.BetAmount,
				"multiplier":  game.Multiplier,
				"win_amount":  game.WinAmount,
				"status":      game.Status,
				"crash_point": crashPoint,
				"stop_reason": stopReason,
			},
			"wallet": gin.H{
				"old_balance": oldBalance,
				"new_balance": user.Wallet.Balance,
				"currency":    user.Wallet.Currency,
			},
		},
	}
}

// startCrashWorker monitors active games for automatic crashes
func startCrashWorker() {
	ticker := time.NewTicker(100 * time.Millisecond) // Check every 100ms
	defer ticker.Stop()

	for range ticker.C {
		activeGamesMux.RLock()
		gameIDs := make([]uint, 0, len(activeGames))
		for gameID := range activeGames {
			gameIDs = append(gameIDs, gameID)
		}
		activeGamesMux.RUnlock()

		for _, gameID := range gameIDs {
			// Get fresh game data from database
			var game models.Game
			if err := config.DB.First(&game, gameID).Error; err != nil {
				// Game not found, remove from active games
				activeGamesMux.Lock()
				delete(activeGames, gameID)
				activeGamesMux.Unlock()
				continue
			}

			// Check if game is still active using atomic operation
			if game.IsCompletedAtomically() {
				// Game already completed, remove from active games
				activeGamesMux.Lock()
				delete(activeGames, gameID)
				activeGamesMux.Unlock()
				continue
			}

			// Get game settings
			var settings models.GameSettings
			if err := config.DB.Where("is_active = ?", true).First(&settings).Error; err != nil {
				continue
			}

			// Calculate current multiplier
			currentMultiplier := calculateCurrentMultiplier(game.CreatedAt, settings.MultiplierSpeed)

			// Check if game should crash using predetermined crash point
			crashPoint := game.CrashPoint
			if currentMultiplier >= crashPoint {
				// Game should crash, complete it
				activeGamesMux.Lock()
				delete(activeGames, gameID)
				activeGamesMux.Unlock()

				// Complete game in a goroutine to avoid blocking
				go func(g models.Game) {
					_, response := completeGame(&g, "auto_crash")
					if response != nil && response.Success {
						// Log successful auto-crash
						println("Game auto-crashed: ID", g.ID, "at multiplier", currentMultiplier)
					}
				}(game)
			}
		}
	}
}

// GetActiveGamesStatus returns status of all active games for real-time monitoring
func GetActiveGamesStatus(c *gin.Context) {
	activeGamesMux.RLock()
	defer activeGamesMux.RUnlock()

	var activeGamesData []gin.H
	for gameID := range activeGames {
		// Get fresh game data from database
		var freshGame models.Game
		if err := config.DB.First(&freshGame, gameID).Error; err != nil {
			continue
		}

		// Get game settings
		var settings models.GameSettings
		if err := config.DB.Where("is_active = ?", true).First(&settings).Error; err != nil {
			continue
		}

		// Calculate current multiplier
		currentMultiplier := calculateCurrentMultiplier(freshGame.CreatedAt, settings.MultiplierSpeed)

		activeGamesData = append(activeGamesData, gin.H{
			"game_id":            gameID,
			"user_id":            freshGame.UserID,
			"bet_amount":         freshGame.BetAmount,
			"current_multiplier": currentMultiplier,
			"created_at":         freshGame.CreatedAt,
			"elapsed_time":       time.Since(freshGame.CreatedAt).Seconds(),
		})
	}

	c.JSON(http.StatusOK, GameResponse{
		Success: true,
		Message: "Active games status retrieved successfully",
		Data: gin.H{
			"active_games": activeGamesData,
			"total_active": len(activeGamesData),
		},
	})
}

// GetGameCrashInfo returns crash information for a specific game
func GetGameCrashInfo(c *gin.Context) {
	userID := c.GetUint("user_id")
	gameID := c.Param("id")

	var game models.Game
	if err := config.DB.First(&game, gameID).Error; err != nil {
		c.JSON(http.StatusNotFound, GameResponse{
			Success: false,
			Message: "Game not found",
		})
		return
	}

	// Check if game belongs to user
	if game.UserID != userID {
		c.JSON(http.StatusForbidden, GameResponse{
			Success: false,
			Message: "Access denied",
		})
		return
	}

	// Get game settings
	var settings models.GameSettings
	if err := config.DB.Where("is_active = ?", true).First(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GameResponse{
			Success: false,
			Message: "Game settings not found",
		})
		return
	}

	// Calculate current multiplier and use predetermined crash point
	currentMultiplier := calculateCurrentMultiplier(game.CreatedAt, settings.MultiplierSpeed)
	crashPoint := game.CrashPoint

	// Check if game is still active
	isActive := !game.IsCompleted
	timeToCrash := 0.0
	if isActive && currentMultiplier < crashPoint {
		// Calculate time until crash
		timeToCrash = (crashPoint - currentMultiplier) / settings.MultiplierSpeed
	}

	c.JSON(http.StatusOK, GameResponse{
		Success: true,
		Message: "Game crash info retrieved successfully",
		Data: gin.H{
			"game": gin.H{
				"id":                 game.ID,
				"bet_amount":         game.BetAmount,
				"current_multiplier": currentMultiplier,
				"crash_point":        crashPoint,
				"is_active":          isActive,
				"time_to_crash":      timeToCrash,
				"elapsed_time":       time.Since(game.CreatedAt).Seconds(),
				"multiplier_speed":   settings.MultiplierSpeed,
			},
		},
	})
}
