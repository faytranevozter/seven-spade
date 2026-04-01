package game

import (
	"math/rand"
	"sort"
)

// BotDifficulty represents AI difficulty levels
type BotDifficulty int

const (
	BotEasy   BotDifficulty = iota
	BotMedium
	BotHard
)

func (d BotDifficulty) String() string {
	switch d {
	case BotEasy:
		return "easy"
	case BotMedium:
		return "medium"
	case BotHard:
		return "hard"
	default:
		return "unknown"
	}
}

// BotChooseMove selects a move for the AI player.
func BotChooseMove(state *GameState, seat int, difficulty BotDifficulty, rng *rand.Rand) Move {
	switch difficulty {
	case BotEasy:
		return botEasyMove(state, seat, rng)
	case BotMedium:
		return botMediumMove(state, seat, rng)
	case BotHard:
		return botHardMove(state, seat, rng)
	default:
		return botEasyMove(state, seat, rng)
	}
}

// botEasyMove: pick a random valid play, or random face-down.
func botEasyMove(state *GameState, seat int, rng *rand.Rand) Move {
	plays := ValidMoves(state, seat)
	if len(plays) > 0 {
		return plays[rng.Intn(len(plays))]
	}

	// Must face-down: pick a random card
	hand := state.Hands[seat]
	card := hand[rng.Intn(len(hand))]
	return Move{
		Seat: seat,
		Type: MoveFaceDown,
		Card: card,
	}
}

// botMediumMove: strategic play with heuristics.
// Priorities:
// 1. Play cards that extend sequences (prefer extending towards ends)
// 2. If starting a new suit, prefer suits where we have the most cards
// 3. If face-down, discard the lowest-value card
func botMediumMove(state *GameState, seat int, rng *rand.Rand) Move {
	plays := ValidMoves(state, seat)
	if len(plays) > 0 {
		return botMediumSelectPlay(state, seat, plays, rng)
	}

	// Must face-down: discard the lowest penalty card
	return botMediumSelectFaceDown(state, seat)
}

func botMediumSelectPlay(state *GameState, seat int, plays []Move, rng *rand.Rand) Move {
	type scoredMove struct {
		move  Move
		score int
	}

	scored := make([]scoredMove, 0, len(plays))

	hand := state.Hands[seat]

	for _, m := range plays {
		score := 0
		card := m.Card

		// Prefer cards that extend sequences where we have adjacent cards
		idx := sequenceIndex(card.Suit)
		seq := state.Sequences[idx]

		if card.Rank == Seven && !seq.Started {
			// Starting a new suit: score based on how many cards of this suit we hold
			suitCount := countSuit(hand, card.Suit)
			score += suitCount * 10

			// Prefer starting suits where we have cards near 7
			for _, c := range hand {
				if c.Suit == card.Suit {
					dist := abs(int(c.Rank) - 7)
					if dist <= 3 {
						score += (4 - dist) * 5
					}
				}
			}
		} else {
			// Extending a sequence: moderate base score
			score += 20

			// Bonus if this opens up another card we hold
			if card.Rank < seq.LowEnd {
				// Extending downward: check if we hold the next lower card
				nextLow := card.Rank - 1
				if nextLow >= Ace && hasCard(hand, Card{Suit: card.Suit, Rank: nextLow}) {
					score += 30
				}
			} else if card.Rank > seq.HighEnd {
				// Extending upward
				nextHigh := card.Rank + 1
				if nextHigh <= King && hasCard(hand, Card{Suit: card.Suit, Rank: nextHigh}) {
					score += 30
				}
			}

			// Prefer playing high-value cards to avoid face-down penalties
			score += int(card.Rank)
		}

		scored = append(scored, scoredMove{move: m, score: score})
	}

	// Sort by score descending
	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Pick among top-scored moves (add slight randomness)
	if len(scored) > 1 && scored[0].score == scored[1].score {
		// Tie: pick randomly among ties
		ties := 1
		for ties < len(scored) && scored[ties].score == scored[0].score {
			ties++
		}
		return scored[rng.Intn(ties)].move
	}

	return scored[0].move
}

func botMediumSelectFaceDown(state *GameState, seat int) Move {
	hand := state.Hands[seat]

	// Discard the card with the lowest penalty value
	minPenalty := 999
	var minCard Card

	for _, card := range hand {
		p := cardPenalty(card, state.AceDirection)
		if p < minPenalty {
			minPenalty = p
			minCard = card
		}
	}

	return Move{
		Seat: seat,
		Type: MoveFaceDown,
		Card: minCard,
	}
}

// botHardMove: same as medium for now (future: minimax).
func botHardMove(state *GameState, seat int, rng *rand.Rand) Move {
	return botMediumMove(state, seat, rng)
}

// --- helpers ---

func countSuit(hand []Card, suit Suit) int {
	count := 0
	for _, c := range hand {
		if c.Suit == suit {
			count++
		}
	}
	return count
}

func hasCard(hand []Card, card Card) bool {
	for _, c := range hand {
		if c.Equal(card) {
			return true
		}
	}
	return false
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
