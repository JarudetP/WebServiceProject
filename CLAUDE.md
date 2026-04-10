# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Game Data Platform — a web service with a Go/Gin backend, React/TypeScript frontend, and PostgreSQL database. Users register, purchase subscription packages (Standard/Platinum/Enterprise), obtain API keys, and query game data within rate limits.

## Development Commands

### Database
```bash
cd Backend && docker compose up -d   # Start PostgreSQL on port 5435
```

### Backend (Go + Air hot-reload)
```bash
cd Backend/app && air                # Dev server with hot reload on :8080
cd Backend/app && go build -o ./tmp/main.exe .  # Manual build
```

### Frontend (React + Vite)
```bash
cd Frontend && npm install           # Install dependencies
cd Frontend && npm run dev           # Dev server with HMR
cd Frontend && npx tsc -b && npx vite build  # Production build
cd Frontend && npx eslint .          # Lint
```

## Architecture

### Backend (`Backend/app/`)

Layered architecture: **Handler → Service → Repository** with dependency injection wired in `main.go`.

- **Modules**: `user/`, `pkg/` (packages/subscriptions), `game/` — each with handler, service, repository, model files
- **Middleware** (`middleware/middleware.go`): `Auth()` validates X-API-Key, `RateLimit()` enforces per-subscription limits, `RequireJWT()` validates Bearer tokens, `RequireSelf()` ensures user can only access own data, `Admin()` checks admin role
- **Database** (`db/postgres.go`): Raw SQL via `database/sql` + `lib/pq`, no ORM. Single global `DB` connection.
- **Schema** (`Backend/initdb/init.sql`): 10 tables including games, users, packages, subscriptions, api_keys, api_usage_logs

### Dual Authentication

- **JWT tokens** (Bearer header): Used for user-facing endpoints (profile, top-up, key management). Access token 15min, refresh token 2 days.
- **API keys** (X-API-Key header): Used for game data endpoints. Keys are random 32-byte hex strings logged per request.

### Route Structure

| Group | Auth | Key Endpoints |
|-------|------|---------------|
| `/api/users/*` | JWT (RequireJWT + RequireSelf) | register, login, refresh, profile, topup, keys CRUD |
| `/api/packages/*` | JWT or API key depending on route | list, purchase, active subscription |
| `/api/games/*` | API Key + RateLimit | CRUD (create/update/delete require Admin) |

### Frontend (`Frontend/src/`)

- **Router** (`App.tsx`): PrivateRoute/PublicRoute guards based on localStorage token
- **Services** (`services/`): `api.ts` configures axios with interceptors that attach JWT or API key per endpoint; `auth.service.ts`, `game.service.ts`, `package.service.ts`
- **Pages**: Login, Register, Dashboard (stats + recharts pie chart), Games (table), GameDetails, Profile (keys + subscriptions + wallet)
- **Styling**: Tailwind CSS with custom color theme defined in `tailwind.config.js`

### Subscription & Rate Limiting

Packages define `request_limit` and `limit_refresh_interval`. Rate limiting counts requests in `api_usage_logs` within the interval window. `-1` request_limit = unlimited. Admins bypass rate limiting.

### Background Worker

`main.go` runs a goroutine that updates `current_players` on all games every 10 minutes with random values (player simulator for demo purposes).

## Configuration

Backend uses `.env` in `Backend/app/` with: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `PORT`, `JWT_SECRET`, `REFRESH_SECRET`.

Docker Compose maps PostgreSQL to host port **5435** (not default 5432).
