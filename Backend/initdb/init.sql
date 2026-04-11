-- ============================================================
-- Game Data Platform - Database Schema
-- PostgreSQL 16
-- ============================================================

-- Enable UUID extension (optional, for future use)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- 1. GAMES
-- ============================================================
CREATE TABLE games (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(255) NOT NULL,
    total_players   BIGINT NOT NULL DEFAULT 0,
    current_players BIGINT NOT NULL DEFAULT 0,
    revenue         DECIMAL(15, 2) NOT NULL DEFAULT 0.00,
    genre           VARCHAR(100) NOT NULL,
    region          VARCHAR(100) NOT NULL,
    platform        VARCHAR(100) NOT NULL,
    publisher       VARCHAR(255) NOT NULL,
    developer       VARCHAR(255) NOT NULL,
    image_url       VARCHAR(500) NOT NULL DEFAULT '',
    timestamp       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_games_genre    ON games (genre);
CREATE INDEX idx_games_region   ON games (region);
CREATE INDEX idx_games_platform ON games (platform);
INSERT INTO games (name, total_players, current_players, revenue, genre, region, platform, publisher, developer, timestamp) VALUES
('Epic Adventure', 5000000, 32450, 15000000.00, 'Action RPG', 'North America', 'PC', 'Epic Games', 'Epic Games', NOW()),
('Space Odyssey', 3000000, 24120, 9000000.00, 'Sci-Fi', 'Europe', 'Console', 'Galactic Studios', 'Galactic Studios', NOW()),
('Mystic Quest', 2000000, 48500, 5000000.00, 'Fantasy', 'Asia', 'Mobile', 'Mystic Inc.', 'Mystic Inc.', NOW());
-- ============================================================
-- 2. PACKAGES
-- ============================================================
CREATE TABLE packages (
    id                        SERIAL PRIMARY KEY,
    name                      VARCHAR(50) NOT NULL UNIQUE,
    description               TEXT,
    price                     DECIMAL(10, 2) NOT NULL,
    request_limit             INT NOT NULL,            -- -1 = unlimited
    refresh_interval_minutes  INT NOT NULL,
    historical_data_days      INT NOT NULL,
    has_genre_analytics       BOOLEAN NOT NULL DEFAULT FALSE,
    has_revenue_analytics     BOOLEAN NOT NULL DEFAULT FALSE,
    has_region_breakdown      BOOLEAN NOT NULL DEFAULT FALSE,
    has_webhook               BOOLEAN NOT NULL DEFAULT FALSE,
    has_bulk_export           BOOLEAN NOT NULL DEFAULT FALSE,
    has_custom_reports        BOOLEAN NOT NULL DEFAULT FALSE,
    has_dedicated_support     BOOLEAN NOT NULL DEFAULT FALSE,
    has_sla_guarantee         BOOLEAN NOT NULL DEFAULT FALSE,
    has_realtime_stream       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Seed 3 packages
INSERT INTO packages (
    name, description, price,
    request_limit, refresh_interval_minutes, historical_data_days,
    has_genre_analytics, has_revenue_analytics, has_region_breakdown,
    has_webhook, has_bulk_export, has_custom_reports,
    has_dedicated_support, has_sla_guarantee, has_realtime_stream
) VALUES
(
    'Standard',
    'Basic access to game player data. Updated every 1-2 hours with up to 100 requests per day. Historical data up to 2 months.',
    29.00,
    100, 90, 60,
    FALSE, FALSE, FALSE,
    FALSE, FALSE, FALSE,
    FALSE, FALSE, FALSE
),
(
    'Platinum',
    'Advanced access with faster updates every 5 minutes, higher request limits, genre analytics, and up to 2 years of historical data.',
    149.00,
    5000, 5, 730,
    TRUE, FALSE, FALSE,
    FALSE, FALSE, FALSE,
    FALSE, FALSE, FALSE
),
(
    'Enterprise',
    'Full platform access with real-time streaming (1-min refresh), unlimited requests, 5 years of history, revenue analytics, region breakdown, webhooks, bulk export, custom reports, dedicated support, and SLA guarantee.',
    499.00,
    -1, 1, 1825,
    TRUE, TRUE, TRUE,
    TRUE, TRUE, TRUE,
    TRUE, TRUE, TRUE
);

-- ============================================================
-- 3. USERS
-- ============================================================
CREATE TABLE users (
    id            SERIAL PRIMARY KEY,
    username      VARCHAR(100) NOT NULL UNIQUE,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name     VARCHAR(255),
    company       VARCHAR(255),
    role          VARCHAR(20) NOT NULL DEFAULT 'user',
    balance       DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users (email);

-- To create an admin user:
-- 1. Register via API: POST /api/users/register
-- 2. Then run: UPDATE users SET role = 'admin' WHERE username = 'testuser';

-- ============================================================
-- 4. SUBSCRIPTIONS
-- ============================================================
CREATE TABLE subscriptions (
    id          SERIAL PRIMARY KEY,
    user_id     INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    package_id  INT NOT NULL REFERENCES packages(id) ON DELETE RESTRICT,
    status      VARCHAR(20) NOT NULL DEFAULT 'active'
                CHECK (status IN ('active', 'expired', 'cancelled')),
    started_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_user_id    ON subscriptions (user_id);
CREATE INDEX idx_subscriptions_package_id ON subscriptions (package_id);
CREATE INDEX idx_subscriptions_status     ON subscriptions (status);

-- ============================================================
-- 5. API KEYS
-- ============================================================
CREATE TABLE api_keys (
    id         SERIAL PRIMARY KEY,
    user_id    INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key        VARCHAR(255) NOT NULL UNIQUE,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

CREATE INDEX idx_api_keys_user_id ON api_keys (user_id);
CREATE INDEX idx_api_keys_key     ON api_keys (key);

-- ============================================================
-- 6. API USAGE LOGS
-- ============================================================
CREATE TABLE api_usage_logs (
    id               BIGSERIAL PRIMARY KEY,
    user_id          INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    api_key_id       INT NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    endpoint         VARCHAR(255) NOT NULL,
    method           VARCHAR(10) NOT NULL,
    status_code      INT,
    response_time_ms INT,
    requested_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_usage_user_id      ON api_usage_logs (user_id);
CREATE INDEX idx_api_usage_requested_at ON api_usage_logs (requested_at);
CREATE INDEX idx_api_usage_api_key_id   ON api_usage_logs (api_key_id);

-- ============================================================
-- 7. GAME PLAYER HISTORY
-- ============================================================
CREATE TABLE game_player_history (
    id              BIGSERIAL PRIMARY KEY,
    game_id         INT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    total_players   BIGINT NOT NULL DEFAULT 0,
    current_players BIGINT NOT NULL DEFAULT 0,
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_game_player_history_game_recorded
    ON game_player_history (game_id, recorded_at);

-- Seed 7 days of historical data (every 30 mins)
INSERT INTO game_player_history (game_id, total_players, current_players, recorded_at)
SELECT 
    g.id, 
    g.total_players, 
    (20000 + (random() * 30001))::INT, 
    gs.recorded_at
FROM games g
CROSS JOIN generate_series(
    NOW() - INTERVAL '7 days', 
    NOW(), 
    INTERVAL '30 minutes'
) AS gs(recorded_at);

-- ============================================================
-- 8. GENRE PLAYER STATS
-- ============================================================
CREATE TABLE genre_player_stats (
    id              BIGSERIAL PRIMARY KEY,
    genre           VARCHAR(100) NOT NULL,
    total_players   BIGINT NOT NULL DEFAULT 0,
    current_players BIGINT NOT NULL DEFAULT 0,
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_genre_stats_genre_recorded
    ON genre_player_stats (genre, recorded_at);

-- ============================================================
-- 9. PAYMENTS
-- ============================================================
CREATE TABLE payments (
    id              SERIAL PRIMARY KEY,
    user_id         INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscription_id INT NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    amount          DECIMAL(10, 2) NOT NULL,
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    payment_method  VARCHAR(50),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('success', 'failed', 'pending', 'refunded')),
    transaction_id  VARCHAR(255),
    paid_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_user_id         ON payments (user_id);
CREATE INDEX idx_payments_subscription_id ON payments (subscription_id);
CREATE INDEX idx_payments_status          ON payments (status);

-- ============================================================
-- 10. WEBHOOK CONFIGS (Enterprise only)
-- ============================================================
CREATE TABLE webhook_configs (
    id          SERIAL PRIMARY KEY,
    user_id     INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url         VARCHAR(500) NOT NULL,
    event_type  VARCHAR(100) NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    secret      VARCHAR(255),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_configs_user_id ON webhook_configs (user_id);

-- ============================================================
-- AUTO-UPDATE updated_at TRIGGER
-- ============================================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_games_updated_at
    BEFORE UPDATE ON games
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================
-- DONE
-- ============================================================
