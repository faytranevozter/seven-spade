package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- Room ---

type RoomStatus string

const (
	RoomStatusWaiting  RoomStatus = "waiting"
	RoomStatusPlaying  RoomStatus = "playing"
	RoomStatusFinished RoomStatus = "finished"
)

type Room struct {
	ID         uuid.UUID      `json:"id"          gorm:"type:uuid;primarykey"`
	Code       string         `json:"code"        gorm:"uniqueIndex;size:6"`
	HostUserID uuid.UUID      `json:"host_user_id" gorm:"type:uuid;index"`
	Status     RoomStatus     `json:"status"      gorm:"default:waiting"`
	BotEnabled bool           `json:"bot_enabled" gorm:"default:false"`
	BotCount   int            `json:"bot_count"   gorm:"default:0"`
	MaxPlayers int            `json:"max_players" gorm:"default:4"`
	TurnTimer  int            `json:"turn_timer"  gorm:"default:30"` // seconds, 0=off
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-"           gorm:"index"`

	// Relations
	Players []RoomPlayer `json:"players,omitempty" gorm:"foreignKey:RoomID"`
	Host    *User        `json:"host,omitempty"    gorm:"foreignKey:HostUserID"`
}

type RoomPlayer struct {
	ID            uuid.UUID `json:"id"             gorm:"type:uuid;primarykey"`
	RoomID        uuid.UUID `json:"room_id"        gorm:"type:uuid;index"`
	UserID        uuid.UUID `json:"user_id"        gorm:"type:uuid"`
	Seat          int       `json:"seat"`
	IsBot         bool      `json:"is_bot"         gorm:"default:false"`
	BotDifficulty string    `json:"bot_difficulty"` // easy, medium, hard
	JoinedAt      time.Time `json:"joined_at"`

	// Relations
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

type RoomFilter struct {
	ID         *uuid.UUID
	Code       *string
	HostUserID *uuid.UUID
	Status     *RoomStatus
}

func (f *RoomFilter) Query(q *gorm.DB) {
	if f.ID != nil {
		q.Where("id = ?", *f.ID)
	}
	if f.Code != nil {
		q.Where("code = ?", *f.Code)
	}
	if f.HostUserID != nil {
		q.Where("host_user_id = ?", *f.HostUserID)
	}
	if f.Status != nil {
		q.Where("status = ?", *f.Status)
	}
}
