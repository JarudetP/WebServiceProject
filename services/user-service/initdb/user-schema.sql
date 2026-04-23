-- ============================================================
-- User Service - Database Schema
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

CREATE TABLE api_keys (
    id         SERIAL PRIMARY KEY,
    user_id    INT NOT NULL, -- soft link to users(id) in this DB
    key        VARCHAR(255) NOT NULL UNIQUE,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ
);

CREATE INDEX idx_api_keys_user_id ON api_keys (user_id);
CREATE INDEX idx_api_keys_key     ON api_keys (key);

CREATE TABLE api_usage_logs (
    id               BIGSERIAL PRIMARY KEY,
    user_id          INT NOT NULL,
    api_key_id       INT NOT NULL,
    endpoint         VARCHAR(255) NOT NULL,
    method           VARCHAR(10) NOT NULL,
    status_code      INT,
    response_time_ms INT,
    requested_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_usage_user_id      ON api_usage_logs (user_id);
CREATE INDEX idx_api_usage_requested_at ON api_usage_logs (requested_at);
CREATE INDEX idx_api_usage_api_key_id   ON api_usage_logs (api_key_id);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
