CREATE TABLE IF NOT EXISTS profiles (
                          id BIGSERIAL PRIMARY KEY,
                          auth_user_id BIGINT UNIQUE NOT NULL,
                          username VARCHAR(100) UNIQUE NOT NULL,
                          avatar_url VARCHAR(255),
                          bio TEXT,
                          created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS movies (
                        id BIGSERIAL PRIMARY KEY,
                        title VARCHAR(255) NOT NULL,
                        original_title VARCHAR(255),
                        description TEXT,
                        director VARCHAR(150),
                        release_date DATE,
                        duration_minutes INT,
                        poster_url VARCHAR(255),
                        average_rating DECIMAL(3, 2) NOT NULL DEFAULT 0.00,
                        views_count BIGINT NOT NULL DEFAULT 0,
                        created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS genres (
                        id SERIAL PRIMARY KEY,
                        name VARCHAR(50) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS movie_genres (
                              movie_id BIGINT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
                              genre_id INT NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
                              PRIMARY KEY (movie_id, genre_id)
);

CREATE TABLE IF NOT EXISTS ratings (
                         id BIGSERIAL PRIMARY KEY,
                         movie_id BIGINT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
                         profile_id BIGINT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
                         score SMALLINT NOT NULL CHECK (score >= 1 AND score <= 10),
                         created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
                         UNIQUE(movie_id, profile_id)
);

CREATE TABLE IF NOT EXISTS user_favorites (
                                              user_id BIGINT NOT NULL,
                                              movie_id BIGINT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
                                              created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                              PRIMARY KEY (user_id, movie_id)
);
CREATE INDEX IF NOT EXISTS idx_user_favorites_user_created
    ON user_favorites (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS comments (
                                        id BIGSERIAL PRIMARY KEY,
                                        movie_id BIGINT NOT NULL REFERENCES movies(id) ON DELETE CASCADE,
                                        user_id BIGINT NOT NULL,
                                        text TEXT NOT NULL,
                                        created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                        updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_comments_movie_created
    ON comments (movie_id, created_at DESC);