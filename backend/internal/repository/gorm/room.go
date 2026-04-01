package gormrepo

import (
	"app/domain"
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type RoomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) *RoomRepository {
	db.AutoMigrate(&domain.Room{}, &domain.RoomPlayer{})
	return &RoomRepository{db: db}
}

func (r *RoomRepository) Create(ctx context.Context, room *domain.Room) error {
	room.CreatedAt = time.Now()
	room.UpdatedAt = time.Now()
	err := r.db.Create(room).Error
	if err != nil {
		logrus.Error("RoomRepository.Create:", err)
	}
	return err
}

func (r *RoomRepository) FindByCode(ctx context.Context, code string) (*domain.Room, error) {
	var room domain.Room
	err := r.db.Preload("Players").Preload("Players.User").Preload("Host").
		Where("code = ?", code).First(&room).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *RoomRepository) FindByID(ctx context.Context, id string) (*domain.Room, error) {
	var room domain.Room
	err := r.db.Preload("Players").Preload("Players.User").Preload("Host").
		Where("id = ?", id).First(&room).Error
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *RoomRepository) Update(ctx context.Context, room *domain.Room) error {
	room.UpdatedAt = time.Now()
	return r.db.Save(room).Error
}

func (r *RoomRepository) AddPlayer(ctx context.Context, player *domain.RoomPlayer) error {
	player.JoinedAt = time.Now()
	return r.db.Create(player).Error
}

func (r *RoomRepository) RemovePlayer(ctx context.Context, roomID, userID string) error {
	return r.db.Where("room_id = ? AND user_id = ?", roomID, userID).Delete(&domain.RoomPlayer{}).Error
}

func (r *RoomRepository) CountPlayers(ctx context.Context, roomID string) int64 {
	var count int64
	r.db.Model(&domain.RoomPlayer{}).Where("room_id = ?", roomID).Count(&count)
	return count
}

func (r *RoomRepository) FindWaitingRooms(ctx context.Context) ([]domain.Room, error) {
	var rooms []domain.Room
	err := r.db.Preload("Players").
		Where("status = ?", domain.RoomStatusWaiting).
		Find(&rooms).Error
	return rooms, err
}
