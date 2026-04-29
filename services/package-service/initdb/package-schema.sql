-- ============================================================
-- Package Service - Database Schema
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
('Standard', 'Basic access to game player data. Updated every 1-2 hours with up to 100 requests per day. Historical data up to 2 months.', 29.00, 100, 90, 60, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE),
('Platinum', 'Advanced access with faster updates every 5 minutes, higher request limits, genre analytics, and up to 2 years of historical data.', 149.00, 5000, 5, 730, TRUE, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE),
('Enterprise', 'Full platform access with real-time streaming (1-min refresh), unlimited requests, 5 years of history, revenue analytics, region breakdown, webhooks, bulk export, custom reports, dedicated support, and SLA guarantee.', 499.00, -1, 1, 1825, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE);

CREATE TABLE subscriptions (
    id          SERIAL PRIMARY KEY,
    user_id     INT NOT NULL, -- soft reference to user-service DB
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

CREATE TABLE payments (
    id              SERIAL PRIMARY KEY,
    user_id         INT NOT NULL,
    subscription_id INT NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    amount          DECIMAL(10, 2) NOT NULL,
    currency        VARCHAR(3) NOT NULL DEFAULT 'THB',
    payment_method  VARCHAR(50),
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('success', 'failed', 'pending', 'refunded')),
    transaction_id  VARCHAR(255),
    paid_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_user_id         ON payments (user_id);
CREATE INDEX idx_payments_subscription_id ON payments (subscription_id);
CREATE INDEX idx_payments_status          ON payments (status);

CREATE TABLE webhook_configs (
    id          SERIAL PRIMARY KEY,
    user_id     INT NOT NULL,
    url         VARCHAR(500) NOT NULL,
    event_type  VARCHAR(100) NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    secret      VARCHAR(255),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_webhook_configs_user_id ON webhook_configs (user_id);
