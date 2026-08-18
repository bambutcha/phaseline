-- name: InsertEvent :one
INSERT INTO game_events (
    game_id,
    t,
    kind,
    payload_json
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: InsertEvents :copyfrom
INSERT INTO game_events (
    game_id,
    t,
    kind,
    payload_json
) VALUES (
    $1, $2, $3, $4
);

-- name: ListEventsByGame :many
SELECT * FROM game_events
WHERE game_id = $1
ORDER BY t ASC, id ASC;
