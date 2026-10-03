-- ============================================================================
-- PROFILES
-- ============================================================================

-- name: CreateProfile :one
INSERT INTO profiles (auth_user_id, username, avatar_url, bio)
VALUES ($1, $2, $3, $4)
RETURNING id, auth_user_id, username, avatar_url, bio, created_at;

-- name: GetProfileByAuthID :one
SELECT id, auth_user_id, username, avatar_url, bio, created_at
FROM profiles
WHERE auth_user_id = $1 LIMIT 1;

-- name: UpdateProfile :one
UPDATE profiles
SET
    username = COALESCE(sqlc.narg('username'), username),
    avatar_url = COALESCE(sqlc.narg('avatar_url'), avatar_url),
    bio = COALESCE(sqlc.narg('bio'), bio)
WHERE auth_user_id = sqlc.arg('auth_user_id')
RETURNING id, auth_user_id, username, avatar_url, bio, created_at;

-- ============================================================================
-- MOVIES & GENRES
-- ============================================================================

-- name: GetMovieByID :one
SELECT
    m.id, m.title, m.original_title, m.description, m.director,
    m.release_date, m.duration_minutes, m.poster_url, m.average_rating,
    m.views_count, m.created_at,
    COALESCE(ARRAY_AGG(g.name) FILTER (WHERE g.name IS NOT NULL), '{}')::TEXT[] AS genres
FROM movies m
         LEFT JOIN movie_genres mg ON m.id = mg.movie_id
         LEFT JOIN genres g ON mg.genre_id = g.id
WHERE m.id = $1
GROUP BY m.id;

-- name: ListMovies :many
SELECT
    m.id, m.title, m.original_title, m.description, m.director,
    m.release_date, m.duration_minutes, m.poster_url, m.average_rating,
    m.views_count, m.created_at,
    COALESCE(ARRAY_AGG(g.name) FILTER (WHERE g.name IS NOT NULL), '{}')::TEXT[] AS genres
FROM movies m
         LEFT JOIN movie_genres mg ON m.id = mg.movie_id
         LEFT JOIN genres g ON mg.genre_id = g.id
WHERE ($1::INT IS NULL OR mg.genre_id = $1)
GROUP BY m.id
ORDER BY m.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateMovie :one
INSERT INTO movies (title, original_title, description, director, release_date, duration_minutes, poster_url)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;

-- name: AddGenreToMovie :exec
INSERT INTO movie_genres (movie_id, genre_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: IncrementMovieViews :exec
UPDATE movies
SET views_count = views_count + 1
WHERE id = $1;

-- name: UpdateMovieRating :exec
UPDATE movies
SET average_rating = (
    SELECT COALESCE(AVG(score), 0.0)
    FROM ratings
    WHERE movie_id = $1
)
WHERE id = $1;

-- ============================================================================
-- RATINGS
-- ============================================================================

-- name: UpsertRating :exec
INSERT INTO ratings (movie_id, profile_id, score)
VALUES ($1, $2, $3)
ON CONFLICT (movie_id, profile_id)
    DO UPDATE SET score = EXCLUDED.score, created_at = CURRENT_TIMESTAMP;

-- ============================================================================
-- COMMENTS
-- ============================================================================

-- name: CreateComment :one
INSERT INTO comments (movie_id, profile_id, content)
VALUES ($1, $2, $3)
RETURNING id, movie_id, profile_id, content, created_at;

-- name: ListCommentsByMovieID :many
SELECT
    c.id, c.movie_id, c.profile_id, c.content, c.created_at,
    p.username, p.avatar_url
FROM comments c
         JOIN profiles p ON c.profile_id = p.id
WHERE c.movie_id = $1
ORDER BY c.created_at DESC
LIMIT $2 OFFSET $3;