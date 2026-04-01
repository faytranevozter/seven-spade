package main

import (
	"app/domain"
	auth "app/domain/model/auth"
	"app/game"
	"app/helpers/connection"
	gormrepo "app/internal/repository/gorm"
	"app/internal/rest"
	"app/internal/rest/middleware"
	"app/internal/ws"
	"app/matchmaking"
	"app/room"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func init() {
	_ = godotenv.Load()
}

func main() {
	// default log level
	logLevel := logger.Info

	if os.Getenv("GO_ENV") == "production" || os.Getenv("GO_ENV") == "prod" {
		gin.SetMode(gin.ReleaseMode)
		logLevel = logger.Silent
	}

	timeoutStr := os.Getenv("TIMEOUT")
	if timeoutStr == "" {
		timeoutStr = "5"
	}
	timeout, _ := strconv.Atoi(timeoutStr)
	timeoutContext := time.Duration(timeout) * time.Second

	// logger
	writers := make([]io.Writer, 0)
	if logSTDOUT, _ := strconv.ParseBool(os.Getenv("LOG_TO_STDOUT")); logSTDOUT {
		writers = append(writers, os.Stdout)
	}

	if logFILE, _ := strconv.ParseBool(os.Getenv("LOG_TO_FILE")); logFILE {
		logMaxSize, _ := strconv.Atoi(os.Getenv("LOG_MAX_SIZE"))
		if logMaxSize == 0 {
			logMaxSize = 50
		}

		logFilename := os.Getenv("LOG_FILENAME")
		if logFilename == "" {
			logFilename = "server.log"
		}

		lg := &lumberjack.Logger{
			Filename:   logFilename,
			MaxSize:    logMaxSize,
			MaxBackups: 1,
			LocalTime:  true,
		}

		writers = append(writers, lg)
	}

	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(io.MultiWriter(writers...))
	gin.DefaultWriter = logrus.StandardLogger().Writer()

	// init redis database
	var redisClient *redis.Client
	if useRedis, err := strconv.ParseBool(os.Getenv("USE_REDIS")); err == nil && useRedis {
		redisClient = connection.NewRedis(timeoutContext, os.Getenv("REDIS_URL"))
	}

	// Initialize GORM with PostgreSQL
	gormDB, err := gorm.Open(
		connection.NewPostgresGORM(timeoutContext, os.Getenv("DB_URL")),
		&gorm.Config{
			Logger: logger.New(
				logrus.New(),
				logger.Config{
					SlowThreshold:             200 * time.Millisecond,
					LogLevel:                  logLevel,
					IgnoreRecordNotFoundError: true,
					Colorful:                  false,
				},
			),
		},
	)
	if err != nil {
		panic("error connecting to database: " + err.Error())
	}

	// --- Init Repositories ---
	userRepo := gormrepo.NewUserRepository(gormDB)
	roomRepo := gormrepo.NewRoomRepository(gormDB)
	gameRepo := gormrepo.NewGameRepository(gormDB)

	// --- Init Services ---
	roomService := room.NewService(roomRepo)
	matchmakingService := matchmaking.NewService()

	// --- Init WebSocket Hub ---
	hub := ws.NewHub()
	go hub.Run()

	// --- Init Game Manager ---
	gameManager := ws.NewGameManager(hub)
	hub.GameManager = gameManager

	// Handle game completion: persist results and update ELO
	gameManager.OnGameComplete = func(session *ws.GameSession) {
		ctx := context.Background()
		persistGameResults(ctx, gameRepo, session)
	}

	// Wire up WebSocket message routing
	hub.OnMessage = func(client *ws.Client, msg ws.WSMessage) {
		handleWSMessage(client, msg, hub, gameManager, roomRepo, gameRepo)
	}

	// Start matchmaking background worker
	matchmakingCtx, matchmakingCancel := context.WithCancel(context.Background())
	defer matchmakingCancel()
	go matchmakingService.StartMatchmaking(matchmakingCtx)

	matchmakingService.OnMatchFound = func(match matchmaking.MatchResult) {
		logrus.Infof("Match found via matchmaking, creating room...")
		// Auto-create room — simplified, in production would use room service
	}

	// --- Init Middleware ---
	mdl := middleware.NewMiddleware(redisClient)

	// --- Init Gin ---
	ginEngine := gin.New()
	ginEngine.Use(mdl.Logger(io.MultiWriter(writers...)))
	ginEngine.Use(mdl.Cors())

	// Health check
	ginEngine.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, map[string]any{
			"message": "Seven Spade API",
			"version": "1.0.0",
		})
	})

	// --- Register Routes ---
	apiGroup := ginEngine.Group("/api")

	// Auth (OAuth)
	rest.NewAuthHandler(ginEngine.Group(""), mdl, func(ctx context.Context, info rest.OAuthUserInfo) (uuid.UUID, string, error) {
		return upsertOAuthUser(ctx, userRepo, info)
	})

	// Rooms
	rest.NewRoomHandler(apiGroup, roomService, mdl)

	// Games (history, leaderboard)
	rest.NewGameHandler(apiGroup, gameRepo, mdl)

	// Matchmaking
	rest.NewMatchmakingHandler(apiGroup, matchmakingService, mdl)

	// WebSocket
	rest.NewWSHandler(ginEngine.Group(""), hub, mdl)

	// User profile endpoints
	apiGroup.GET("/users/me", mdl.Auth(), func(c *gin.Context) {
		claim := c.MustGet("token_data").(auth.JWTClaimUser)
		ctx := c.Request.Context()
		uid, err := uuid.Parse(claim.UserID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}
		user, err := userRepo.FetchOneUser(ctx, domain.UserFilter{ID: &uid})
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		stats, _ := gameRepo.GetOrCreateStats(ctx, user.ID.String())
		c.JSON(http.StatusOK, gin.H{
			"id":           user.ID,
			"display_name": user.DisplayName,
			"email":        user.Email,
			"avatar_url":   user.AvatarURL,
			"elo_rating":   user.EloRating,
			"stats":        stats,
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "5050"
	}

	logrus.Infof("Seven Spade server running on port %s", port)
	ginEngine.Run(":" + port)
}

// upsertOAuthUser finds or creates a user based on OAuth info
func upsertOAuthUser(ctx context.Context, repo *gormrepo.UserRepository, info rest.OAuthUserInfo) (uuid.UUID, string, error) {
	provider := info.Provider
	providerID := info.ProviderID

	// Try to find existing user
	user, err := repo.FetchOneUser(ctx, domain.UserFilter{
		OAuthProvider: &provider,
		OAuthID:       &providerID,
	})
	if err == nil && user != nil {
		// Update avatar if changed
		if user.AvatarURL != info.AvatarURL {
			user.AvatarURL = info.AvatarURL
			user.UpdatedAt = time.Now()
			repo.UpdateUser(ctx, user)
		}
		return user.ID, user.DisplayName, nil
	}

	// Create new user
	newUser := &domain.User{
		ID:            uuid.New(),
		DisplayName:   info.DisplayName,
		Email:         info.Email,
		AvatarURL:     info.AvatarURL,
		OAuthProvider: info.Provider,
		OAuthID:       info.ProviderID,
		EloRating:     1000,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := repo.CreateUser(ctx, newUser); err != nil {
		return uuid.Nil, "", err
	}

	return newUser.ID, newUser.DisplayName, nil
}

// handleWSMessage routes incoming WebSocket messages to the appropriate handler
func handleWSMessage(client *ws.Client, msg ws.WSMessage, hub *ws.Hub, gm *ws.GameManager, roomRepo *gormrepo.RoomRepository, gameRepo *gormrepo.GameRepository) {
	switch msg.Type {
	case ws.MsgTypePlayCard:
		payload, err := ws.ParsePayload[ws.PayloadPlayCard](msg)
		if err != nil {
			client.SendMessage(ws.WSMessage{Type: ws.MsgTypeError, Payload: ws.PayloadError{Message: "invalid payload"}})
			return
		}
		gm.HandlePlayCard(client, payload)

	case ws.MsgTypeFaceDown:
		payload, err := ws.ParsePayload[ws.PayloadFaceDown](msg)
		if err != nil {
			client.SendMessage(ws.WSMessage{Type: ws.MsgTypeError, Payload: ws.PayloadError{Message: "invalid payload"}})
			return
		}
		gm.HandleFaceDown(client, payload)

	case ws.MsgTypeRequestState:
		gm.HandleRequestState(client)

	case ws.MsgTypeStartGame:
		handleStartGame(client, hub, gm, roomRepo, gameRepo)

	case ws.MsgTypeRoomChat:
		payload, err := ws.ParsePayload[ws.PayloadChat](msg)
		if err != nil {
			return
		}
		hub.BroadcastToRoom(client.RoomCode, ws.WSMessage{
			Type: ws.MsgTypeRoomChatBcast,
			Payload: map[string]string{
				"user_id":      client.UserID.String(),
				"display_name": client.DisplayName,
				"message":      payload.Message,
			},
		})

	default:
		client.SendMessage(ws.WSMessage{Type: ws.MsgTypeError, Payload: ws.PayloadError{Message: "unknown message type"}})
	}
}

// handleStartGame is called when the host triggers game start via WebSocket
func handleStartGame(client *ws.Client, hub *ws.Hub, gm *ws.GameManager, roomRepo *gormrepo.RoomRepository, gameRepo *gormrepo.GameRepository) {
	ctx := context.Background()

	roomData, err := roomRepo.FindByCode(ctx, client.RoomCode)
	if err != nil {
		client.SendMessage(ws.WSMessage{Type: ws.MsgTypeError, Payload: ws.PayloadError{Message: "room not found"}})
		return
	}

	// Only host can start
	if roomData.HostUserID != client.UserID {
		client.SendMessage(ws.WSMessage{Type: ws.MsgTypeError, Payload: ws.PayloadError{Message: "only the host can start the game"}})
		return
	}

	if roomData.Status != domain.RoomStatusWaiting {
		client.SendMessage(ws.WSMessage{Type: ws.MsgTypeError, Payload: ws.PayloadError{Message: "game already started"}})
		return
	}

	// Build session players from room players
	sessionPlayers := make([]ws.SessionPlayer, 0, 4)
	for _, p := range roomData.Players {
		sp := ws.SessionPlayer{
			UserID:      p.UserID,
			DisplayName: "Player",
			Seat:        p.Seat,
			IsBot:       p.IsBot,
		}
		if p.User != nil {
			sp.DisplayName = p.User.DisplayName
		}
		if p.IsBot {
			switch p.BotDifficulty {
			case "hard":
				sp.BotDifficulty = game.BotHard
			case "easy":
				sp.BotDifficulty = game.BotEasy
			default:
				sp.BotDifficulty = game.BotMedium
			}
		}
		sessionPlayers = append(sessionPlayers, sp)
	}

	// Add bots if enabled and seats are empty
	if roomData.BotEnabled && roomData.BotCount > 0 {
		existingSeats := make(map[int]bool)
		for _, p := range sessionPlayers {
			existingSeats[p.Seat] = true
		}

		botsAdded := 0
		for seat := 0; seat < 4 && botsAdded < roomData.BotCount; seat++ {
			if !existingSeats[seat] {
				sessionPlayers = append(sessionPlayers, ws.SessionPlayer{
					UserID:        uuid.New(),
					DisplayName:   botName(botsAdded),
					Seat:          seat,
					IsBot:         true,
					BotDifficulty: game.BotMedium,
				})
				botsAdded++
			}
		}
	}

	if len(sessionPlayers) != 4 {
		client.SendMessage(ws.WSMessage{Type: ws.MsgTypeError, Payload: ws.PayloadError{Message: "need exactly 4 players to start"}})
		return
	}

	// Sort by seat
	sortPlayersBySeat(sessionPlayers)

	// Assign seats to connected clients
	for _, c := range hub.GetRoomClients(client.RoomCode) {
		for _, p := range sessionPlayers {
			if c.UserID == p.UserID {
				c.Seat = p.Seat
			}
		}
	}

	// Create game record in DB
	gameID := uuid.New()
	turnTimeout := time.Duration(roomData.TurnTimer) * time.Second

	dbGame := &domain.Game{
		ID:           gameID,
		RoomID:       roomData.ID,
		AceDirection: "undecided",
		Status:       domain.GameDBStatusInProgress,
		StartedAt:    time.Now(),
	}
	gameRepo.CreateGame(ctx, dbGame)

	// Create game players
	gamePlayers := make([]domain.GamePlayer, 0, 4)
	for _, p := range sessionPlayers {
		gamePlayers = append(gamePlayers, domain.GamePlayer{
			ID:            uuid.New(),
			GameID:        gameID,
			UserID:        p.UserID,
			Seat:          p.Seat,
			IsBot:         p.IsBot,
			BotDifficulty: p.BotDifficulty.String(),
		})
	}
	gameRepo.CreateGamePlayers(ctx, gamePlayers)

	// Update room status
	roomData.Status = domain.RoomStatusPlaying
	roomRepo.Update(ctx, roomData)

	// Broadcast game starting
	hub.BroadcastToRoom(client.RoomCode, ws.WSMessage{
		Type:    ws.MsgTypeGameStarting,
		Payload: ws.PayloadGameStarting{GameID: gameID.String()},
	})

	// Start the game session
	if err := gm.StartGame(client.RoomCode, gameID, sessionPlayers, turnTimeout); err != nil {
		client.SendMessage(ws.WSMessage{Type: ws.MsgTypeError, Payload: ws.PayloadError{Message: err.Error()}})
	}
}

// persistGameResults saves game results to the database
func persistGameResults(ctx context.Context, gameRepo *gormrepo.GameRepository, session *ws.GameSession) {
	now := time.Now()

	// Update game record
	dbGame, err := gameRepo.FindGameByID(ctx, session.GameID.String())
	if err != nil {
		logrus.Errorf("Failed to find game for persistence: %v", err)
		return
	}

	dbGame.Status = domain.GameDBStatusFinished
	dbGame.EndedAt = &now
	dbGame.AceDirection = session.State.AceDirection.String()
	gameRepo.UpdateGame(ctx, dbGame)

	// Update game players with results
	if session.State.Results != nil {
		gamePlayers := make([]domain.GamePlayer, 0)
		for _, r := range session.State.Results {
			for _, gp := range dbGame.Players {
				if gp.Seat == r.Seat {
					gp.PenaltyPoints = r.PenaltyPoints
					gp.FinalRank = r.Rank
					gamePlayers = append(gamePlayers, gp)
					break
				}
			}
		}
		gameRepo.UpdateGamePlayers(ctx, gamePlayers)
	}

	// Persist moves
	moves := make([]domain.GameMove, 0, len(session.State.MoveHistory))
	for _, m := range session.State.MoveHistory {
		moves = append(moves, domain.GameMove{
			ID:        uuid.New(),
			GameID:    session.GameID,
			Seat:      m.Seat,
			MoveNum:   m.MoveNum,
			MoveType:  m.Type.String(),
			CardSuit:  m.Card.Suit.String(),
			CardRank:  int(m.Card.Rank),
			CreatedAt: now,
		})
	}
	gameRepo.CreateMoves(ctx, moves)

	// Update ELO for human players
	ratings := make([]int, 0)
	ranks := make([]int, 0)
	humanIndices := make([]int, 0)

	for _, p := range session.Players {
		if !p.IsBot {
			for _, r := range session.State.Results {
				if r.Seat == p.Seat {
					stats, _ := gameRepo.GetOrCreateStats(ctx, p.UserID.String())
					if stats != nil {
						ratings = append(ratings, stats.EloRating)
						ranks = append(ranks, r.Rank)
						humanIndices = append(humanIndices, len(ratings)-1)
					}
					break
				}
			}
		}
	}

	if len(ratings) > 1 {
		newRatings := matchmaking.UpdateELO(ratings, ranks)
		idx := 0
		for _, p := range session.Players {
			if !p.IsBot {
				stats, _ := gameRepo.GetOrCreateStats(ctx, p.UserID.String())
				if stats != nil && idx < len(newRatings) {
					stats.EloRating = newRatings[idx]
					stats.TotalGames++
					for _, r := range session.State.Results {
						if r.Seat == p.Seat {
							stats.TotalPenaltyPoints += r.PenaltyPoints
							if r.Rank == 1 {
								stats.Wins++
							} else {
								stats.Losses++
							}
							break
						}
					}
					gameRepo.UpdateStats(ctx, stats)
					idx++
				}
			}
		}
	}

	logrus.Infof("Game %s results persisted", session.GameID)
}

func botName(index int) string {
	names := []string{"Bot Alpha", "Bot Beta", "Bot Gamma"}
	if index < len(names) {
		return names[index]
	}
	return "Bot"
}

func sortPlayersBySeat(players []ws.SessionPlayer) {
	for i := 0; i < len(players); i++ {
		for j := i + 1; j < len(players); j++ {
			if players[j].Seat < players[i].Seat {
				players[i], players[j] = players[j], players[i]
			}
		}
	}
}

// ensure json import is used
var _ = json.Marshal
