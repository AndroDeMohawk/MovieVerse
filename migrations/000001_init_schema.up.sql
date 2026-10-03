CREATE TABLE profiles (
                          id BIGSERIAL PRIMARY KEY,
                          auth_user_id BIGINT UNIQUE NOT NULL,
                          username VARCHAR(100) UNIQUE NOT NULL,
                          avatar_url VARCHAR(255),
                          bio TEXT,
                          created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE movies (
                        id BIGSERIAL PRIMARY KEY,
                        title VARCHAR(255) NOT NULL,
                        original_title VARCHAR(255),
                        description TEXT,
                        director VARCHAR(150),
                        release_date DATE,
                        duration_minutes INT,
                        poster_url VARCHAR(255),
                        average_rating DECIMAL(3, 2) DEFAULT 0.00,
                        views_count BIGINT DEFAULT 0,
                        created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE genres (
                        id SERIAL PRIMARY KEY,
                        name VARCHAR(50) UNIQUE NOT NULL
);

CREATE TABLE movie_genres (
                              movie_id BIGINT REFERENCES movies(id) ON DELETE CASCADE,
                              genre_id INT REFERENCES genres(id) ON DELETE CASCADE,
                              PRIMARY KEY (movie_id, genre_id)
);

CREATE TABLE ratings (
                         id BIGSERIAL PRIMARY KEY,
                         movie_id BIGINT REFERENCES movies(id) ON DELETE CASCADE,
                         profile_id BIGINT REFERENCES profiles(id) ON DELETE CASCADE,
                         score SMALLINT NOT NULL CHECK (score >= 1 AND score <= 10),
                         created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
                         UNIQUE(movie_id, profile_id)
);

CREATE TABLE comments (
                          id BIGSERIAL PRIMARY KEY,
                          movie_id BIGINT REFERENCES movies(id) ON DELETE CASCADE,
                          profile_id BIGINT REFERENCES profiles(id) ON DELETE CASCADE,
                          content TEXT NOT NULL,
                          created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);