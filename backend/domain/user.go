package domain

import (
	gorm_model "app/domain/model/gorm"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID            uuid.UUID      `json:"id"             gorm:"type:uuid;primarykey"`
	DisplayName   string         `json:"display_name"`
	Email         string         `json:"email"          gorm:"uniqueIndex"`
	Password      string         `json:"-"`
	AvatarURL     string         `json:"avatar_url"`
	OAuthProvider string         `json:"oauth_provider" gorm:"column:oauth_provider"` // google, github, telegram
	OAuthID       string         `json:"-"              gorm:"index;column:oauth_id"`
	EloRating     int            `json:"elo_rating"     gorm:"default:1000"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-"              gorm:"index"`
}

var UserAllowedSort = []string{"display_name", "email", "elo_rating", "created_at", "updated_at"}

type UserFilter struct {
	gorm_model.DefaultFilter
	ID            *uuid.UUID
	Email         *string
	OAuthProvider *string
	OAuthID       *string
}

func (f *UserFilter) Query(q *gorm.DB) {
	f.DefaultFilter.Query(q)

	if f.ID != nil {
		q.Where("id = ?", *f.ID)
	}
	if f.Email != nil {
		q.Where("email = ?", *f.Email)
	}
	if f.OAuthProvider != nil {
		q.Where("oauth_provider = ?", *f.OAuthProvider)
	}
	if f.OAuthID != nil {
		q.Where("oauth_id = ?", *f.OAuthID)
	}
}
