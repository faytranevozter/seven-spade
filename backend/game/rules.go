package game

// ValidMoves returns all legal moves for the player at the given seat.
// The first turn must play 7♠. Otherwise, returns playable cards.
// If no card can be played, returns an empty slice (player must face-down).
func ValidMoves(state *GameState, seat int) []Move {
	if state.Status != StatusInProgress || state.CurrentTurn != seat {
		return nil
	}

	hand := state.Hands[seat]
	if len(hand) == 0 {
		return nil
	}

	moves := make([]Move, 0)

	// On the very first move of the game, must play 7♠
	if state.MoveCount == 0 {
		sevenSpades := Card{Suit: Spades, Rank: Seven}
		for _, card := range hand {
			if card.Equal(sevenSpades) {
				moves = append(moves, Move{
					Seat: seat,
					Type: MovePlayCard,
					Card: card,
				})
				return moves
			}
		}
		// Should never happen if cards are dealt correctly
		return nil
	}

	// Check each card in hand against all sequences
	for _, card := range hand {
		if canPlayCard(state, card) {
			moves = append(moves, Move{
				Seat: seat,
				Type: MovePlayCard,
				Card: card,
			})
		}
	}

	return moves
}

// canPlayCard checks if a card can be legally played on the table.
func canPlayCard(state *GameState, card Card) bool {
	idx := sequenceIndex(card.Suit)
	seq := &state.Sequences[idx]

	// Starting a new suit with a 7
	if card.Rank == Seven && !seq.Started {
		return true
	}

	// Extending an existing sequence
	if seq.Started {
		return seq.CanPlayRank(card.Rank, state.AceDirection)
	}

	return false
}

// MustFaceDown returns true if the player has no valid card plays.
func MustFaceDown(state *GameState, seat int) bool {
	moves := ValidMoves(state, seat)
	return len(moves) == 0
}

// CanPlayCardPublic is exported for use by services.
func CanPlayCardPublic(state *GameState, card Card) bool {
	return canPlayCard(state, card)
}

// ValidFaceDownCards returns all cards that can be placed face-down.
// When a player has no valid plays, they must place exactly one card face-down.
// Any card in their hand is eligible.
func ValidFaceDownCards(state *GameState, seat int) []Card {
	if state.Status != StatusInProgress || state.CurrentTurn != seat {
		return nil
	}

	if !MustFaceDown(state, seat) {
		return nil // player has valid plays, can't face-down
	}

	// Return a copy of the hand (any card can be discarded face-down)
	return state.HandFor(seat)
}
