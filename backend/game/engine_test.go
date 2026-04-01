package game

import (
	"math/rand"
	"testing"
)

func TestNewDeck(t *testing.T) {
	deck := NewDeck()
	if len(deck.Cards) != 52 {
		t.Errorf("expected 52 cards, got %d", len(deck.Cards))
	}

	// Check uniqueness
	seen := make(map[Card]bool)
	for _, card := range deck.Cards {
		if seen[card] {
			t.Errorf("duplicate card: %s", card)
		}
		seen[card] = true
	}
}

func TestDeal(t *testing.T) {
	deck := NewDeck()
	rng := rand.New(rand.NewSource(42))
	deck.Shuffle(rng)
	hands := deck.Deal(4)

	if len(hands) != 4 {
		t.Fatalf("expected 4 hands, got %d", len(hands))
	}

	for i, hand := range hands {
		if len(hand) != 13 {
			t.Errorf("hand %d has %d cards, expected 13", i, len(hand))
		}
	}

	// All 52 cards should be distributed
	allCards := make(map[Card]bool)
	for _, hand := range hands {
		for _, card := range hand {
			allCards[card] = true
		}
	}
	if len(allCards) != 52 {
		t.Errorf("expected 52 unique cards across all hands, got %d", len(allCards))
	}
}

func TestNewGameState_FindsSevenSpades(t *testing.T) {
	deck := NewDeck()
	rng := rand.New(rand.NewSource(42))
	deck.Shuffle(rng)
	hands := deck.Deal(4)
	state := NewGameState(hands)

	// The player at CurrentTurn should hold 7♠
	sevenSpades := Card{Suit: Spades, Rank: Seven}
	found := false
	for _, card := range state.Hands[state.CurrentTurn] {
		if card.Equal(sevenSpades) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("player at seat %d should hold 7♠", state.CurrentTurn)
	}
}

func TestFirstMoveMustBeSevenSpades(t *testing.T) {
	deck := NewDeck()
	rng := rand.New(rand.NewSource(42))
	deck.Shuffle(rng)
	hands := deck.Deal(4)
	state := NewGameState(hands)

	moves := ValidMoves(state, state.CurrentTurn)
	if len(moves) != 1 {
		t.Fatalf("expected exactly 1 valid first move, got %d", len(moves))
	}

	if !moves[0].Card.Equal(Card{Suit: Spades, Rank: Seven}) {
		t.Errorf("first move must be 7♠, got %s", moves[0].Card)
	}
}

func TestPlaySevenSpades(t *testing.T) {
	deck := NewDeck()
	rng := rand.New(rand.NewSource(42))
	deck.Shuffle(rng)
	hands := deck.Deal(4)
	state := NewGameState(hands)

	move := Move{
		Seat: state.CurrentTurn,
		Type: MovePlayCard,
		Card: Card{Suit: Spades, Rank: Seven},
	}

	newState, err := ApplyMove(state, move)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Spades sequence should be started
	if !newState.Sequences[sequenceIndex(Spades)].Started {
		t.Error("Spades sequence should be started")
	}

	// Move count should be 1
	if newState.MoveCount != 1 {
		t.Errorf("expected move count 1, got %d", newState.MoveCount)
	}
}

func TestCannotPlayOutOfTurn(t *testing.T) {
	deck := NewDeck()
	rng := rand.New(rand.NewSource(42))
	deck.Shuffle(rng)
	hands := deck.Deal(4)
	state := NewGameState(hands)

	wrongSeat := (state.CurrentTurn + 1) % 4
	move := Move{
		Seat: wrongSeat,
		Type: MovePlayCard,
		Card: Card{Suit: Spades, Rank: Seven},
	}

	_, err := ApplyMove(state, move)
	if err == nil {
		t.Error("expected error for playing out of turn")
	}
}

