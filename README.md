# ♠ Seven Spade

A real-time online multiplayer card game for 4 players, built with Go and React.

## Game Rules

Seven Spade is a classic card game where players take turns playing cards to build sequences on a shared table.

- **Starting move:** The player holding the **7♠** goes first and must play it.
- **Sequence building:** Each suit has a sequence that expands outward from 7 (e.g., 6→5→4… and 8→9→10…). A new suit's sequence starts when its 7 is played.
- **Face-down penalty:** If you have no valid moves, you must place one or more cards face-down as a penalty pile.
- **Ace direction lock:** The first Ace played (as 1 low or 14 high) locks the direction for all suits.
- **Suit closing:** A suit's sequence closes when both ends (Ace and King) are reached.
- **Game over:** The game ends when all 52 cards are played or no moves remain.
- **Scoring:** Penalty = sum of face-down card values (2–10 = face value, J=11, Q=12, K=13, A=1 or 14 based on ace direction). **Lowest penalty wins.**

## Tech Stack

| Layer      | Technology                                    |
|------------|-----------------------------------------------|
| Backend    | Go · Gin · GORM · gorilla/websocket           |
| Frontend   | React 19 · TypeScript · Vite · TailwindCSS v4 |
| State      | Zustand                                        |
| Animation  | Framer Motion                                  |
| Database   | PostgreSQL 16                                  |
| Cache      | Redis 7                                        |
| Auth       | OAuth2 (Google, GitHub, Telegram) + JWT        |
| Deploy     | Docker Compose                                 |

## Project Structure

```
seven-spade/
├── backend/                   # Go backend
│   ├── app/main.go            # Entry point, wiring
│   ├── game/                  # Pure game engine (card, state, rules, engine, scoring, bot)
│   ├── domain/                # Domain entities (user, room, game)
│   ├── internal/
│   │   ├── repository/gorm/   # Database repositories
│   │   ├── rest/              # REST handlers & middleware
│   │   └── ws/                # WebSocket hub, client, game session
│   ├── room/                  # Room service
│   ├── matchmaking/           # ELO-based matchmaking service
│   ├── Dockerfile
│   └── .env.example
├── frontend/                  # React frontend
│   ├── src/
│   │   ├── components/        # Card, GameBoard, PlayerHand, SuitSequence, etc.
│   │   ├── pages/             # Login, Dashboard, RoomLobby, Game, Leaderboard
│   │   ├── stores/            # Zustand stores (auth, game)
│   │   ├── lib/               # API client, WebSocket client, card utils
│   │   └── types/             # TypeScript type definitions
│   ├── Dockerfile
│   └── nginx.conf
└── docker-compose.yml         # Full-stack orchestration
```

## Prerequisites

- **Go** 1.23+ (or use Docker)
- **Node.js** 22+ and npm (or use Docker)
- **PostgreSQL** 16+
- **Redis** 7+
- **Docker & Docker Compose** (for containerized setup)

## Running Locally (Development)

### 1. Start database services

```bash
# Start only Postgres and Redis via Docker Compose
docker compose up -d postgres redis
```

### 2. Configure backend

```bash
cd backend
cp .env.example .env
# Edit .env — fill in JWT secret and any OAuth credentials you want to use
```

### 3. Run backend

```bash
cd backend
go run ./app
# Server starts on http://localhost:5050
```

### 4. Run frontend

```bash
cd frontend
npm install
npm run dev
# Dev server starts on http://localhost:5173
# Vite proxies /api, /auth, /ws to backend automatically
```

### 5. Open the app

Visit **http://localhost:5173** in your browser.

## Running with Docker Compose (Production)

```bash
# Create backend .env
cp backend/.env.example backend/.env
# Edit backend/.env with your secrets

# Build and start everything
docker compose up --build -d

# App is available at http://localhost (port 80)
```

## API Endpoints

| Method | Path                              | Description            |
|--------|-----------------------------------|------------------------|
| GET    | `/`                               | Health check           |
| GET    | `/auth/google`                    | Google OAuth login     |
| GET    | `/auth/github`                    | GitHub OAuth login     |
| GET    | `/auth/telegram`                  | Telegram OAuth login   |
| GET    | `/api/users/me`                   | Current user profile   |
| POST   | `/api/rooms`                      | Create a room          |
| GET    | `/api/rooms/:code`                | Get room details       |
| POST   | `/api/rooms/:code/join`           | Join a room            |
| PATCH  | `/api/rooms/:code/settings`       | Update room settings   |
| DELETE | `/api/rooms/:code/players/:userId`| Kick player            |
| GET    | `/api/games/history`              | Game history           |
| GET    | `/api/games/:id`                  | Game detail            |
| GET    | `/api/games/leaderboard`          | ELO leaderboard        |
| POST   | `/api/matchmaking/queue`          | Join matchmaking       |
| DELETE | `/api/matchmaking/queue`          | Leave matchmaking      |
| GET    | `/api/matchmaking/status`         | Queue status           |
| WS     | `/ws/room/:code?token=...`        | WebSocket game channel |

## WebSocket Messages

**Client → Server:**
- `play_card` — Play a card `{suit, rank}`
- `face_down` — Face down cards `{cards: [{suit, rank}, ...]}`
- `start_game` — Host starts the game
- `request_state` — Request current game state
- `room_chat` — Send chat message

**Server → Client:**
- `game_state` — Full personalized game state (hand hidden from others)
- `move_made` — A move was made
- `game_over` — Game finished with results
- `game_starting` — Game is starting
- `turn_change` — Turn changed
- `room_chat_bcast` — Chat message broadcast
- `error` — Error message

## Running Tests

```bash
cd backend
go test ./game/... -v    # Game engine tests (14 tests)
go test ./...            # All backend tests
```

## License

MIT
