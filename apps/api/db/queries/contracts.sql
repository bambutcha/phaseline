-- name: InsertContract :one
INSERT INTO contracts (
    game_id,
    payload_json,
    status
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: ListContractsByGame :many
SELECT * FROM contracts
WHERE game_id = $1
ORDER BY id;

-- name: UpdateContractStatus :one
UPDATE contracts
SET
    status = $2,
    accepted_at = $3,
    delivered_at = $4
WHERE id = $1
RETURNING *;
