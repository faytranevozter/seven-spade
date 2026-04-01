package gormrepo

import (
	"app/domain"
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type GameRepository struct {
	db *gorm.DB
}

func NewGameRepository(db *gorm.DB) *GameRepository {
	db.AutoMigrate(&domain.Game{}, &domain.GamePlayer{}, &domain.GameMove{}, &domain.PlayerStats{})
	return &GameRepository{db: db}
}

func (r *GameRepository) CreateGame(ctx context.Context, game *domain.Game) error {
	game.CreatedAt = time.Now()
	game.UpdatedAt = time.Now()
	err := r.db.Create(game).Error
	if err != nil {
		logrus.Error("GameRepository.CreateGame:", err)
	}
	return err
}

func (r *GameRepository) CreateGamePlayer(ctx context.Context, player *domain.GamePlayer) error {
	return r.db.Create(player).Error
}

func (r *GameRepository) CreateGamePlayers(ctx context.Context, players []domain.GamePlayer) error {
	return r.db.Create(&players).Error
}

func (r *GameRepository) UpdateGame(ctx context.Context, game *domain.Game) error {
	game.UpdatedAt = time.Now()
	return r.db.Save(game).Error
}

func (r *GameRepository) UpdateGamePlayers(ctx context.Context, players []domain.GamePlayer) error {
	tx := r.db.Begin()
	for _, p := range players {
		if err := tx.Save(&p).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (r *GameRepository) CreateMoves(ctx context.Context, moves []domain.GameMove) error {
	if len(moves) == 0 {
		return nil
	}
	return r.db.Create(&moves).Error
}

func (r *GameRepository) FindGameByID(ctx context.Context, id string) (*domain.Game, error) {
	var game domain.Game
	err := r.db.Preload("Players").Preload("Players.User").
		Where("id = ?", id).First(&game).Error
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func (r *GameRepository) FindGamesByUserID(ctx context.Context, userID string, limit, offset int) ([]domain.Game, int64, error) {
	var games []domain.Game
	var total int64

	query := r.db.Model(&domain.Game{}).
		Joins("JOIN game_players ON game_players.game_id = games.id").
		Where("game_players.user_id = ?", userID)

	query.Count(&total)

	err := query.Preload("Players").Preload("Players.User").
		Order("games.created_at DESC").
		Limit(limit).Offset(offset).
		Find(&games).Error

	return games, total, err
}

// --- Player Stats ---

func (r *GameRepository) GetOrCreateStats(ctx context.Context, userID string) (*domain.PlayerStats, error) {
	var stats domain.PlayerStats
	err := r.db.Where("user_id = ?", userID).First(&stats).Error
	if err == gorm.ErrRecordNotFound {
		stats = domain.PlayerStats{
			EloRating: 1000,
			UpdatedAt: time.Now(),
		}
		// Parse UUID
		if err := r.db.Exec("INSERT INTO player_stats (user_id, elo_rating, updated_at) VALUES (?, 1000, ?) ON CONFLICT (user_id) DO NOTHING", userID, time.Now()).Error; err != nil {
			logrus.Error("GetOrCreateStats:", err)
		}
		r.db.Where("user_id = ?", userID).First(&stats)
		return &stats, nil
	}
	return &stats, err
}

func (r *GameRepository) UpdateStats(ctx context.Context, stats *domain.PlayerStats) error {
	stats.UpdatedAt = time.Now()
	return r.db.Save(stats).Error
}

func (r *GameRepository) GetLeaderboard(ctx context.Context, limit, offset int) ([]domain.PlayerStats, int64, error) {
	var stats []domain.PlayerStats
	var total int64

	r.db.Model(&domain.PlayerStats{}).Count(&total)

	err := r.db.Preload("User").
		Order("elo_rating DESC").
		Limit(limit).Offset(offset).
		Find(&stats).Error

	return stats, total, err
}
