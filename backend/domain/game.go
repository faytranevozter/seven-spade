package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- Game ---

type GameDBStatus string

const (
	GameDBStatusInProgress GameDBStatus = "in_progress"
	GameDBStatusFinished   GameDBStatus = "finished"
)

type Game struct {
	ID           uuid.UUID      `json:"id"            gorm:"type:uuid;primarykey"`
	RoomID       uuid.UUID      `json:"room_id"       gorm:"type:uuid;index"`
	AceDirection string         `json:"ace_direction"` // undecided, low, high
	Status       GameDBStatus   `json:"status"        gorm:"default:in_progress"`
	StartedAt    time.Time      `json:"started_at"`
	EndedAt      *time.Time     `json:"ended_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-"             gorm:"index"`

	// Relations
	Players []GamePlayer `json:"players,omitempty" gorm:"foreignKey:GameID"`
	Moves   []GameMove   `json:"moves,omitempty"   gorm:"foreignKey:GameID"`
}

type GamePlayer struct {
	ID            uuid.UUID `json:"id"             gorm:"type:uuid;primarykey"`
	GameID        uuid.UUID `json:"game_id"        gorm:"type:uuid;index"`
	UserID        uuid.UUID `json:"user_id"        gorm:"type:uuid;index"`
	Seat          int       `json:"seat"`
	IsBot         bool      `json:"is_bot"         gorm:"default:false"`
	BotDifficulty string    `json:"bot_difficulty"`
	PenaltyPoints int       `json:"penalty_points" gorm:"default:0"`
	FinalRank     int       `json:"final_rank"     gorm:"default:0"`

	// Relations
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type GameMove struct {
	ID        uuid.UUID `json:"id"         gorm:"type:uuid;primarykey"`
	GameID    uuid.UUID `json:"game_id"    gorm:"type:uuid;index"`
	Seat      int       `json:"seat"`
	MoveNum   int       `json:"move_num"`
	MoveType  string    `json:"move_type"` // play, face_down
	CardSuit  string    `json:"card_suit"`
	CardRank  int       `json:"card_rank"`
	CreatedAt time.Time `json:"created_at"`
}

// --- Player Stats ---

type PlayerStats struct {
	UserID            uuid.UUID `json:"user_id"             gorm:"type:uuid;primarykey"`
	TotalGames        int       `json:"total_games"         gorm:"default:0"`
	Wins              int       `json:"wins"                gorm:"default:0"`
	Losses            int       `json:"losses"              gorm:"default:0"`
	TotalPenaltyPoints int      `json:"total_penalty_points" gorm:"default:0"`
	EloRating         int       `json:"elo_rating"          gorm:"default:1000"`
	UpdatedAt         time.Time `json:"updated_at"`

	// Relations
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// --- Filters ---

type GameFilter struct {
	ID     *uuid.UUID
	RoomID *uuid.UUID
	Status *GameDBStatus
}

func (f *GameFilter) Query(q *gorm.DB) {
	if f.ID != nil {
		q.Where("id = ?", *f.ID)
	}
	if f.RoomID != nil {
		q.Where("room_id = ?", *f.RoomID)
	}
	if f.Status != nil {
		q.Where("status = ?", *f.Status)
	}
}
