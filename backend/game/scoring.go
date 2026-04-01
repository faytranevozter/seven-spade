package game

import "sort"

// CalculatePenalties computes penalty points for all players based on
// their face-down cards. Returns results sorted by penalty (ascending).
func CalculatePenalties(state *GameState) []PlayerResult {
	results := make([]PlayerResult, state.NumPlayers)

	for seat := 0; seat < state.NumPlayers; seat++ {
		total := 0
		for _, card := range state.FaceDown[seat] {
			total += cardPenalty(card, state.AceDirection)
		}

		results[seat] = PlayerResult{
			Seat:          seat,
			PenaltyPoints: total,
			FaceDownCards: state.FaceDownFor(seat),
		}
	}

	// Assign ranks based on penalty (lower is better)
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].PenaltyPoints < results[j].PenaltyPoints
	})

	for i := range results {
		results[i].Rank = i + 1
	}

	// Handle ties: players with equal penalties share the better rank
	for i := 1; i < len(results); i++ {
		if results[i].PenaltyPoints == results[i-1].PenaltyPoints {
			results[i].Rank = results[i-1].Rank
		}
	}

	return results
}

// cardPenalty returns the penalty value for a single face-down card.
// 2–10 = face value, J=11, Q=12, K=13, A=1 (low) or 14 (high).
func cardPenalty(card Card, aceDir AceDirection) int {
	if card.Rank == Ace {
		switch aceDir {
		case AceHigh:
			return 14
		case AceLow:
			return 1
		default:
			// If ace direction was never decided, worst case = 14
			return 14
		}
	}
	return int(card.Rank)
}
