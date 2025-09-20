package models

import (
	"sync/atomic"

	"gorm.io/gorm"
)

type Game struct {
	gorm.Model
	UserID      uint    `gorm:"not null"`
	BetAmount   float64 `gorm:"not null"`
	Multiplier  float64 `gorm:"not null;default:1.0"`
	WinAmount   float64 `gorm:"not null;default:0"`
	CrashPoint  float64 `gorm:"not null;default:0"` // Crash point determined at game start
	Status      string  `gorm:"type:enum('active', 'won', 'lost');default:'active'"`
	IsCompleted bool    `gorm:"not null;default:false"`
	User        *User   `gorm:"belongsTo:User"`

	// Atomic field for race condition prevention (not stored in DB)
	completedFlag int32 `gorm:"-"` // -1 = not completed, 1 = completed
}

// TryCompleteGame attempts to mark the game as completed atomically
// Returns true if successfully marked as completed, false if already completed
func (g *Game) TryCompleteGame() bool {
	return atomic.CompareAndSwapInt32(&g.completedFlag, 0, 1)
}

// IsCompletedAtomically checks if game is completed using atomic operation
func (g *Game) IsCompletedAtomically() bool {
	return atomic.LoadInt32(&g.completedFlag) == 1
}

type GameSettings struct {
	gorm.Model
	MaxMultiplier   float64 `gorm:"not null;default:100.0"`
	MinBetAmount    float64 `gorm:"not null;default:1000"`
	MaxBetAmount    float64 `gorm:"not null;default:1000000"`
	MultiplierSpeed float64 `gorm:"not null;default:0.1"` // How fast multiplier increases
	IsActive        bool    `gorm:"not null;default:true"`
}
