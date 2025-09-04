-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: CreateFeedFollow :one
WITH inserted AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
    )
    RETURNING *
)
SELECT inserted.id, inserted.created_at, inserted.updated_at, users.id AS "user ID", users.name AS "username", feeds.id AS "feed ID",  feeds.name AS "feed name"
FROM inserted
JOIN users ON inserted.user_id = users.id
JOIN feeds ON inserted.feed_id = feeds.id
ORDER BY inserted.created_at DESC
LIMIT 1;


-- name: GetUserByName :one
SELECT * FROM users WHERE name = $1;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetAllUsers :many
SELECT * FROM users;

-- name: GetAllFeeds :many
SELECT * FROM feeds;

-- name: GetFeedByUrl :one
SELECT * FROM feeds WHERE url = $1;

-- name: GetFeedFollowsForUser :many
SELECT feed_follows.id, users.name AS "username", feeds.name AS "feed name", feeds.url AS "feed url"
FROM feed_follows
JOIN users ON feed_follows.user_id = users.id
JOIN feeds ON feed_follows.feed_id = feeds.id
WHERE users.id = $1
ORDER BY feed_follows.created_at DESC;

-- name: ResetUsers :exec
DELETE FROM users;