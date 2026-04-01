# CLAUDE.md — Seven Spade

Project-specific instructions for Claude when working on this codebase.

## Quick Reference

- **Backend**: Go + Gin + GORM + WebSocket, entry at `backend/app/main.go`
- **Frontend**: React 19 + TypeScript + Vite + TailwindCSS v4 + Zustand
- **DB**: PostgreSQL 16, Redis 7
- **Auth**: OAuth2 (Google/GitHub/Telegram) → JWT

## Before Making Changes

1. Run `cd backend && go build ./...` to verify backend compiles.
2. Run `cd backend && go test ./game/...` to verify game engine tests pass (14 tests).
3. Run `cd frontend && npx tsc --noEmit` to verify frontend types.

## Environment Notes

- Go binary: `/usr/local/go/bin/go` — run `export PATH=$PATH:/usr/local/go/bin` if `go` is not found.
- Node.js via nvm: `export NVM_DIR="$HOME/.nvm" && . "$NVM_DIR/nvm.sh"` before using `npm`/`npx`.
- **Local dev**: backend reads `backend/.env` — copy from `backend/.env.example`.
- **Docker Compose**: uses root `.env` — copy from `.env.example`.
- For local dev, start Postgres + Redis: `docker compose up -d postgres redis`.

## Architecture Rules

- **Game engine** (`backend/game/`) must stay pure — no database, no HTTP, no WebSocket imports. Test via `go test ./game/...`.
- **Domain models** (`backend/domain/`) must not import internal packages.
- **Repositories** (`backend/internal/repository/gorm/`) handle only DB operations.
- **WebSocket game session** (`backend/internal/ws/game_session.go`) orchestrates game flow between engine and clients.
- **Frontend state** lives in Zustand stores, not in component state (except local UI state like form inputs).

## Key Type Mappings (Backend ↔ Frontend)

| Backend (Go) | Frontend (TypeScript) | Notes |
|---|---|---|
| `game.Card` | `Card` in `types/game.ts` | `{suit, rank}` |
| `game.Suit` (string) | `Suit` (string union) | `spades\|hearts\|diamonds\|clubs` |
| `game.Rank` (int) | `Rank` (number 1-13) | 1=Ace, 11=J, 12=Q, 13=K |
| `ws.PayloadGameState` | `GameState` | Server personalizes per player (hides others' hands) |
| `domain.User` | `User` | UUID id, display_name, elo_rating |
| `domain.Room` | `Room` | code (6-char), players array |

## WebSocket Flow

1. Client connects to `/ws/room/:code?token=JWT`
2. Server validates JWT, registers client to hub with room code
3. Client sends `request_state` → server replies with `game_state`
4. Host sends `start_game` → server creates game session, broadcasts `game_starting`
5. Game loop: `play_card` / `face_down` → server validates via engine → broadcasts `move_made` + `game_state`
6. Bots auto-play via goroutines with 1–3s delay
7. Game ends → server broadcasts `game_over`, persists results + updates ELO

## Common Pitfalls

- `token_data` in Gin context is `auth.JWTClaimUser` (from `domain/model/auth/jwt.go`), not `domain.User`.
- `JWTClaimUser.UserID` is a `string`, not `uuid.UUID` — parse with `uuid.Parse()` before using in filters.
- GORM `UserFilter.Query()` mutates `*gorm.DB` in-place (standard template pattern) — don't reassign.
- Frontend Vite proxy handles `/api`, `/auth`, `/ws` routes in dev mode — no CORS issues locally.
- The original go-template had a `Name` field on User — it was renamed to `DisplayName`. Watch for stale references.
