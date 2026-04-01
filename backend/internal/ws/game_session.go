package ws

import (
	"app/game"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// GameSession holds the in-memory state for an active game
type GameSession struct {
	GameID    uuid.UUID
	RoomCode  string
	State     *game.GameState
	Players   []SessionPlayer
	mu        sync.Mutex
	turnTimer *time.Timer
	turnTimeout time.Duration
}

// SessionPlayer represents a player in the game session
type SessionPlayer struct {
	UserID      uuid.UUID
	DisplayName string
	Seat        int
	IsBot       bool
	BotDifficulty game.BotDifficulty
}

// GameManager manages all active game sessions
type GameManager struct {
	sessions map[string]*GameSession // keyed by room code
	mu       sync.RWMutex
	hub      *Hub
	rng      *rand.Rand

	// Callback for persisting game results
	OnGameComplete func(session *GameSession)
}

// NewGameManager creates a new game manager
func NewGameManager(hub *Hub) *GameManager {
	return &GameManager{
		sessions: make(map[string]*GameSession),
		hub:      hub,
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// StartGame creates a new game session for a room
func (gm *GameManager) StartGame(roomCode string, gameID uuid.UUID, players []SessionPlayer, turnTimeout time.Duration) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if _, exists := gm.sessions[roomCode]; exists {
		return fmt.Errorf("game already in progress for room %s", roomCode)
	}

	// Create and shuffle deck
	deck := game.NewDeck()
	deck.Shuffle(gm.rng)
	hands := deck.Deal(len(players))

	// Create game state
	state := game.NewGameState(hands)

	session := &GameSession{
		GameID:      gameID,
		RoomCode:    roomCode,
		State:       state,
		Players:     players,
		turnTimeout: turnTimeout,
	}

	gm.sessions[roomCode] = session

	// Send initial game state to each player
	gm.broadcastGameState(session)

	// Notify whose turn it is
	gm.notifyTurn(session)

	// If the first player is a bot, execute their move
	if players[state.CurrentTurn].IsBot {
		go gm.executeBotTurn(session)
	}

	return nil
}

// HandlePlayCard processes a play_card message from a client
func (gm *GameManager) HandlePlayCard(client *Client, payload PayloadPlayCard) {
	gm.mu.RLock()
	session, exists := gm.sessions[client.RoomCode]
	gm.mu.RUnlock()

	if !exists {
		client.SendMessage(WSMessage{Type: MsgTypeError, Payload: PayloadError{Message: "no active game"}})
		return
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	if session.State.CurrentTurn != client.Seat {
		client.SendMessage(WSMessage{Type: MsgTypeInvalidMove, Payload: PayloadInvalidMove{Reason: "not your turn"}})
		return
	}

	// Parse suit
	suit, err := parseSuit(payload.Suit)
	if err != nil {
		client.SendMessage(WSMessage{Type: MsgTypeInvalidMove, Payload: PayloadInvalidMove{Reason: err.Error()}})
		return
	}

	card := game.Card{Suit: suit, Rank: game.Rank(payload.Rank)}
	move := game.Move{Seat: client.Seat, Type: game.MovePlayCard, Card: card}

	prevAceDir := session.State.AceDirection
	newState, err := game.ApplyMove(session.State, move)
	if err != nil {
		client.SendMessage(WSMessage{Type: MsgTypeInvalidMove, Payload: PayloadInvalidMove{Reason: err.Error()}})
		return
	}

	session.State = newState
	gm.cancelTurnTimer(session)

	// Broadcast the move
	gm.broadcastMove(session, move, true)

	// Check if ace got locked
	if prevAceDir == game.AceUndecided && newState.AceDirection != game.AceUndecided {
		gm.hub.BroadcastToRoom(session.RoomCode, WSMessage{
			Type:    MsgTypeAceLocked,
			Payload: PayloadAceLocked{Direction: newState.AceDirection.String()},
		})
	}

	// Check suit closures
	for i, seq := range newState.Sequences {
		if seq.Closed && card.Suit == game.Suit(i) {
			gm.hub.BroadcastToRoom(session.RoomCode, WSMessage{
				Type:    MsgTypeSuitClosed,
				Payload: PayloadSuitClosed{Suit: seq.Suit.String()},
			})
		}
	}

	// Check game over
	if newState.Status == game.StatusFinished {
		gm.handleGameOver(session)
		return
	}

	// Send updated state and notify next turn
	gm.broadcastGameState(session)
	gm.notifyTurn(session)

	// If next player is a bot, execute their turn
	if session.Players[newState.CurrentTurn].IsBot {
		go gm.executeBotTurn(session)
	}
}

// HandleFaceDown processes a face_down message from a client
func (gm *GameManager) HandleFaceDown(client *Client, payload PayloadFaceDown) {
	gm.mu.RLock()
	session, exists := gm.sessions[client.RoomCode]
	gm.mu.RUnlock()

	if !exists {
		client.SendMessage(WSMessage{Type: MsgTypeError, Payload: PayloadError{Message: "no active game"}})
		return
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	if session.State.CurrentTurn != client.Seat {
		client.SendMessage(WSMessage{Type: MsgTypeInvalidMove, Payload: PayloadInvalidMove{Reason: "not your turn"}})
		return
	}

	suit, err := parseSuit(payload.Suit)
	if err != nil {
		client.SendMessage(WSMessage{Type: MsgTypeInvalidMove, Payload: PayloadInvalidMove{Reason: err.Error()}})
		return
	}

	card := game.Card{Suit: suit, Rank: game.Rank(payload.Rank)}
	move := game.Move{Seat: client.Seat, Type: game.MoveFaceDown, Card: card}

	newState, err := game.ApplyMove(session.State, move)
	if err != nil {
		client.SendMessage(WSMessage{Type: MsgTypeInvalidMove, Payload: PayloadInvalidMove{Reason: err.Error()}})
		return
	}

	session.State = newState
	gm.cancelTurnTimer(session)

	// Broadcast the move (card is hidden for face-down)
	gm.broadcastMove(session, move, false)

	if newState.Status == game.StatusFinished {
		gm.handleGameOver(session)
		return
	}

	gm.broadcastGameState(session)
	gm.notifyTurn(session)

	if session.Players[newState.CurrentTurn].IsBot {
		go gm.executeBotTurn(session)
	}
}

// HandleRequestState resends the full game state to a client (reconnection)
func (gm *GameManager) HandleRequestState(client *Client) {
	gm.mu.RLock()
	session, exists := gm.sessions[client.RoomCode]
	gm.mu.RUnlock()

	if !exists {
		client.SendMessage(WSMessage{Type: MsgTypeError, Payload: PayloadError{Message: "no active game"}})
		return
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	gm.sendGameStateToPlayer(session, client.Seat)
}

// HandleDisconnect handles a player disconnecting
func (gm *GameManager) HandleDisconnect(client *Client) {
	gm.mu.RLock()
	session, exists := gm.sessions[client.RoomCode]
	gm.mu.RUnlock()

	if !exists {
		return
	}

	// Broadcast updated player connection status
	gm.broadcastGameState(session)
}

// executeBotTurn runs the AI logic for a bot player after a delay
func (gm *GameManager) executeBotTurn(session *GameSession) {
	// Random delay for realism (1-3 seconds)
	delay := time.Duration(1000+gm.rng.Intn(2000)) * time.Millisecond
	time.Sleep(delay)

	session.mu.Lock()
	defer session.mu.Unlock()

	if session.State.Status != game.StatusInProgress {
		return
	}

	currentSeat := session.State.CurrentTurn
	if currentSeat >= len(session.Players) || !session.Players[currentSeat].IsBot {
		return
	}

	botDiff := session.Players[currentSeat].BotDifficulty
	move := game.BotChooseMove(session.State, currentSeat, botDiff, gm.rng)

	prevAceDir := session.State.AceDirection
	newState, err := game.ApplyMove(session.State, move)
	if err != nil {
		logrus.Errorf("Bot move error: %v", err)
		return
	}

	session.State = newState

	// Broadcast
	showCard := move.Type == game.MovePlayCard
	gm.broadcastMove(session, move, showCard)

	if prevAceDir == game.AceUndecided && newState.AceDirection != game.AceUndecided {
		gm.hub.BroadcastToRoom(session.RoomCode, WSMessage{
			Type:    MsgTypeAceLocked,
			Payload: PayloadAceLocked{Direction: newState.AceDirection.String()},
		})
	}

	if move.Type == game.MovePlayCard {
		for i, seq := range newState.Sequences {
			if seq.Closed && move.Card.Suit == game.Suit(i) {
				gm.hub.BroadcastToRoom(session.RoomCode, WSMessage{
					Type:    MsgTypeSuitClosed,
					Payload: PayloadSuitClosed{Suit: seq.Suit.String()},
				})
			}
		}
	}

	if newState.Status == game.StatusFinished {
		gm.handleGameOver(session)
		return
	}

	gm.broadcastGameState(session)
	gm.notifyTurn(session)

	// Chain bot moves
	if session.Players[newState.CurrentTurn].IsBot {
		go gm.executeBotTurn(session)
	}
}

// broadcastGameState sends personalized game state to each player
func (gm *GameManager) broadcastGameState(session *GameSession) {
	for _, player := range session.Players {
		if !player.IsBot {
			gm.sendGameStateToPlayer(session, player.Seat)
		}
	}
}

// sendGameStateToPlayer sends the game state filtered for a specific player
func (gm *GameManager) sendGameStateToPlayer(session *GameSession, seat int) {
	state := session.State

	// Build valid moves for this player
	var validCards []PayloadCard
	mustFaceDown := false

	if state.CurrentTurn == seat && state.Status == game.StatusInProgress {
		moves := game.ValidMoves(state, seat)
		for _, m := range moves {
			validCards = append(validCards, PayloadCard{
				Suit: m.Card.Suit.String(),
				Rank: int(m.Card.Rank),
			})
		}
		mustFaceDown = game.MustFaceDown(state, seat)
	}

	// Build hand
	hand := make([]PayloadCard, 0)
	if seat >= 0 && seat < state.NumPlayers {
		for _, c := range state.Hands[seat] {
			hand = append(hand, PayloadCard{Suit: c.Suit.String(), Rank: int(c.Rank)})
		}
	}

	// Build sequences
	var seqs [4]PayloadSequence
	for i, s := range state.Sequences {
		seqs[i] = PayloadSequence{
			Suit:    s.Suit.String(),
			LowEnd:  int(s.LowEnd),
			HighEnd: int(s.HighEnd),
			Started: s.Started,
			Closed:  s.Closed,
		}
	}

	// Build face-down counts
	fdCounts := make([]int, state.NumPlayers)
	for i := range state.NumPlayers {
		fdCounts[i] = len(state.FaceDown[i])
	}

	// Build players info
	players := make([]PayloadPlayerInfo, 0, len(session.Players))
	for _, p := range session.Players {
		players = append(players, PayloadPlayerInfo{
			UserID:      p.UserID.String(),
			DisplayName: p.DisplayName,
			Seat:        p.Seat,
			IsBot:       p.IsBot,
			IsConnected: p.IsBot || gm.hub.IsConnected(session.RoomCode, p.UserID.String()),
		})
	}

	payload := PayloadGameState{
		GameID:         session.GameID.String(),
		YourSeat:       seat,
		YourHand:       hand,
		HandCounts:     state.HandCounts,
		Sequences:      seqs,
		CurrentTurn:    state.CurrentTurn,
		AceDirection:   state.AceDirection.String(),
		ValidMoves:     validCards,
		MustFaceDown:   mustFaceDown,
		FaceDownCounts: fdCounts,
		Players:        players,
		Status:         state.Status.String(),
	}

	gm.hub.SendToSeat(session.RoomCode, seat, WSMessage{
		Type:    MsgTypeGameState,
		Payload: payload,
	})
}

// notifyTurn notifies the current player it's their turn
func (gm *GameManager) notifyTurn(session *GameSession) {
	seat := session.State.CurrentTurn
	if !session.Players[seat].IsBot {
		gm.hub.SendToSeat(session.RoomCode, seat, WSMessage{
			Type: MsgTypeYourTurn,
		})

		// Start turn timer if configured
		if session.turnTimeout > 0 {
			gm.startTurnTimer(session)
		}
	}
}

// broadcastMove sends move info to all players
func (gm *GameManager) broadcastMove(session *GameSession, move game.Move, showCard bool) {
	payload := PayloadMoveMade{
		Seat:     move.Seat,
		MoveType: move.Type.String(),
		MoveNum:  move.MoveNum,
	}
	if showCard {
		payload.Card = &PayloadCard{
			Suit: move.Card.Suit.String(),
			Rank: int(move.Card.Rank),
		}
	}

	gm.hub.BroadcastToRoom(session.RoomCode, WSMessage{
		Type:    MsgTypeMoveMade,
		Payload: payload,
	})
}

// handleGameOver broadcasts results and cleans up
func (gm *GameManager) handleGameOver(session *GameSession) {
	results := session.State.Results

	payloadResults := make([]PayloadPlayerResult, 0, len(results))
	for _, r := range results {
		player := session.Players[r.Seat]
		fdCards := make([]PayloadCard, 0, len(r.FaceDownCards))
		for _, c := range r.FaceDownCards {
			fdCards = append(fdCards, PayloadCard{Suit: c.Suit.String(), Rank: int(c.Rank)})
		}

		payloadResults = append(payloadResults, PayloadPlayerResult{
			UserID:        player.UserID.String(),
			DisplayName:   player.DisplayName,
			Seat:          r.Seat,
			PenaltyPoints: r.PenaltyPoints,
			FaceDownCards: fdCards,
			Rank:          r.Rank,
		})
	}

	gm.hub.BroadcastToRoom(session.RoomCode, WSMessage{
		Type:    MsgTypeGameOver,
		Payload: PayloadGameOver{Results: payloadResults},
	})

	// Trigger persistence callback
	if gm.OnGameComplete != nil {
		go gm.OnGameComplete(session)
	}

	// Remove session
	gm.mu.Lock()
	delete(gm.sessions, session.RoomCode)
	gm.mu.Unlock()
}

// startTurnTimer starts a countdown timer for the current turn
func (gm *GameManager) startTurnTimer(session *GameSession) {
	gm.cancelTurnTimer(session)
	session.turnTimer = time.AfterFunc(session.turnTimeout, func() {
		gm.handleTurnTimeout(session)
	})
}

// cancelTurnTimer cancels any active turn timer
func (gm *GameManager) cancelTurnTimer(session *GameSession) {
	if session.turnTimer != nil {
		session.turnTimer.Stop()
		session.turnTimer = nil
	}
}

// handleTurnTimeout auto-plays a face-down for the timed-out player
func (gm *GameManager) handleTurnTimeout(session *GameSession) {
	session.mu.Lock()
	defer session.mu.Unlock()

	if session.State.Status != game.StatusInProgress {
		return
	}

	seat := session.State.CurrentTurn
	if session.Players[seat].IsBot {
		return // bots handle their own turns
	}

	// Auto face-down the lowest penalty card
	hand := session.State.Hands[seat]
	if len(hand) == 0 {
		return
	}

	// Pick lowest value card
	minCard := hand[0]
	minPenalty := cardPenaltyValue(minCard, session.State.AceDirection)
	for _, c := range hand[1:] {
		p := cardPenaltyValue(c, session.State.AceDirection)
		if p < minPenalty {
			minPenalty = p
			minCard = c
		}
	}

	move := game.Move{Seat: seat, Type: game.MoveFaceDown, Card: minCard}
	newState, err := game.ApplyMove(session.State, move)
	if err != nil {
		logrus.Errorf("Turn timeout auto-move error: %v", err)
		return
	}

	session.State = newState

	// Notify
	gm.hub.BroadcastToRoom(session.RoomCode, WSMessage{
		Type: MsgTypeTurnTimeout,
		Payload: map[string]interface{}{
			"seat": seat,
		},
	})

	gm.broadcastMove(session, move, false)

	if newState.Status == game.StatusFinished {
		gm.handleGameOver(session)
		return
	}

	gm.broadcastGameState(session)
	gm.notifyTurn(session)

	if session.Players[newState.CurrentTurn].IsBot {
		go gm.executeBotTurn(session)
	}
}

// GetSession returns a session by room code
func (gm *GameManager) GetSession(roomCode string) *GameSession {
	gm.mu.RLock()
	defer gm.mu.RUnlock()
	return gm.sessions[roomCode]
}

// --- Helpers ---

func parseSuit(s string) (game.Suit, error) {
	switch s {
	case "spades":
		return game.Spades, nil
	case "hearts":
		return game.Hearts, nil
	case "diamonds":
		return game.Diamonds, nil
	case "clubs":
		return game.Clubs, nil
	default:
		return 0, fmt.Errorf("invalid suit: %s", s)
	}
}

func cardPenaltyValue(card game.Card, aceDir game.AceDirection) int {
	if card.Rank == game.Ace {
		if aceDir == game.AceHigh {
			return 14
		}
		return 1
	}
	return int(card.Rank)
}

// ParsePayload is a helper to unmarshal WSMessage payloads
func ParsePayload[T any](msg WSMessage) (T, error) {
	var result T
	data, err := json.Marshal(msg.Payload)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(data, &result)
	return result, err
}
