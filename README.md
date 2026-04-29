# Game Data Platform — System Architecture

Subscription-based REST API platform for real-time game analytics. Users register, purchase tiered packages, generate API keys, and query live game data within rate-limited access tiers.

---

## Architecture Overview

```
                          ┌──────────────────────────────┐
                          │         Frontend             │
                          │   React + TypeScript + Vite  │
                          │        localhost:5173         │
                          └────┬──────────────┬──────────┘
                               │              │
                    JWT/Bearer │              │ X-API-Key
                    REST calls │              │ REST + SSE
                               │              │
               ┌───────────────▼──┐    ┌──────▼───────────────────────┐
               │   User Service   │◄───│       Game Service           │
               │     :8081        │    │          :8083               │
               │                  │    │  1. Validate API Key         │
               │  - Auth (JWT)    │    │  2. Load package features    │
               │  - API Keys      │    │  3. Enforce rate limit       │
               │  - Wallet/TopUp  │    │  4. Serve game data          │
               │  - Usage logging │    │  5. Analytics / SSE / Export │
               └──────┬───────────┘    └──────────────┬───────────────┘
                      │                               │
              balance │                               │ subscription +
              deduct  │              ┌────────────────▼  package features
                      │              │    Package Service  │
                      └──────────────►        :8082        │
                                     │                    │
                                     │  - List packages   │
                                     │  - Purchase/Upgrade│
                                     │  - Subscriptions   │
                                     └────────────────────┘
                      │                      │                    │
               ┌──────▼──────┐    ┌──────────▼──────┐    ┌───────▼───────┐
               │   user_db   │    │   package_db    │    │   game_db     │
               │ PostgreSQL  │    │   PostgreSQL    │    │  PostgreSQL   │
               │   :5431     │    │     :5432       │    │    :5437      │
               └─────────────┘    └─────────────────┘    └───────────────┘
```

---

## Services

### User Service — port 8081

Handles identity, authentication, wallet, API key management, and usage tracking.

**Public endpoints:**
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/users/register` | Register new user |
| POST | `/api/users/login` | Login → returns access + refresh token |
| POST | `/api/users/refresh` | Refresh access token |
| GET | `/api/users/:id` | Get user profile |
| POST | `/api/users/:id/topup` | Add balance to wallet |
| POST | `/api/users/:id/keys` | Generate API key |
| GET | `/api/users/:id/keys` | List API keys |
| DELETE | `/api/users/:id/keys/:key` | Delete API key |
| GET | `/api/users/:id/stats` | API usage history (30 days) |

**Internal endpoints (called by other services):**
| Method | Path | Called by |
|--------|------|-----------|
| GET | `/internal/keys/:key/validate` | Game Service — verify API key |
| GET | `/internal/usage/count?user_id&minutes` | Game Service — count requests in window |
| POST | `/internal/usage/log` | Game Service — log request after serving |
| POST | `/internal/users/:id/deduct` | Package Service — deduct wallet balance |

**JWT details:**
- Access token: HS256, expires in **15 minutes**
- Refresh token: HS256, expires in **2 days**
- Claims: `user_id`, `username`, `role`

---

### Package Service — port 8082

Handles package catalog, subscription lifecycle, and payment recording.

**Endpoints:**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/packages` | List all packages (sorted by price) |
| GET | `/api/packages/:id` | Get single package details |
| POST | `/api/packages/purchase?user_id=` | Purchase / upgrade / renew |
| GET | `/api/packages/subscription?user_id=` | Get active subscription |

**Purchase logic:**
| Scenario | Behavior |
|----------|----------|
| No subscription | Pay full price, new subscription created |
| Same package | Pay full price, expiry extended by 1 day |
| Higher-tier package | Pay price difference only, subscription updated |
| Lower-tier package | Rejected — must wait for current sub to expire |

**Payment method values logged:** `wallet_purchase`, `wallet_upgrade`, `wallet_renewal`

---

### Game Service — port 8083

Serves game data. All public routes require `X-API-Key` and an active subscription. Feature-gated routes additionally check the package's feature flags.

**Public routes (require X-API-Key):**
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/games` | List all games |
| GET | `/api/games/:id` | Get single game |
| GET | `/api/games/:id/history` | Player count history (respects `historical_data_days`) |

**Feature-gated routes (require X-API-Key + package flag):**
| Method | Path | Required flag |
|--------|------|---------------|
| GET | `/api/games/analytics/genre` | `has_genre_analytics` |
| GET | `/api/games/analytics/revenue` | `has_revenue_analytics` |
| GET | `/api/games/analytics/region` | `has_region_breakdown` |
| GET | `/api/games/export` | `has_bulk_export` |
| GET | `/api/games/stream` | `has_realtime_stream` (SSE) |

**Admin routes (require JWT + admin role):**
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/games` | Create game (multipart/form-data) |
| PUT | `/api/games/:id` | Update game |
| DELETE | `/api/games/:id` | Delete game |

**Background processes:**
- **Player simulator** — updates `current_players` every 30 minutes using a sine-wave model (30k–60k range, peaks at 9 PM)
- **History backfill** — on startup, fills 7 days of historical data if the table is empty

---

## Key Data Flows

### 1. User Registration and Login

```
Client → POST /api/users/register
       → bcrypt hash password
       → INSERT INTO users
       ← 201 { user }

Client → POST /api/users/login
       → lookup by email, bcrypt compare
       → sign access token (15m) + refresh token (2d)
       ← 200 { access_token, refresh_token }
```

### 2. Purchasing a Package

