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

type StartGameRequest struct {
	BetAmount float64 `json:"bet_amount" binding:"required,gt=0"`
}

type StopGameRequest struct {
	GameID uint `json:"game_id" binding:"required"`
}

type GameResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

var (
	activeGames     = make(map[uint]*models.Game)
	activeGamesMux  sync.RWMutex
	crashWorkerOnce sync.Once
)

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

	var user models.User
	if err := config.DB.Preload("Wallet").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, GameResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	if user.Status == "banned" {
		c.JSON(http.StatusForbidden, GameResponse{
			Success: false,
			Message: "Account is banned",
		})
		return
	}

	var settings models.GameSettings
	if err := config.DB.Where("is_active = ?", true).First(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GameResponse{
			Success: false,
			Message: "Game settings not found",
		})
		return
	}

	if req.BetAmount < settings.MinBetAmount || req.BetAmount > settings.MaxBetAmount {
		c.JSON(http.StatusBadRequest, GameResponse{
			Success: false,
			Message: fmt.Sprintf("Bet amount must be between %.2f and %.2f", settings.MinBetAmount, settings.MaxBetAmount),
		})
		return
	}

	if user.Wallet.Balance < req.BetAmount {
		c.JSON(http.StatusBadRequest, GameResponse{
			Success: false,
			Message: "Insufficient wallet balance",
		})
		return
	}

	tx := config.DB.Begin()
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

	crashPoint := simulateGameCrash(settings)
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

	activeGamesMux.Lock()
	activeGames[game.ID] = &game
	activeGamesMux.Unlock()

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

	var game models.Game
	if err := config.DB.First(&game, req.GameID).Error; err != nil {
		c.JSON(http.StatusNotFound, GameResponse{
			Success: false,
			Message: "Game not found",
		})
		return
	}

	if game.UserID != userID {
		c.JSON(http.StatusForbidden, GameResponse{
			Success: false,
			Message: "Access denied",
		})
		return
	}

	if game.IsCompletedAtomically() {
		c.JSON(http.StatusBadRequest, GameResponse{
			Success: false,
			Message: "Game is already completed",
		})
		return
	}

	activeGamesMux.Lock()
	delete(activeGames, game.ID)
	activeGamesMux.Unlock()
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

	if game.UserID != userID {
		c.JSON(http.StatusForbidden, GameResponse{
			Success: false,
			Message: "Access denied",
		})
		return
	}

	var settings models.GameSettings
	if err := config.DB.Where("is_active = ?", true).First(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GameResponse{
			Success: false,
			Message: "Game settings not found",
		})
		return
	}
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

func calculateCurrentMultiplier(createdAt time.Time, speed float64) float64 {
	elapsed := time.Since(createdAt).Seconds()
	multiplier := 1.0 + (elapsed * speed)

	if multiplier > 100.0 {
		multiplier = 100.0
	}

	return multiplier
}

func simulateGameCrash(settings models.GameSettings) float64 {
	crashPoint := settings.MaxMultiplier

	if crashPoint < 1.01 {
		crashPoint = 1.01
	}

	return crashPoint
}

func completeGame(game *models.Game, stopReason string) (*gin.Context, *GameResponse) {
	if !game.TryCompleteGame() {
		return nil, &GameResponse{
			Success: false,
			Message: "Game already completed",
		}
	}

	var user models.User
	if err := config.DB.Preload("Wallet").First(&user, game.UserID).Error; err != nil {
		return nil, &GameResponse{
			Success: false,
			Message: "User not found",
		}
	}

	var settings models.GameSettings
	if err := config.DB.Where("is_active = ?", true).First(&settings).Error; err != nil {
		return nil, &GameResponse{
			Success: false,
			Message: "Game settings not found",
		}
	}

	currentMultiplier := calculateCurrentMultiplier(game.CreatedAt, settings.MultiplierSpeed)
	crashPoint := game.CrashPoint

	tx := config.DB.Begin()
	game.Multiplier = currentMultiplier
	game.IsCompleted = true

	var winAmount float64
	var gameStatus string
	var transactionType string
	var description string

	if currentMultiplier < crashPoint {
		winAmount = game.BetAmount * currentMultiplier
		gameStatus = "won"
		transactionType = "win"
		description = fmt.Sprintf("Game won with multiplier %.2fx", currentMultiplier)
		user.Wallet.Balance += winAmount
	} else {
		winAmount = 0
		gameStatus = "lost"
		transactionType = "loss"
		description = fmt.Sprintf("Game lost - crashed at multiplier %.2fx", crashPoint)
	}

	game.WinAmount = winAmount
	game.Status = gameStatus

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

func startCrashWorker() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		activeGamesMux.RLock()
		gameIDs := make([]uint, 0, len(activeGames))
		for gameID := range activeGames {
			gameIDs = append(gameIDs, gameID)
		}
		activeGamesMux.RUnlock()

		for _, gameID := range gameIDs {
			var game models.Game
			if err := config.DB.First(&game, gameID).Error; err != nil {
				activeGamesMux.Lock()
				delete(activeGames, gameID)
				activeGamesMux.Unlock()
				continue
			}

			if game.IsCompletedAtomically() {
				activeGamesMux.Lock()
				delete(activeGames, gameID)
				activeGamesMux.Unlock()
				continue
			}

			var settings models.GameSettings
			if err := config.DB.Where("is_active = ?", true).First(&settings).Error; err != nil {
				continue
			}

			currentMultiplier := calculateCurrentMultiplier(game.CreatedAt, settings.MultiplierSpeed)
			crashPoint := game.CrashPoint
			if currentMultiplier >= crashPoint {
				activeGamesMux.Lock()
				delete(activeGames, gameID)
				activeGamesMux.Unlock()

				go func(g models.Game) {
					_, response := completeGame(&g, "auto_crash")
					if response != nil && response.Success {
						println("Game auto-crashed: ID", g.ID, "at multiplier", currentMultiplier)
					}
				}(game)
			}
		}
	}
}

func GetActiveGamesStatus(c *gin.Context) {
	activeGamesMux.RLock()
	defer activeGamesMux.RUnlock()

	var activeGamesData []gin.H
	for gameID := range activeGames {
		var freshGame models.Game
		if err := config.DB.First(&freshGame, gameID).Error; err != nil {
			continue
		}

		var settings models.GameSettings
		if err := config.DB.Where("is_active = ?", true).First(&settings).Error; err != nil {
			continue
		}

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

	if game.UserID != userID {
		c.JSON(http.StatusForbidden, GameResponse{
			Success: false,
			Message: "Access denied",
		})
		return
	}

	var settings models.GameSettings
	if err := config.DB.Where("is_active = ?", true).First(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, GameResponse{
			Success: false,
			Message: "Game settings not found",
		})
		return
	}

	currentMultiplier := calculateCurrentMultiplier(game.CreatedAt, settings.MultiplierSpeed)
	crashPoint := game.CrashPoint

	isActive := !game.IsCompleted
	timeToCrash := 0.0
	if isActive && currentMultiplier < crashPoint {
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
