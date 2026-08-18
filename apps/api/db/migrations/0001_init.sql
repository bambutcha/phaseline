-- +goose Up
CREATE TABLE games (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    seed TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('lobby', 'active', 'finished')),
    map_json JSONB NOT NULL,
    crisis TEXT NOT NULL,
    terminator_json JSONB NOT NULL,
    colony_score INT NOT NULL DEFAULT 0,
    earth_score INT NOT NULL DEFAULT 0,
    rover_type TEXT NOT NULL CHECK (rover_type IN ('swift', 'hauler')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ
);

CREATE TABLE contracts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    game_id UUID NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    payload_json JSONB NOT NULL,
    status TEXT NOT NULL,
    accepted_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ
);

CREATE TABLE game_events (
    id BIGSERIAL PRIMARY KEY,
    game_id UUID NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    t DOUBLE PRECISION NOT NULL,
    kind TEXT NOT NULL,
    payload_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ghost_runs (
    seed TEXT PRIMARY KEY,
    game_id UUID NOT NULL REFERENCES games (id) ON DELETE CASCADE,
    colony_score INT NOT NULL,
    earth_score INT NOT NULL,
    replay_json JSONB NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_games_seed ON games (seed);
CREATE INDEX idx_contracts_game ON contracts (game_id);
CREATE INDEX idx_events_game ON game_events (game_id, t);

-- +goose Down
DROP TABLE IF EXISTS ghost_runs;
DROP TABLE IF EXISTS game_events;
DROP TABLE IF EXISTS contracts;
DROP TABLE IF EXISTS games;
