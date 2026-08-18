-- name: CreateGame :one
INSERT INTO games (
    seed,
    status,
    map_json,
    crisis,
    terminator_json,
    rover_type
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetGame :one
SELECT * FROM games
WHERE id = $1;

-- name: ListGamesBySeed :many
SELECT * FROM games
WHERE seed = $1
ORDER BY created_at DESC;

-- name: UpdateGameStatus :one
UPDATE games
SET
    status = $2,
    colony_score = $3,
    earth_score = $4,
    finished_at = $5
WHERE id = $1
RETURNING *;
