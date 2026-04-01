package game

import (
	"fmt"
	"math/rand"
)

// Suit represents a card suit
type Suit int

const (
	Spades Suit = iota
	Hearts
	Diamonds
	Clubs
)

var SuitNames = map[Suit]string{
	Spades:   "spades",
	Hearts:   "hearts",
	Diamonds: "diamonds",
	Clubs:    "clubs",
}

var SuitSymbols = map[Suit]string{
	Spades:   "♠",
	Hearts:   "♥",
	Diamonds: "♦",
	Clubs:    "♣",
}

func (s Suit) String() string {
	return SuitNames[s]
}

func (s Suit) Symbol() string {
	return SuitSymbols[s]
}

// AllSuits returns all four suits
func AllSuits() []Suit {
	return []Suit{Spades, Hearts, Diamonds, Clubs}
}

// Rank represents a card rank (1=Ace, 2-10, 11=Jack, 12=Queen, 13=King)
type Rank int

const (
	Ace   Rank = 1
	Two   Rank = 2
	Three Rank = 3
	Four  Rank = 4
	Five  Rank = 5
	Six   Rank = 6
	Seven Rank = 7
	Eight Rank = 8
	Nine  Rank = 9
	Ten   Rank = 10
	Jack  Rank = 11
	Queen Rank = 12
	King  Rank = 13
)

var RankNames = map[Rank]string{
	Ace: "A", Two: "2", Three: "3", Four: "4", Five: "5",
	Six: "6", Seven: "7", Eight: "8", Nine: "9", Ten: "10",
	Jack: "J", Queen: "Q", King: "K",
}

func (r Rank) String() string {
	return RankNames[r]
}

// PenaltyValue returns the face-down penalty value for a rank.
// For Ace, use PenaltyValueAce with the ace direction.
func (r Rank) PenaltyValue() int {
	if r == Ace {
		return 1 // default; caller should use PenaltyValueAce
	}
	return int(r)
}

// AllRanks returns ranks Ace through King
func AllRanks() []Rank {
	return []Rank{Ace, Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten, Jack, Queen, King}
}

// Card represents a playing card
type Card struct {
	Suit Suit `json:"suit"`
	Rank Rank `json:"rank"`
}

func (c Card) String() string {
	return fmt.Sprintf("%s%s", c.Rank.String(), c.Suit.Symbol())
}

func (c Card) Equal(other Card) bool {
	return c.Suit == other.Suit && c.Rank == other.Rank
}

// Deck represents a standard 52-card deck
type Deck struct {
	Cards []Card
}

// NewDeck creates a standard 52-card deck in order
func NewDeck() *Deck {
	cards := make([]Card, 0, 52)
	for _, suit := range AllSuits() {
		for _, rank := range AllRanks() {
			cards = append(cards, Card{Suit: suit, Rank: rank})
		}
	}
	return &Deck{Cards: cards}
}

// Shuffle randomizes the deck using the provided source
func (d *Deck) Shuffle(rng *rand.Rand) {
	rng.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
}

// Deal distributes cards to n players evenly.
// For 52 cards and 4 players, each gets 13 cards.
func (d *Deck) Deal(numPlayers int) [][]Card {
	hands := make([][]Card, numPlayers)
	for i := range hands {
		hands[i] = make([]Card, 0, len(d.Cards)/numPlayers)
	}
	for i, card := range d.Cards {
		hands[i%numPlayers] = append(hands[i%numPlayers], card)
	}
	return hands
}
