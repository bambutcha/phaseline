-- name: UpsertGhost :one
INSERT INTO ghost_runs (
    seed,
    game_id,
    colony_score,
    earth_score,
    replay_json,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, now()
)
ON CONFLICT (seed) DO UPDATE SET
    game_id = EXCLUDED.game_id,
    colony_score = EXCLUDED.colony_score,
    earth_score = EXCLUDED.earth_score,
    replay_json = EXCLUDED.replay_json,
    updated_at = now()
WHERE ghost_runs.colony_score <= EXCLUDED.colony_score
RETURNING *;

-- name: GetGhostBySeed :one
SELECT * FROM ghost_runs
WHERE seed = $1;
