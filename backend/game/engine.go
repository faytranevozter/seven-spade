package game

import (
	"fmt"
)

// ApplyMove applies a move to the game state and returns the updated state.
// Returns an error if the move is invalid.
func ApplyMove(state *GameState, move Move) (*GameState, error) {
	if state.Status != StatusInProgress {
		return nil, fmt.Errorf("game is not in progress")
	}

	if move.Seat != state.CurrentTurn {
		return nil, fmt.Errorf("not seat %d's turn (current turn: %d)", move.Seat, state.CurrentTurn)
	}

	if state.HandCounts[move.Seat] == 0 {
		return nil, fmt.Errorf("seat %d has no cards", move.Seat)
	}

	switch move.Type {
	case MovePlayCard:
		return applyPlayCard(state, move)
	case MoveFaceDown:
		return applyFaceDown(state, move)
	default:
		return nil, fmt.Errorf("unknown move type: %d", move.Type)
	}
}

func applyPlayCard(state *GameState, move Move) (*GameState, error) {
	card := move.Card
	seat := move.Seat

	// Verify player holds this card
	cardIdx := findCardInHand(state.Hands[seat], card)
	if cardIdx == -1 {
		return nil, fmt.Errorf("seat %d does not hold %s", seat, card)
	}

	// Verify the card can be played
	if !canPlayCard(state, card) {
		return nil, fmt.Errorf("card %s cannot be played", card)
	}

	// Remove card from hand
	state.Hands[seat] = removeCardAt(state.Hands[seat], cardIdx)
	state.HandCounts[seat]--

	// Update the sequence
	idx := sequenceIndex(card.Suit)
	seq := &state.Sequences[idx]

	if card.Rank == Seven && !seq.Started {
		// Starting a new suit
		seq.Started = true
		seq.LowEnd = Seven
		seq.HighEnd = Seven
	} else if card.Rank == Ace {
		// Ace played: determine or enforce direction
		if state.AceDirection == AceUndecided {
			// First ace locks the direction for all suits
			if seq.LowEnd == Two {
				state.AceDirection = AceLow
				seq.LowEnd = Ace
			} else if seq.HighEnd == King {
				state.AceDirection = AceHigh
				seq.HighEnd = Ace // conceptually 14
			}
		} else if state.AceDirection == AceLow {
			seq.LowEnd = Ace
		} else {
			seq.HighEnd = Ace // conceptually 14
		}
		// Close the suit
		seq.Closed = true
	} else if card.Rank < seq.LowEnd {
		seq.LowEnd = card.Rank
	} else if card.Rank > seq.HighEnd {
		seq.HighEnd = card.Rank
	}

	// Check if suit is fully closed (both ends reached, not just ace)
	// A suit is closed when an Ace is played; we already handle that above.

	// Record the move
	state.MoveCount++
	move.MoveNum = state.MoveCount
	state.MoveHistory = append(state.MoveHistory, move)

	// Advance turn
	advanceTurn(state)

	// Check game over
	if state.IsGameOver() {
		state.Status = StatusFinished
		state.Results = CalculatePenalties(state)
	}

	return state, nil
}

func applyFaceDown(state *GameState, move Move) (*GameState, error) {
	card := move.Card
	seat := move.Seat

	// Verify player holds this card
	cardIdx := findCardInHand(state.Hands[seat], card)
	if cardIdx == -1 {
		return nil, fmt.Errorf("seat %d does not hold %s", seat, card)
	}

	// Verify player truly has no valid plays
	if !MustFaceDown(state, seat) {
		return nil, fmt.Errorf("seat %d has valid plays and cannot face-down", seat)
	}

	// Remove card from hand, add to face-down pile
	state.Hands[seat] = removeCardAt(state.Hands[seat], cardIdx)
	state.HandCounts[seat]--
	state.FaceDown[seat] = append(state.FaceDown[seat], card)

	// Record the move
	state.MoveCount++
	move.MoveNum = state.MoveCount
	state.MoveHistory = append(state.MoveHistory, move)

	// Advance turn
	advanceTurn(state)

	// Check game over
	if state.IsGameOver() {
		state.Status = StatusFinished
		state.Results = CalculatePenalties(state)
	}

	return state, nil
}

// advanceTurn moves to the next player who still has cards.
func advanceTurn(state *GameState) {
	for i := 1; i <= state.NumPlayers; i++ {
		next := (state.CurrentTurn + i) % state.NumPlayers
		if state.HandCounts[next] > 0 {
			state.CurrentTurn = next
			return
		}
	}
	// All players out of cards - game should be over
}

// findCardInHand returns the index of the card in the hand, or -1.
func findCardInHand(hand []Card, card Card) int {
	for i, c := range hand {
		if c.Equal(card) {
			return i
		}
	}
	return -1
}

// removeCardAt removes the card at index i from the slice.
func removeCardAt(hand []Card, i int) []Card {
	return append(hand[:i], hand[i+1:]...)
}
