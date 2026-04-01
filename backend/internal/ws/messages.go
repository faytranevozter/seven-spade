package ws

// Message types for WebSocket communication

// --- Client → Server ---

const (
	// Room events
	MsgTypeRoomChat     = "room_chat"
	MsgTypeStartGame    = "start_game"
	MsgTypeLeaveRoom    = "leave_room"

	// Game events
	MsgTypePlayCard     = "play_card"
	MsgTypeFaceDown     = "face_down"
	MsgTypeRequestState = "request_state"
)

// --- Server → Client ---

const (
	// Room events
	MsgTypePlayerJoined     = "player_joined"
	MsgTypePlayerLeft       = "player_left"
	MsgTypePlayerKicked     = "player_kicked"
	MsgTypeSettingsChanged  = "settings_changed"
	MsgTypeGameStarting     = "game_starting"
	MsgTypeHostChanged      = "host_changed"
	MsgTypeRoomChatBcast    = "room_chat_broadcast"

	// Game events
	MsgTypeGameState        = "game_state"
	MsgTypeYourTurn         = "your_turn"
	MsgTypeMoveMade         = "move_made"
	MsgTypeInvalidMove      = "invalid_move"
	MsgTypeSuitClosed       = "suit_closed"
	MsgTypeAceLocked        = "ace_locked"
	MsgTypeGameOver         = "game_over"
	MsgTypeTurnTimeout      = "turn_timeout"

	// Matchmaking
	MsgTypeMatchFound       = "match_found"
	MsgTypeMatchmakingStatus = "matchmaking_status"

	// System
	MsgTypeError            = "error"
	MsgTypePong             = "pong"
)

// WSMessage is the envelope for all WebSocket messages
type WSMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload,omitempty"`
}

// --- Payload structs ---

type PayloadPlayCard struct {
	Suit string `json:"suit"`
	Rank int    `json:"rank"`
}

type PayloadFaceDown struct {
	Suit string `json:"suit"`
	Rank int    `json:"rank"`
}

type PayloadChat struct {
	Message string `json:"message"`
}

// Server payloads

type PayloadPlayerJoined struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	Seat        int    `json:"seat"`
	IsBot       bool   `json:"is_bot"`
}

type PayloadPlayerLeft struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	Seat        int    `json:"seat"`
}

type PayloadGameStarting struct {
	GameID string `json:"game_id"`
}

type PayloadHostChanged struct {
	NewHostID   string `json:"new_host_id"`
	DisplayName string `json:"display_name"`
}

// PayloadGameState is sent to each player with their view of the game.
// Other players' hands are hidden; only card counts are shown.
type PayloadGameState struct {
	GameID       string                `json:"game_id"`
	YourSeat     int                   `json:"your_seat"`
	YourHand     []PayloadCard         `json:"your_hand"`
	HandCounts   []int                 `json:"hand_counts"`
	Sequences    [4]PayloadSequence    `json:"sequences"`
	CurrentTurn  int                   `json:"current_turn"`
	AceDirection string                `json:"ace_direction"`
	ValidMoves   []PayloadCard         `json:"valid_moves"`
	MustFaceDown bool                  `json:"must_face_down"`
	FaceDownCounts []int              `json:"face_down_counts"`
	Players      []PayloadPlayerInfo   `json:"players"`
	Status       string                `json:"status"`
}

type PayloadCard struct {
	Suit string `json:"suit"`
	Rank int    `json:"rank"`
}

type PayloadSequence struct {
	Suit    string `json:"suit"`
	LowEnd  int    `json:"low_end"`
	HighEnd int    `json:"high_end"`
	Started bool   `json:"started"`
	Closed  bool   `json:"closed"`
}

type PayloadPlayerInfo struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	Seat        int    `json:"seat"`
	IsBot       bool   `json:"is_bot"`
	IsConnected bool   `json:"is_connected"`
}

type PayloadMoveMade struct {
	Seat     int         `json:"seat"`
	MoveType string      `json:"move_type"` // play, face_down
	Card     *PayloadCard `json:"card,omitempty"` // nil for face_down (hidden)
	MoveNum  int         `json:"move_num"`
}

type PayloadSuitClosed struct {
	Suit string `json:"suit"`
}

type PayloadAceLocked struct {
	Direction string `json:"direction"` // low, high
}

type PayloadGameOver struct {
	Results []PayloadPlayerResult `json:"results"`
}

type PayloadPlayerResult struct {
	UserID        string        `json:"user_id"`
	DisplayName   string        `json:"display_name"`
	Seat          int           `json:"seat"`
	PenaltyPoints int           `json:"penalty_points"`
	FaceDownCards []PayloadCard `json:"face_down_cards"`
	Rank          int           `json:"rank"`
}

type PayloadInvalidMove struct {
	Reason string `json:"reason"`
}

type PayloadMatchFound struct {
	RoomCode string `json:"room_code"`
}

type PayloadError struct {
	Message string `json:"message"`
}
