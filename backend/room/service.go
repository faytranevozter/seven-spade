package room

import (
	"app/domain"
	"app/domain/model/auth"
	request_model "app/domain/model/request"
	"app/domain/model/response"
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type RoomRepository interface {
	Create(ctx context.Context, room *domain.Room) error
	FindByCode(ctx context.Context, code string) (*domain.Room, error)
	FindByID(ctx context.Context, id string) (*domain.Room, error)
	Update(ctx context.Context, room *domain.Room) error
	AddPlayer(ctx context.Context, player *domain.RoomPlayer) error
	RemovePlayer(ctx context.Context, roomID, userID string) error
	CountPlayers(ctx context.Context, roomID string) int64
}

type Service struct {
	contextTimeout time.Duration
	roomRepo       RoomRepository
}

func NewService(repo RoomRepository) *Service {
	return &Service{
		contextTimeout: time.Second * 10,
		roomRepo:       repo,
	}
}

func (s *Service) CreateRoom(ctx context.Context, claim auth.JWTClaimUser, payload request_model.CreateRoomRequest) (int, response.Base) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	userID, _ := uuid.Parse(claim.UserID)

	code := generateRoomCode()

	room := domain.Room{
		ID:         uuid.New(),
		Code:       code,
		HostUserID: userID,
		Status:     domain.RoomStatusWaiting,
		BotEnabled: payload.BotEnabled,
		BotCount:   payload.BotCount,
		MaxPlayers: 4,
		TurnTimer:  payload.TurnTimer,
	}

	if room.TurnTimer == 0 {
		room.TurnTimer = 30
	}

	err := s.roomRepo.Create(ctx, &room)
	if err != nil {
		return http.StatusInternalServerError, response.Error(domain.ErrInternalServerCode, err.Error())
	}

	// Add host as first player
	player := domain.RoomPlayer{
		ID:       uuid.New(),
		RoomID:   room.ID,
		UserID:   userID,
		Seat:     0,
		IsBot:    false,
		JoinedAt: time.Now(),
	}
	if err := s.roomRepo.AddPlayer(ctx, &player); err != nil {
		return http.StatusInternalServerError, response.Error(domain.ErrInternalServerCode, err.Error())
	}

	// Reload with relations
	created, _ := s.roomRepo.FindByCode(ctx, code)
	if created != nil {
		return http.StatusOK, response.Success(created)
	}

	return http.StatusOK, response.Success(room)
}

func (s *Service) GetRoom(ctx context.Context, code string) (int, response.Base) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	room, err := s.roomRepo.FindByCode(ctx, code)
	if err != nil {
		return http.StatusNotFound, response.Error(domain.ErrNotFoundCode, "room not found")
	}

	return http.StatusOK, response.Success(room)
}

func (s *Service) JoinRoom(ctx context.Context, claim auth.JWTClaimUser, code string) (int, response.Base) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	userID, _ := uuid.Parse(claim.UserID)

	room, err := s.roomRepo.FindByCode(ctx, code)
	if err != nil {
		return http.StatusNotFound, response.Error(domain.ErrNotFoundCode, "room not found")
	}

	if room.Status != domain.RoomStatusWaiting {
		return http.StatusBadRequest, response.Error(domain.ErrBadRequestCode, "game already in progress")
	}

	// Check if already in room
	for _, p := range room.Players {
		if p.UserID == userID {
			return http.StatusOK, response.Success(room) // already joined
		}
	}

	humanCount := 0
	for _, p := range room.Players {
		if !p.IsBot {
			humanCount++
		}
	}

	maxHumans := room.MaxPlayers
	if room.BotEnabled {
		maxHumans = room.MaxPlayers - room.BotCount
	}

	if humanCount >= maxHumans {
		return http.StatusBadRequest, response.Error(domain.ErrBadRequestCode, "room is full")
	}

	// Find next available seat
	seatTaken := make(map[int]bool)
	for _, p := range room.Players {
		seatTaken[p.Seat] = true
	}
	seat := -1
	for i := 0; i < room.MaxPlayers; i++ {
		if !seatTaken[i] {
			seat = i
			break
		}
	}
	if seat == -1 {
		return http.StatusBadRequest, response.Error(domain.ErrBadRequestCode, "no seats available")
	}

	player := domain.RoomPlayer{
		ID:       uuid.New(),
		RoomID:   room.ID,
		UserID:   userID,
		Seat:     seat,
		IsBot:    false,
		JoinedAt: time.Now(),
	}

	if err := s.roomRepo.AddPlayer(ctx, &player); err != nil {
		return http.StatusInternalServerError, response.Error(domain.ErrInternalServerCode, err.Error())
	}

	updated, _ := s.roomRepo.FindByCode(ctx, code)
	return http.StatusOK, response.Success(updated)
}

func (s *Service) KickPlayer(ctx context.Context, claim auth.JWTClaimUser, code string, targetUserID string) (int, response.Base) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	hostID, _ := uuid.Parse(claim.UserID)

	room, err := s.roomRepo.FindByCode(ctx, code)
	if err != nil {
		return http.StatusNotFound, response.Error(domain.ErrNotFoundCode, "room not found")
	}

	if room.HostUserID != hostID {
		return http.StatusForbidden, response.Error(domain.ErrBadRequestCode, "only the host can kick players")
	}

	if targetUserID == claim.UserID {
		return http.StatusBadRequest, response.Error(domain.ErrBadRequestCode, "cannot kick yourself")
	}

	if err := s.roomRepo.RemovePlayer(ctx, room.ID.String(), targetUserID); err != nil {
		return http.StatusInternalServerError, response.Error(domain.ErrInternalServerCode, err.Error())
	}

	updated, _ := s.roomRepo.FindByCode(ctx, code)
	return http.StatusOK, response.Success(updated)
}

func (s *Service) UpdateSettings(ctx context.Context, claim auth.JWTClaimUser, code string, payload request_model.UpdateRoomSettingsRequest) (int, response.Base) {
	ctx, cancel := context.WithTimeout(ctx, s.contextTimeout)
	defer cancel()

	hostID, _ := uuid.Parse(claim.UserID)

	room, err := s.roomRepo.FindByCode(ctx, code)
	if err != nil {
		return http.StatusNotFound, response.Error(domain.ErrNotFoundCode, "room not found")
	}

	if room.HostUserID != hostID {
		return http.StatusForbidden, response.Error(domain.ErrBadRequestCode, "only the host can update settings")
	}

	if payload.BotEnabled != nil {
		room.BotEnabled = *payload.BotEnabled
	}
	if payload.BotCount != nil {
		if *payload.BotCount < 0 || *payload.BotCount > 3 {
			return http.StatusBadRequest, response.Error(domain.ErrBadRequestCode, "bot count must be 0-3")
		}
		room.BotCount = *payload.BotCount
	}
	if payload.TurnTimer != nil {
		room.TurnTimer = *payload.TurnTimer
	}

	if err := s.roomRepo.Update(ctx, room); err != nil {
		return http.StatusInternalServerError, response.Error(domain.ErrInternalServerCode, err.Error())
	}

	updated, _ := s.roomRepo.FindByCode(ctx, code)
	return http.StatusOK, response.Success(updated)
}

func generateRoomCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no I, O, 0, 1 to avoid confusion
	code := make([]byte, 6)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		code[i] = chars[n.Int64()]
	}
	return fmt.Sprintf("%s", code)
}
