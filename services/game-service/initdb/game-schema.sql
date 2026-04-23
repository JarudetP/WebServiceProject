-- ============================================================
-- Game Service - Database Schema
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

CREATE TABLE game_player_history (
    id              BIGSERIAL PRIMARY KEY,
    game_id         INT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    total_players   BIGINT NOT NULL DEFAULT 0,
    current_players BIGINT NOT NULL DEFAULT 0,
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_game_player_history_game_recorded
    ON game_player_history (game_id, recorded_at);

CREATE TABLE genre_player_stats (
    id              BIGSERIAL PRIMARY KEY,
    genre           VARCHAR(100) NOT NULL,
    total_players   BIGINT NOT NULL DEFAULT 0,
    current_players BIGINT NOT NULL DEFAULT 0,
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_genre_stats_genre_recorded
    ON genre_player_stats (genre, recorded_at);

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
