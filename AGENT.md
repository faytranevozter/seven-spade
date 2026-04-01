# AGENT.md — Seven Spade

Instructions for AI coding agents working on this project.

## Project Overview

Seven Spade is a real-time 4-player card game. The codebase is a monorepo with a Go backend and React/TypeScript frontend.

## Architecture

- **Backend** (`/backend`): Go, Gin framework, GORM ORM, gorilla/websocket. Clean Architecture pattern (domain → repository → service → handler).
- **Frontend** (`/frontend`): React 19, TypeScript, Vite, TailwindCSS v4, Zustand state management, Framer Motion animations.
- **Database**: PostgreSQL 16 (GORM auto-migrates), Redis 7 for caching/matchmaking.
- **Real-time**: WebSocket via gorilla/websocket on backend, native WebSocket on frontend.

## Key Directories

| Path | Purpose |
|------|---------|
| `backend/game/` | Pure game engine — no I/O, fully unit-tested |
| `backend/internal/ws/` | WebSocket hub, client management, game session orchestration |
| `backend/internal/rest/` | REST handlers and middleware |
| `backend/internal/repository/gorm/` | Database access layer |
| `backend/domain/` | Domain models and filters |
| `backend/app/main.go` | Entry point — wires everything together |
| `frontend/src/components/` | React components (Card SVG, GameBoard, PlayerHand, etc.) |
| `frontend/src/pages/` | Route-level pages |
| `frontend/src/stores/` | Zustand stores (auth, game) |
| `frontend/src/lib/` | API client, WebSocket client, utilities |

## Build & Test Commands

```bash
# Backend
cd backend
go build ./...           # Compile check
go test ./game/... -v    # Game engine tests
go test ./...            # All tests
go run ./app             # Run dev server (port 5050)

# Frontend
cd frontend
npm install              # Install deps
npm run dev              # Dev server (port 5173)
npm run build            # Production build
npx tsc --noEmit         # Type check only

# Docker
docker compose up --build -d       # Full stack
docker compose up -d postgres redis # DB only (for local dev)
```

## Environment Setup

- Go is at `/usr/local/go/bin` — you may need: `export PATH=$PATH:/usr/local/go/bin`
- Node.js is via nvm: `export NVM_DIR="$HOME/.nvm" && . "$NVM_DIR/nvm.sh"`
- **Local dev** backend config: `backend/.env` (copy from `backend/.env.example`)
- **Docker Compose** config: root `.env` (copy from `.env.example`)

## Coding Conventions

- **Backend**: Go standard formatting (`gofmt`). Error handling follows Go idioms. Domain models use `uuid.UUID` for IDs. GORM tags on struct fields. JSON tags use `snake_case`.
- **Frontend**: TypeScript strict mode. Functional components only. Zustand for global state. TailwindCSS for styling (no CSS modules). Framer Motion for animations.
- **Naming**: Files use `snake_case.go` (backend) and `PascalCase.tsx` (frontend components) / `camelCase.ts` (utilities/stores).
- **WebSocket messages**: Types are string constants defined in `backend/internal/ws/messages.go`. Payloads use `snake_case` JSON.

## Game Engine

The game engine in `backend/game/` is pure logic with no side effects:
- `card.go` — Card, Suit, Rank types, deck creation and shuffling
- `state.go` — GameState, SuitSequence, move tracking
- `rules.go` — ValidMoves(), MustFaceDown(), move validation
- `engine.go` — ApplyMove() state transitions
- `scoring.go` — Penalty calculation and ranking
- `bot.go` — AI players (Easy/Medium/Hard difficulty)

Changes to game rules should be made here and covered by tests in `engine_test.go`.

## Common Tasks

- **Add a new REST endpoint**: Create handler in `backend/internal/rest/`, register route in `backend/app/main.go`.
- **Add a new WS message type**: Add constant in `messages.go`, add payload struct, handle in `handleWSMessage()` in `main.go`.
- **Add a new frontend page**: Create in `frontend/src/pages/`, add route in `App.tsx`.
- **Modify game rules**: Edit `backend/game/rules.go` or `engine.go`, update tests.
