package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"sync"
	"time"

	"github.com/AndroDeMohawk/MovieVerse/internal/client/auth"
	db "github.com/AndroDeMohawk/MovieVerse/internal/repository/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const pendingKey = "movie:views:pending"

var (
	ErrPermissionDenied = errors.New("permission denied: admin role required")
	ErrMovieNotFound    = errors.New("movie not found")
)

type CreateMovieInput struct {
	Title           string
	OriginalTitle   *string
	Description     *string
	Director        *string
	ReleaseDate     time.Time
	DurationMinutes *int32
	PosterURL       *string
	GenreIDs        []int32
}

type MovieRepository interface {
	GetMovieByID(ctx context.Context, id int64) (db.GetMovieByIDRow, error)
}

type Movie struct {
	repo       MovieRepository
	rdb        *redis.Client
	q          *db.Queries
	pool       *pgxpool.Pool
	authClient *auth.Client
	log        *slog.Logger
}

func NewMovie(
	repo MovieRepository,
	rdb *redis.Client,
	q *db.Queries,
	pool *pgxpool.Pool,
	authClient *auth.Client,
	log *slog.Logger,
) *Movie {
	return &Movie{
		repo:       repo,
		rdb:        rdb,
		q:          q,
		pool:       pool,
		authClient: authClient,
		log:        log,
	}
}

func (u *Movie) CreateMovie(ctx context.Context, userID int64, input CreateMovieInput) (int64, error) {
	const op = "usecase.movie.CreateMovie"

	isAdmin, err := u.authClient.IsAdmin(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("%s: auth check failed: %w", op, err)
	}
	if !isAdmin {
		return 0, ErrPermissionDenied
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: begin tx failed: %w", op, err)
	}
	defer tx.Rollback(ctx)

	qtx := u.q.WithTx(tx)

	movieID, err := qtx.CreateMovie(ctx, db.CreateMovieParams{
		Title:         input.Title,
		OriginalTitle: input.OriginalTitle,
		Description:   input.Description,
		Director:      input.Director,
		ReleaseDate: pgtype.Date{
			Time:  input.ReleaseDate,
			Valid: true,
		},
		DurationMinutes: input.DurationMinutes,
		PosterUrl:       input.PosterURL,
	})
	if err != nil {
		return 0, fmt.Errorf("%s: create movie failed: %w", op, err)
	}

	for _, genreID := range input.GenreIDs {
		if err := qtx.AddGenreToMovie(ctx, db.AddGenreToMovieParams{
			MovieID: movieID,
			GenreID: genreID,
		}); err != nil {
			return 0, fmt.Errorf("%s: add genre failed: %w", op, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("%s: commit tx failed: %w", op, err)
	}

	u.log.Info("movie created successfully", slog.Int64("movie_id", movieID), slog.Int64("by_user", userID))
	return movieID, nil
}

func (u *Movie) GetMovie(ctx context.Context, id int64) (db.GetMovieByIDRow, error) {
	const op = "usecase.movie.GetMovie"

	// 1. Кэширование полностью прозрачно выполняется внутри u.repo (CachedMovieRepository)
	movie, err := u.repo.GetMovieByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GetMovieByIDRow{}, ErrMovieNotFound
		}
		return db.GetMovieByIDRow{}, fmt.Errorf("%s: %w", op, err)
	}

	// 2. Асинхронно инкрементируем счетчик просмотров в Redis
	go u.incrementViews(id)

	return movie, nil
}

func (u *Movie) incrementViews(movieID int64) {
	ctx := context.Background()
	viewsKey := fmt.Sprintf("movie:views:%d", movieID)

	pipe := u.rdb.Pipeline()
	pipe.Incr(ctx, viewsKey)
	pipe.SAdd(ctx, pendingKey, movieID)

	if _, err := pipe.Exec(ctx); err != nil {
		u.log.Warn("failed to increment views in redis", slog.Int64("movie_id", movieID), slog.Any("error", err))
	}
}

func (u *Movie) ListMovies(ctx context.Context, genreID *int32, page, limit int32) ([]db.ListMoviesRow, error) {
	const op = "usecase.movie.ListMovies"

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	movies, err := u.q.ListMovies(ctx, db.ListMoviesParams{
		GenreID: genreID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return movies, nil
}

func (u *Movie) StartViewsSync(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			u.flushViewsToDB(ctx)
		case <-ctx.Done():
			u.log.Info("stopping views sync, executing final flush")
			u.flushViewsToDB(context.Background())
			return
		}
	}
}

func (u *Movie) flushViewsToDB(ctx context.Context) {
	movieIDs, err := u.rdb.SMembers(ctx, pendingKey).Result()
	if err != nil {
		u.log.Error("failed to get pending movie IDs from redis", slog.Any("error", err))
		return
	}
	if len(movieIDs) == 0 {
		return
	}

	const numWorkers = 10
	jobs := make(chan string, len(movieIDs))
	var wg sync.WaitGroup

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idStr := range jobs {
				u.workerFlushMovieViews(ctx, idStr)
			}
		}()
	}

	for _, idStr := range movieIDs {
		jobs <- idStr
	}
	close(jobs)

	wg.Wait()
	u.log.Debug("flushed movie views to database", slog.Int("count", len(movieIDs)))
}

func (u *Movie) workerFlushMovieViews(ctx context.Context, idStr string) {
	movieID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		u.rdb.SRem(ctx, pendingKey, idStr)
		return
	}

	viewsKey := fmt.Sprintf("movie:views:%d", movieID)

	delta, err := u.rdb.GetDel(ctx, viewsKey).Int64()
	if err != nil {
		if err == redis.Nil {
			u.rdb.SRem(ctx, pendingKey, idStr)
			return
		}
		u.log.Error("failed to get views from redis", slog.Int64("movie_id", movieID), slog.Any("error", err))
		return
	}

	if delta <= 0 {
		u.rdb.SRem(ctx, pendingKey, idStr)
		return
	}

	if err := u.rdb.SRem(ctx, pendingKey, idStr).Err(); err != nil {
		u.log.Error("failed to remove from pending", slog.Int64("movie_id", movieID), slog.Any("error", err))
	}

	err = u.q.IncrementMovieViewsBy(ctx, db.IncrementMovieViewsByParams{
		ID:         movieID,
		ViewsCount: delta,
	})
	if err != nil {
		u.log.Error("failed to flush views to postgres, rolling back to redis",
			slog.Int64("movie_id", movieID),
			slog.Int64("delta", delta),
			slog.Any("error", err),
		)
		rbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		rbPipe := u.rdb.Pipeline()
		rbPipe.IncrBy(rbCtx, viewsKey, delta)
		rbPipe.SAdd(rbCtx, pendingKey, idStr)
		if _, execErr := rbPipe.Exec(rbCtx); execErr != nil {
			u.log.Error("CRITICAL: failed to rollback views to redis",
				slog.Int64("movie_id", movieID),
				slog.Int64("delta", delta),
				slog.Any("error", execErr),
			)
		}
		return
	}

	if err := u.rdb.Del(ctx, fmt.Sprintf("movie:%d", movieID)).Err(); err != nil {
		u.log.Error("failed to invalidate movie cache", slog.Int64("movie_id", movieID), slog.Any("error", err))
	}
}