```
Client → POST /api/packages/purchase?user_id=1  { package_id: 2 }
       → fetch package price
       → check existing subscription
       → calculate amount (full / diff / extend)
       → POST /internal/users/1/deduct  { amount }   ← calls User Service
       → record subscription + payment
       ← 201 { subscription }
```

### 3. API Key Request to Game Service

```
Client → GET /api/games  [X-API-Key: abc123]
       │
       ├─ GET /internal/keys/abc123/validate         → User Service
       │  ← { user_id, api_key_id, role }
       │
       ├─ GET /api/packages/subscription?user_id=1   → Package Service
       │  ← { package_id: 2, expires_at, ... }
       │
       ├─ GET /api/packages/2                        → Package Service
       │  ← { request_limit, refresh_interval_minutes, has_*, ... }
       │  → store all feature flags in request context
       │
       ├─ GET /internal/usage/count?user_id=1&minutes=5  → User Service
       │  ← { count: 42 }
       │  → if count >= limit: 429 Too Many Requests
       │
       ├─ serve game data  ← 200 { games[] }
       │
       └─ POST /internal/usage/log                   → User Service (async)
```

### 4. Feature-Gated Endpoint (e.g. Genre Analytics)

```
Client → GET /api/games/analytics/genre  [X-API-Key: abc123]
       → AuthAPIKey middleware (same as above)
       → RequireFeature("has_genre_analytics") middleware
         → reads flag from context (set during API key check)
         → if false: 403 { "your current plan does not include this feature" }
         → if true: proceed
       → SELECT genre, COUNT(*), SUM(players), SUM(revenue) FROM games GROUP BY genre
       ← 200 [ { genre, game_count, total_players, ... } ]
```

### 5. Realtime Stream (SSE)

```
Client → GET /api/games/stream  [X-API-Key: abc123]
       → AuthAPIKey + RequireFeature("has_realtime_stream")
       → Content-Type: text/event-stream
       → initial snapshot: data: [{...games}]\n\n
       → every 30 seconds: data: [{...games}]\n\n
       → client disconnect: goroutine exits via context.Done()
```

---

## Package Tiers

| Feature | Standard ($29) | Platinum ($149) | Enterprise ($499) |
|---------|:-:|:-:|:-:|
| Request limit | 100 / 90 min | 5,000 / 5 min | Unlimited |
| Historical data | 60 days | 730 days | 1,825 days |
| Genre analytics | — | Yes | Yes |
| Revenue analytics | — | — | Yes |
| Region breakdown | — | — | Yes |
| Webhook | — | — | Yes |
| Bulk export | — | — | Yes |
| Custom reports | — | — | Yes |
| Dedicated support | — | — | Yes |
| SLA guarantee | — | — | Yes |
| Realtime stream | — | — | Yes |

> Rate limit window resets per `refresh_interval_minutes`. Enterprise (`limit = -1`) skips the check entirely.

---

## Database Schema Summary

### user_db

| Table | Key columns |
|-------|-------------|
| `users` | id, username, email, password_hash, role, balance, is_active |
| `api_keys` | id, user_id, key_hash, is_active |
| `api_usage_logs` | id, user_id, api_key_id, endpoint, method, status_code, created_at |

### package_db

| Table | Key columns |
|-------|-------------|
| `packages` | id, name, price, request_limit, refresh_interval_minutes, historical_data_days, has_* flags |
| `subscriptions` | id, user_id, package_id, status, started_at, expires_at |
| `payments` | id, user_id, subscription_id, amount, payment_method, status |
| `webhook_configs` | id, user_id, url, event_type, is_active, secret |

### game_db

| Table | Key columns |
|-------|-------------|
| `games` | id, name, total_players, current_players, revenue, genre, region, platform, publisher, developer, image_url |
| `game_player_history` | id, game_id, total_players, current_players, recorded_at |
| `genre_player_stats` | id, genre, total_players, current_players, recorded_at |

---

## Tech Stack

| Layer | Stack |
|-------|-------|
| Frontend | React 19, TypeScript, Vite, Tailwind CSS, Recharts, React Router |
| Backend | Go 1.21+, Gin, JWT (golang-jwt), bcrypt |
| Database | PostgreSQL 16 |
| Infrastructure | Docker Compose, Docker named volumes |

---

## Running the Project

### Prerequisites
- Docker Desktop
- Node.js 18+ (frontend only)

### Start backend

```bash
# First time or after code changes:
docker compose up -d --build

# Normal start:
docker compose start
```

### Start frontend

```bash
cd Frontend
npm install
npm run dev
```

Open **http://localhost:5173**

### Rebuild a single service after code change

```bash
docker compose up -d --build game-service
```

### Stop

```bash
docker compose stop          # keeps data
docker compose down -v       # removes containers + volumes
```

---

## Environment Variables (`.env`)

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_USER` | admin | DB username for all 3 databases |
| `POSTGRES_PASSWORD` | admin1234 | DB password |
| `JWT_SECRET` | your_jwt_secret_key | HS256 signing key — **change in production** |
| `USER_SERVICE_PORT` | 8081 | |
| `PACKAGE_SERVICE_PORT` | 8082 | |
| `GAME_SERVICE_PORT` | 8083 | |
| `USER_DB_PORT` | 5431 | Host-side port for user_db |
| `PACKAGE_DB_PORT` | 5432 | Host-side port for package_db |
| `GAME_DB_PORT` | 5437 | Host-side port for game_db |

---

## Notes

- Subscriptions expire after **1 day** in dev mode for rapid testing
- Game images are uploaded, decoded, and re-compressed as JPEG (max 5 MB) on write
- The `game_uploads` Docker volume persists images across container rebuilds
- The player simulator uses a sine-wave model peaking at 21:00, with ±5% noise, bounded 30k–60k
