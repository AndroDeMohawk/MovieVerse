-- ============================================================================
-- PROFILES
-- ============================================================================

-- name: CreateProfile :one
INSERT INTO profiles (auth_user_id, username, avatar_url, bio)
VALUES (
           sqlc.arg('auth_user_id'),
           sqlc.arg('username'),
           sqlc.narg('avatar_url'),
           sqlc.narg('bio')
       )
RETURNING id, auth_user_id, username, avatar_url, bio, created_at;

-- name: GetProfileByAuthID :one
SELECT id, auth_user_id, username, avatar_url, bio, created_at
FROM profiles
WHERE auth_user_id = sqlc.arg('auth_user_id')
LIMIT 1;

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
    m.id,
    m.title,
    m.original_title,
    m.description,
    m.director,
    m.release_date,
    m.duration_minutes,
    m.poster_url,
    m.average_rating::float8 AS average_rating,
    m.views_count,
    m.created_at,
    COALESCE(ARRAY_AGG(g.name) FILTER (WHERE g.name IS NOT NULL), '{}')::TEXT[] AS genres
FROM movies m
         LEFT JOIN movie_genres mg ON m.id = mg.movie_id
         LEFT JOIN genres g ON mg.genre_id = g.id
WHERE m.id = sqlc.arg('id')
GROUP BY m.id;

-- name: IncrementMovieViewsBy :exec
UPDATE movies
SET views_count = views_count + $2
WHERE id = $1;

-- name: ListMovies :many
SELECT
    m.id,
    m.title,
    m.original_title,
    m.description,
    m.director,
    m.release_date,
    m.duration_minutes,
    m.poster_url,
    m.average_rating::float8 AS average_rating,
    m.views_count,
    m.created_at,
    COALESCE(ARRAY_AGG(g.name) FILTER (WHERE g.name IS NOT NULL), '{}')::TEXT[] AS genres
FROM movies m
         LEFT JOIN movie_genres mg ON m.id = mg.movie_id
         LEFT JOIN genres g ON mg.genre_id = g.id
WHERE (sqlc.narg('genre_id')::INT IS NULL OR mg.genre_id = sqlc.narg('genre_id'))
GROUP BY m.id
ORDER BY m.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CreateMovie :one
INSERT INTO movies (
    title,
    original_title,
    description,
    director,
    release_date,
    duration_minutes,
    poster_url
) VALUES (
             sqlc.arg('title'),
             sqlc.narg('original_title'),
             sqlc.narg('description'),
             sqlc.narg('director'),
             sqlc.narg('release_date')::date,
             sqlc.narg('duration_minutes'),
             sqlc.narg('poster_url')
         )
RETURNING id;

-- name: AddGenreToMovie :exec
INSERT INTO movie_genres (movie_id, genre_id)
VALUES (sqlc.arg('movie_id'), sqlc.arg('genre_id'))
ON CONFLICT DO NOTHING;

-- name: IncrementMovieViews :exec
UPDATE movies
SET views_count = views_count + 1
WHERE id = sqlc.arg('id');

-- name: UpdateMovieRating :exec
UPDATE movies
SET average_rating = (
    SELECT COALESCE(AVG(score), 0.0)
    FROM ratings
    WHERE movie_id = sqlc.arg('movie_id')
)
WHERE id = sqlc.arg('movie_id');

-- ============================================================================
-- RATINGS
-- ============================================================================

-- name: UpsertRating :exec
INSERT INTO ratings (movie_id, profile_id, score)
VALUES (sqlc.arg('movie_id'), sqlc.arg('profile_id'), sqlc.arg('score'))
ON CONFLICT (movie_id, profile_id)
    DO UPDATE SET score = EXCLUDED.score, created_at = CURRENT_TIMESTAMP;

-- ============================================================================
-- COMMENTS
-- ============================================================================

-- name: CreateComment :one
INSERT INTO comments (movie_id, user_id, text)
VALUES ($1, $2, $3)
RETURNING id, movie_id, user_id, text, created_at, updated_at;

-- name: DeleteComment :exec
DELETE FROM comments
WHERE id = $1 AND user_id = $2;

-- name: ListMovieComments :many
SELECT id, movie_id, user_id, text, created_at, updated_at
FROM comments
WHERE movie_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetCommentByID :one
SELECT id, movie_id, user_id, text, created_at, updated_at
FROM comments
WHERE id = $1;

-- name: AddFavorite :exec
INSERT INTO user_favorites (user_id, movie_id)
VALUES ($1, $2)
ON CONFLICT (user_id, movie_id) DO NOTHING;

-- name: RemoveFavorite :exec
DELETE FROM user_favorites
WHERE user_id = $1 AND movie_id = $2;

-- name: ListFavorites :many
SELECT
    m.id,
    m.title,
    m.original_title,
    m.description,
    m.release_date,
    m.duration_minutes,
    m.poster_url,
    m.average_rating,
    m.views_count,
    uf.created_at AS added_at
FROM user_favorites uf
         JOIN movies m ON m.id = uf.movie_id
WHERE uf.user_id = $1
ORDER BY uf.created_at DESC
LIMIT $2 OFFSET $3;

-- name: IsFavorite :one
SELECT EXISTS (
    SELECT 1
    FROM user_favorites
    WHERE user_id = $1 AND movie_id = $2
);

