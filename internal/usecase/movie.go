package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/AndroDeMohawk/MovieVerse/internal/client/auth"
	db "github.com/AndroDeMohawk/MovieVerse/internal/repository/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

type Movie struct {
	q          *db.Queries
	pool       *pgxpool.Pool
	authClient *auth.Client
	log        *slog.Logger
}

func NewMovie(q *db.Queries, pool *pgxpool.Pool, authClient *auth.Client, log *slog.Logger) *Movie {
	return &Movie{
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

	movie, err := u.q.GetMovieByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GetMovieByIDRow{}, ErrMovieNotFound
		}
		return db.GetMovieByIDRow{}, fmt.Errorf("%s: %w", op, err)
	}

	// Асинхронно инкрементируем счетчик просмотров
	go func(mID int64) {
		_ = u.q.IncrementMovieViews(context.Background(), mID)
	}(id)

	return movie, nil
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
