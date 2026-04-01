package request_model

// --- Game requests ---

type PlayCardRequest struct {
	Suit string `json:"suit"` // spades, hearts, diamonds, clubs
	Rank int    `json:"rank"` // 1-13
}

type FaceDownRequest struct {
	Suit string `json:"suit"`
	Rank int    `json:"rank"`
}
