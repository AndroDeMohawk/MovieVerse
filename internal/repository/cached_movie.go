package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	db "github.com/AndroDeMohawk/MovieVerse/internal/repository/sqlc"
	"github.com/redis/go-redis/v9"
)

const movieCacheTTL = 10 * time.Minute

type PostgresMovieRepository interface {
	GetMovieByID(ctx context.Context, id int64) (db.GetMovieByIDRow, error)
}

type CachedMovieRepository struct {
	pgRepo PostgresMovieRepository
	cache  *redis.Client // Принимаем стандартный *redis.Client из go-redis/v9
	log    *slog.Logger
}

func NewCachedMovieRepository(pgRepo PostgresMovieRepository, cache *redis.Client, log *slog.Logger) *CachedMovieRepository {
	return &CachedMovieRepository{
		pgRepo: pgRepo,
		cache:  cache,
		log:    log,
	}
}

func (r *CachedMovieRepository) GetMovieByID(ctx context.Context, id int64) (db.GetMovieByIDRow, error) {
	cacheKey := fmt.Sprintf("movie:%d", id)

	// 1. Попытка вычитать из кэша Redis
	val, err := r.cache.Get(ctx, cacheKey).Bytes()
	if err == nil {
		var movie db.GetMovieByIDRow
		if unmarshalErr := json.Unmarshal(val, &movie); unmarshalErr == nil {
			r.log.Debug("movie cache hit", slog.Int64("id", id))
			return movie, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		r.log.Warn("failed to get key from redis", slog.String("error", err.Error()))
	}

	r.log.Debug("movie cache miss", slog.Int64("id", id))

	// 2. Запрос в PostgreSQL
	dbMovie, err := r.pgRepo.GetMovieByID(ctx, id)
	if err != nil {
		return db.GetMovieByIDRow{}, err
	}

	// 3. Сохраняем результат в Redis
	bytes, err := json.Marshal(dbMovie)
	if err == nil {
		if setErr := r.cache.Set(ctx, cacheKey, bytes, movieCacheTTL).Err(); setErr != nil {
			r.log.Warn("failed to set movie cache", slog.String("error", setErr.Error()))
		}
	}

	return dbMovie, nil
}

func (r *CachedMovieRepository) InvalidateMovieCache(ctx context.Context, id int64) error {
	cacheKey := fmt.Sprintf("movie:%d", id)
	if err := r.cache.Del(ctx, cacheKey).Err(); err != nil {
		r.log.Warn("failed to invalidate movie cache", slog.String("error", err.Error()))
		return err
	}
	return nil
}
