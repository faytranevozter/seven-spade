package request_model

// --- Room requests ---

type CreateRoomRequest struct {
	BotEnabled bool `json:"bot_enabled"`
	BotCount   int  `json:"bot_count"`
	TurnTimer  int  `json:"turn_timer"` // seconds
}

type JoinRoomRequest struct {
	Code string `json:"code"`
}

type UpdateRoomSettingsRequest struct {
	BotEnabled    *bool           `json:"bot_enabled,omitempty"`
	BotCount      *int            `json:"bot_count,omitempty"`
	TurnTimer     *int            `json:"turn_timer,omitempty"`
	BotDifficulty []BotSlotConfig `json:"bot_difficulty,omitempty"`
}

type BotSlotConfig struct {
	Seat       int    `json:"seat"`
	Difficulty string `json:"difficulty"` // easy, medium, hard
}
