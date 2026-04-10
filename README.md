# Game Data Platform

Subscription-based REST API for real-time game analytics. Users purchase tiered packages, generate API keys, and query game data within rate-limited access.

---

## Tech Stack

| Layer    | Stack                                              |
|----------|----------------------------------------------------|
| Backend  | Go, Gin, PostgreSQL 16, JWT, bcrypt                |
| Frontend | React 19, TypeScript, Vite, Tailwind CSS, Recharts |
| Infra    | Docker Compose                                     |

---

## Quick Start

### 1. Database

```bash
cd Backend
docker compose up -d
```

PostgreSQL starts on port `5435`. Schema auto-initializes from `initdb/init.sql`.

### 2. Backend

```bash
cd Backend/app
air          # hot reload
# or
go run main.go
```

Runs on `http://localhost:8080`.

### 3. Frontend

```bash
cd Frontend
npm install
npm run dev
```

Runs on `http://localhost:5173`.

---

## API Endpoints

### Auth

| Method | Path                   | Description        |
|--------|------------------------|--------------------|
| POST   | `/api/users/register`  | Create account     |
| POST   | `/api/users/login`     | Get JWT tokens     |
| POST   | `/api/users/refresh`   | Refresh access token |

### Users (JWT required)

| Method | Path                        | Description       |
|--------|-----------------------------|-------------------|
| GET    | `/api/users/:id`            | Get profile       |
| POST   | `/api/users/:id/topup`      | Add balance       |
| POST   | `/api/users/:id/keys`       | Generate API key  |
| GET    | `/api/users/:id/keys`       | List API keys     |
| DELETE | `/api/users/:id/keys/:key`  | Delete API key    |

### Packages

| Method | Path                          | Description           |
|--------|-------------------------------|-----------------------|
| GET    | `/api/packages`               | List all packages     |
| GET    | `/api/packages/:id`           | Get package details   |
| POST   | `/api/packages/purchase`      | Purchase or upgrade   |
| GET    | `/api/packages/subscription`  | Get active subscription |

### Games (API key + rate limit)

| Method | Path              | Description         |
|--------|-------------------|---------------------|
| GET    | `/api/games`      | List all games      |
| GET    | `/api/games/:id`  | Get game details    |
| POST   | `/api/games`      | Create game (admin) |
| PUT    | `/api/games/:id`  | Update game (admin) |
| DELETE | `/api/games/:id`  | Delete game (admin) |

---

## Subscription Tiers

| Tier       | Price  | Requests        | Refresh Interval | Historical Data |
|------------|--------|-----------------|------------------|-----------------|
| Standard   | $29    | 100 / window    | 90 min           | 60 days         |
| Platinum   | $149   | 5,000 / window  | 5 min            | 2 years         |
| Enterprise | $499   | Unlimited       | 1 min            | 5 years         |

Enterprise includes: revenue analytics, region breakdown, webhooks, bulk export, custom reports, dedicated support, SLA.

---

## Architecture

```
Client
  |
  |-- Bearer token --> /api/users/*      --> RequireJWT + RequireSelf
  |-- Bearer token --> /api/packages/*
  |-- X-API-Key    --> /api/games/*      --> Auth + RateLimit
  |
  v
[ Gin Router ] --> [ Middleware ] --> [ Handler ] --> [ Service ] --> [ Repository ] --> [ PostgreSQL ]
```

Each module (`user`, `game`, `pkg`) follows the same layered pattern: handler for HTTP, service for business logic, repository for data access. No ORM — raw SQL throughout.

**Middleware chain:**
- `RequireJWT` — validates Bearer token, sets user context
- `RequireSelf` — ensures `:id` param matches authenticated user
- `Auth` — validates `X-API-Key` header
- `RateLimit` — enforces per-package request limits via `api_usage_logs`
- `Admin` — restricts write operations to admin role

---

## Project Structure

```
Backend/
  app/
    main.go              # entry point, routes, background worker
    db/                  # database connection
    middleware/          # auth, rate limit, JWT, RBAC
    user/                # registration, login, keys, profile
    game/                # CRUD, image upload/compression
    pkg/                 # packages, subscriptions, purchase logic
  initdb/init.sql        # full schema (10 tables)
  docker-compose.yml     # PostgreSQL service

Frontend/
  src/
    pages/               # Dashboard, Games, GameDetails, Profile, Login, Register
    components/layout/   # Sidebar, DashboardLayout
    services/            # axios instance, auth, game, package API clients
    types/               # TypeScript interfaces
```

---

## Notes

- A Postman collection is available at `Backend/app/GameDataPlatform.postman_collection.json`
- Subscriptions expire in 1 day in dev mode for rapid testing
- A background goroutine simulates player counts every 10 minutes
- Game images are auto-compressed to JPEG (max 5MB) on upload
