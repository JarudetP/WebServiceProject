# 🎮 Game Data Platform

Subscription-based REST API platform for real-time game analytics. Users register, purchase tiered packages, generate API keys, and query live game data within rate-limited access tiers.

---

## 🏗️ Architecture (Microservices)

| Service            | Port  | Responsibility                                   |
|--------------------|-------|--------------------------------------------------|
| **User Service**   | 8081  | Registration, Login, API Key management, Balance |
| **Package Service**| 8082  | Packages, Subscriptions, Purchase logic          |
| **Game Service**   | 8083  | Game CRUD, Live player simulator, History        |
| **User DB**        | 5431  | PostgreSQL for user-service                      |
| **Package DB**     | 5432  | PostgreSQL for package-service                   |
| **Game DB**        | 5437  | PostgreSQL for game-service                      |

---

## ⚙️ Tech Stack

| Layer    | Stack                                               |
|----------|-----------------------------------------------------|
| Backend  | Go 1.21+, Gin, PostgreSQL 16, JWT, bcrypt           |
| Frontend | React 19, TypeScript, Vite, Tailwind CSS, Recharts  |
| Infra    | Docker Compose                                      |

---

## 🚀 Getting Started

### Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) installed and running
- [Node.js 18+](https://nodejs.org/) (for the frontend)
- [Go 1.21+](https://go.dev/) (only if you want to run backend services locally without Docker)

---

### Step 1 – Clone the repository

```bash
git clone https://github.com/JarudetP/WebServiceProject.git
cd WebServiceProject
```

---

### Step 2 – Configure Environment Variables

All configuration is managed via the root `.env` file. A default one is already included:

```
POSTGRES_USER=admin
POSTGRES_PASSWORD=admin1234
JWT_SECRET=your_jwt_secret_key

USER_SERVICE_PORT=8081
PACKAGE_SERVICE_PORT=8082
GAME_SERVICE_PORT=8083

USER_DB_PORT=5431
PACKAGE_DB_PORT=5432
GAME_DB_PORT=5437
```

> ⚠️ **Change `JWT_SECRET` before deploying to production!**

---

### Step 3 – Start the Backend (Docker)

This command starts all 3 microservices + their 3 PostgreSQL databases:

```bash
# First time (or after changing Go code):
docker-compose up -d --build

# Normal start (no code changes):
docker-compose start
```

Wait ~20 seconds for databases to initialize. Verify all services are running:

```bash
docker-compose ps
```

All services should show status `running`.

---

### Step 4 – Start the Frontend

```bash
cd Frontend
npm install     # Only needed first time
npm run dev
```

Open your browser at: **http://localhost:5173**

---

## 🧪 API Testing (Postman)

A ready-to-use Postman collection is at `services/postman.json`.

Import it into Postman and use these environment variables:
| Variable        | Value                    |
|-----------------|--------------------------|
| `user_url`      | `http://localhost:8081`  |
| `package_url`   | `http://localhost:8082`  |
| `game_url`      | `http://localhost:8083`  |

---

## 🔑 How to Use the Platform

1. **Register** a new user via `POST /api/users/register`
2. **Login** via `POST /api/users/login` → get your JWT token
3. **Top up balance** via `POST /api/users/:id/topup`
4. **Purchase a package** via `POST /api/packages/purchase`
5. **Generate an API key** via `POST /api/users/:id/keys`
6. **Query game data** via `GET /api/games` using the `X-API-Key` header

---

## 🛑 Stopping the Server

```bash
# Stop all services (keeps data):
docker-compose stop

# Stop and remove all containers + data volumes:
docker-compose down -v
```

---

## 📝 Notes

- Subscriptions expire in **1 day** in dev mode for rapid testing
- The player count simulator runs every **30 minutes** (Game Service)
- Enterprise package (`package_id = 3`) has **unlimited** API requests (`limit = -1`)
- Game images are auto-compressed to JPEG (max 5MB) on upload
- Database data persists between restarts via Docker volumes
