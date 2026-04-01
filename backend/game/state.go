package game

// AceDirection determines how Aces score and close suits.
// Once the first Ace is played, it locks the direction for all suits.
type AceDirection int

const (
	AceUndecided AceDirection = iota // no ace played yet
	AceLow                           // Ace follows 2 (A,2,3,...); penalty = 1
	AceHigh                          // Ace follows King (...Q,K,A); penalty = 14
)

func (d AceDirection) String() string {
	switch d {
	case AceLow:
		return "low"
	case AceHigh:
		return "high"
	default:
		return "undecided"
	}
}

// SuitSequence tracks a single suit's played sequence on the table.
// The sequence grows outward from 7 in both directions.
// LowEnd is the lowest rank currently in the sequence.
// HighEnd is the highest rank currently in the sequence.
// Closed is true when an Ace terminates the sequence.
type SuitSequence struct {
	Suit    Suit `json:"suit"`
	LowEnd  Rank `json:"low_end"`  // lowest rank played (starts at 7, goes down)
	HighEnd Rank `json:"high_end"` // highest rank played (starts at 7, goes up)
	Started bool `json:"started"`  // whether the 7 has been played for this suit
	Closed  bool `json:"closed"`   // whether an Ace has closed this sequence
}

// CanPlayRank checks if a given rank can be added to this sequence.
func (s *SuitSequence) CanPlayRank(rank Rank, aceDir AceDirection) bool {
	if !s.Started {
		return rank == Seven
	}
	if s.Closed {
		return false
	}

	// Can extend low end (play the card one below current low)
	if rank == s.LowEnd-1 && s.LowEnd > Ace {
		return true
	}

	// Can extend high end (play the card one above current high)
	if rank == s.HighEnd+1 && s.HighEnd < King {
		return true
	}

	// Ace special cases
	if rank == Ace {
		switch aceDir {
		case AceLow:
			// Ace can be played below 2
			return s.LowEnd == Two
		case AceHigh:
			// Ace can be played above King
			return s.HighEnd == King
		case AceUndecided:
			// Either direction is possible
			return s.LowEnd == Two || s.HighEnd == King
		}
	}

	return false
}

// MoveType represents the kind of move a player makes
type MoveType int

const (
	MovePlayCard MoveType = iota // play a card onto a sequence
	MoveFaceDown                 // place a card face-down as penalty
)

func (m MoveType) String() string {
	if m == MovePlayCard {
		return "play"
	}
	return "face_down"
}

// Move represents a single player action
type Move struct {
	Seat     int      `json:"seat"`
	Type     MoveType `json:"type"`
	Card     Card     `json:"card"`
	MoveNum  int      `json:"move_num"`
}

// PlayerResult holds end-of-game results for one player
type PlayerResult struct {
	Seat          int    `json:"seat"`
	PenaltyPoints int    `json:"penalty_points"`
	FaceDownCards []Card `json:"face_down_cards"`
	Rank          int    `json:"rank"` // 1 = winner
}

// GameStatus represents the current phase of the game
type GameStatus int

const (
	StatusWaiting    GameStatus = iota // waiting for players
	StatusDealing                      // cards being dealt
	StatusInProgress                   // game is active
	StatusFinished                     // game is over
)

func (s GameStatus) String() string {
	switch s {
	case StatusWaiting:
		return "waiting"
	case StatusDealing:
		return "dealing"
	case StatusInProgress:
		return "in_progress"
	case StatusFinished:
		return "finished"
	default:
		return "unknown"
	}
}

// GameState holds the complete state of a game in progress
type GameState struct {
	// Players
	NumPlayers   int       `json:"num_players"`
	Hands        [][]Card  `json:"-"`          // each player's hand (hidden from others)
	FaceDown     [][]Card  `json:"-"`          // each player's face-down penalty cards
	HandCounts   []int     `json:"hand_counts"` // number of cards in each hand (public info)

	// Table
	Sequences    [4]SuitSequence `json:"sequences"` // one per suit: Spades, Hearts, Diamonds, Clubs

	// Game flow
	CurrentTurn  int          `json:"current_turn"`   // seat index (0-3)
	AceDirection AceDirection `json:"ace_direction"`
	Status       GameStatus   `json:"status"`
	MoveHistory  []Move       `json:"move_history"`
	MoveCount    int          `json:"move_count"`

	// Results (populated when Status == StatusFinished)
	Results      []PlayerResult `json:"results,omitempty"`
}

// NewGameState creates a fresh game state with dealt cards.
// It finds the player holding 7♠ and sets them as first turn.
func NewGameState(hands [][]Card) *GameState {
	numPlayers := len(hands)

	state := &GameState{
		NumPlayers:   numPlayers,
		Hands:        hands,
		FaceDown:     make([][]Card, numPlayers),
		HandCounts:   make([]int, numPlayers),
		AceDirection: AceUndecided,
		Status:       StatusInProgress,
		MoveHistory:  make([]Move, 0, 52),
	}

	// Initialize sequences
	for i, suit := range AllSuits() {
		state.Sequences[i] = SuitSequence{
			Suit:    suit,
			LowEnd:  Seven,
			HighEnd: Seven,
			Started: false,
		}
	}

	// Initialize face-down slices and hand counts
	for i := range numPlayers {
		state.FaceDown[i] = make([]Card, 0)
		state.HandCounts[i] = len(hands[i])
	}

	// Find who holds 7♠
	sevenSpades := Card{Suit: Spades, Rank: Seven}
	for seat, hand := range hands {
		for _, card := range hand {
			if card.Equal(sevenSpades) {
				state.CurrentTurn = seat
				return state
			}
		}
	}

	// Fallback (shouldn't happen with a valid deck)
	state.CurrentTurn = 0
	return state
}

// HandFor returns a copy of the hand for the given seat
func (g *GameState) HandFor(seat int) []Card {
	if seat < 0 || seat >= g.NumPlayers {
		return nil
	}
	hand := make([]Card, len(g.Hands[seat]))
	copy(hand, g.Hands[seat])
	return hand
}

// FaceDownFor returns a copy of the face-down cards for the given seat
func (g *GameState) FaceDownFor(seat int) []Card {
	if seat < 0 || seat >= g.NumPlayers {
		return nil
	}
	cards := make([]Card, len(g.FaceDown[seat]))
	copy(cards, g.FaceDown[seat])
	return cards
}

// IsGameOver checks if all players have emptied their hands
func (g *GameState) IsGameOver() bool {
	for _, count := range g.HandCounts {
		if count > 0 {
			return false
		}
	}
	return true
}

// sequenceIndex returns the index in Sequences for a given suit
func sequenceIndex(suit Suit) int {
	return int(suit)
}