func TestFaceDownWhenNoValidPlays(t *testing.T) {
	// Create a contrived situation where a player has no valid plays
	// Player 0 has: 7♠
	// Player 1 has: K♣ (no valid play after 7♠ is played)
	// Player 2 has: Q♣
	// Player 3 has: J♣

	hands := [][]Card{
		{{Suit: Spades, Rank: Seven}},
		{{Suit: Clubs, Rank: King}},
		{{Suit: Clubs, Rank: Queen}},
		{{Suit: Clubs, Rank: Jack}},
	}

	state := NewGameState(hands)

	// Play 7♠
	state, err := ApplyMove(state, Move{
		Seat: 0, Type: MovePlayCard, Card: Card{Suit: Spades, Rank: Seven},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Player 1's turn: should have no valid plays (only has K♣, no 7♣ started)
	if !MustFaceDown(state, state.CurrentTurn) {
		t.Error("player 1 should have no valid plays")
	}

	// Face down K♣
	state, err = ApplyMove(state, Move{
		Seat: state.CurrentTurn, Type: MoveFaceDown, Card: Card{Suit: Clubs, Rank: King},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Player 1 should have K♣ in face-down pile
	if len(state.FaceDown[1]) != 1 {
		t.Errorf("expected 1 face-down card for player 1, got %d", len(state.FaceDown[1]))
	}
}

func TestGameOver(t *testing.T) {
	// Simple 2-move game: player 0 plays 7♠, player 1 has 6♠, etc.
	hands := [][]Card{
		{{Suit: Spades, Rank: Seven}},
		{{Suit: Spades, Rank: Eight}},
		{{Suit: Spades, Rank: Six}},
		{{Suit: Spades, Rank: Nine}},
	}

	state := NewGameState(hands)

	// Play 7♠
	state, _ = ApplyMove(state, Move{Seat: 0, Type: MovePlayCard, Card: Card{Suit: Spades, Rank: Seven}})
	// Play 8♠
	state, _ = ApplyMove(state, Move{Seat: 1, Type: MovePlayCard, Card: Card{Suit: Spades, Rank: Eight}})
	// Play 6♠
	state, _ = ApplyMove(state, Move{Seat: 2, Type: MovePlayCard, Card: Card{Suit: Spades, Rank: Six}})
	// Play 9♠
	state, _ = ApplyMove(state, Move{Seat: 3, Type: MovePlayCard, Card: Card{Suit: Spades, Rank: Nine}})

	if state.Status != StatusFinished {
		t.Error("game should be finished")
	}

	if state.Results == nil {
		t.Fatal("results should be populated")
	}

	// All players should have 0 penalty (no face-down cards)
	for _, result := range state.Results {
		if result.PenaltyPoints != 0 {
			t.Errorf("seat %d should have 0 penalty, got %d", result.Seat, result.PenaltyPoints)
		}
	}
}

func TestAceLocking(t *testing.T) {
	// Setup: spades sequence at 2,3,4,5,6,7; player plays Ace♠ after 2
	// This should lock ace direction to Low

	state := &GameState{
		NumPlayers: 4,
		Hands: [][]Card{
			{{Suit: Spades, Rank: Ace}},
			{},
			{},
			{},
		},
		FaceDown:     make([][]Card, 4),
		HandCounts:   []int{1, 0, 0, 0},
		AceDirection: AceUndecided,
		Status:       StatusInProgress,
		MoveHistory:  make([]Move, 0),
		MoveCount:    10, // pretend moves have happened
		CurrentTurn:  0,
	}

	// Init face-down slices
	for i := range 4 {
		state.FaceDown[i] = make([]Card, 0)
	}

	// Set up spades sequence: low=2, high=7
	state.Sequences[sequenceIndex(Spades)] = SuitSequence{
		Suit:    Spades,
		LowEnd:  Two,
		HighEnd: Seven,
		Started: true,
	}

	// Play Ace♠
	state, err := ApplyMove(state, Move{
		Seat: 0, Type: MovePlayCard, Card: Card{Suit: Spades, Rank: Ace},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if state.AceDirection != AceLow {
		t.Errorf("expected AceLow, got %s", state.AceDirection)
	}

	if !state.Sequences[sequenceIndex(Spades)].Closed {
		t.Error("spades should be closed after Ace")
	}
}

func TestPenaltyCalculation(t *testing.T) {
	state := &GameState{
		NumPlayers:   4,
		Hands:        make([][]Card, 4),
		HandCounts:   []int{0, 0, 0, 0},
		AceDirection: AceLow,
		Status:       StatusFinished,
		FaceDown: [][]Card{
			{{Suit: Spades, Rank: King}, {Suit: Hearts, Rank: Queen}},  // 13+12=25
			{{Suit: Spades, Rank: Ace}},                                 // 1 (ace low)
			{},                                                          // 0
			{{Suit: Clubs, Rank: Ten}, {Suit: Diamonds, Rank: Five}},    // 10+5=15
		},
	}

	results := CalculatePenalties(state)

	expectedPenalties := map[int]int{0: 25, 1: 1, 2: 0, 3: 15}
	for _, r := range results {
		if r.PenaltyPoints != expectedPenalties[r.Seat] {
			t.Errorf("seat %d: expected penalty %d, got %d", r.Seat, expectedPenalties[r.Seat], r.PenaltyPoints)
		}
	}

	// Winner should be seat 2 (0 penalty)
	if results[0].Seat != 2 || results[0].Rank != 1 {
		t.Errorf("expected seat 2 to be rank 1, got seat %d rank %d", results[0].Seat, results[0].Rank)
	}
}

func TestPenaltyAceHigh(t *testing.T) {
	state := &GameState{
		NumPlayers:   2,
		Hands:        make([][]Card, 2),
		HandCounts:   []int{0, 0},
		AceDirection: AceHigh,
		Status:       StatusFinished,
		FaceDown: [][]Card{
			{{Suit: Spades, Rank: Ace}}, // 14 (ace high)
			{},
		},
	}

	results := CalculatePenalties(state)
	for _, r := range results {
		if r.Seat == 0 && r.PenaltyPoints != 14 {
			t.Errorf("expected 14 penalty for Ace (high), got %d", r.PenaltyPoints)
		}
	}
}

func TestBotChooseMove(t *testing.T) {
	deck := NewDeck()
	rng := rand.New(rand.NewSource(42))
	deck.Shuffle(rng)
	hands := deck.Deal(4)
	state := NewGameState(hands)

	for _, diff := range []BotDifficulty{BotEasy, BotMedium, BotHard} {
		move := BotChooseMove(state, state.CurrentTurn, diff, rng)
		if !move.Card.Equal(Card{Suit: Spades, Rank: Seven}) {
			t.Errorf("bot (%s) first move must be 7♠, got %s", diff, move.Card)
		}
	}
}

func TestFullGameWithBots(t *testing.T) {
	rng := rand.New(rand.NewSource(123))
	deck := NewDeck()
	deck.Shuffle(rng)
	hands := deck.Deal(4)
	state := NewGameState(hands)

	maxMoves := 200 // safety limit
	for state.Status == StatusInProgress && maxMoves > 0 {
		move := BotChooseMove(state, state.CurrentTurn, BotMedium, rng)
		var err error
		state, err = ApplyMove(state, move)
		if err != nil {
			t.Fatalf("move %d error: %v", state.MoveCount, err)
		}
		maxMoves--
	}

	if state.Status != StatusFinished {
		t.Errorf("game did not finish after %d moves", state.MoveCount)
	}

	if state.Results == nil {
		t.Fatal("results should be populated")
	}

	// All hands should be empty
	for seat, count := range state.HandCounts {
		if count != 0 {
			t.Errorf("seat %d still has %d cards", seat, count)
		}
	}

	t.Logf("Game finished in %d moves", state.MoveCount)
	for _, r := range state.Results {
		t.Logf("  Seat %d: penalty=%d, rank=%d, face-down=%d cards",
			r.Seat, r.PenaltyPoints, r.Rank, len(r.FaceDownCards))
	}
}

func TestSequenceCanPlayRank(t *testing.T) {
	seq := SuitSequence{
		Suit:    Spades,
		LowEnd:  Five,
		HighEnd: Nine,
		Started: true,
	}

	// Can play 4 (extend low) and 10 (extend high)
	if !seq.CanPlayRank(Four, AceUndecided) {
		t.Error("should be able to play 4")
	}
	if !seq.CanPlayRank(Ten, AceUndecided) {
		t.Error("should be able to play 10")
	}

	// Cannot play 3, 11, or 7
	if seq.CanPlayRank(Three, AceUndecided) {
		t.Error("should not be able to play 3 (skip)")
	}
	if seq.CanPlayRank(Seven, AceUndecided) {
		t.Error("should not be able to play 7 (already in sequence)")
	}
}
