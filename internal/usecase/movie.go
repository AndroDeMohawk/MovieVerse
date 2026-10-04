package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/AndroDeMohawk/MovieVerse/internal/client/auth"
	db "github.com/AndroDeMohawk/MovieVerse/internal/repository/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrPermissionDenied = errors.New("permission denied")
	ErrMovieNotFound    = errors.New("movie not found")
)

type MovieUsecase struct {
	q          *db.Queries
	pool       *pgxpool.Pool
	authClient *auth.Client
	log        *slog.Logger
}

func NewMovieUsecase(q *db.Queries, pool *pgxpool.Pool, authClient *auth.Client, log *slog.Logger) *MovieUsecase {
	return &MovieUsecase{
		q:          q,
		pool:       pool,
		authClient: authClient,
		log:        log,
	}

}

type CreateMovieInput struct {
	Title           string
	OriginalTitle   pgtype.Text
	Description     pgtype.Text
	Director        pgtype.Text
	DurationMinutes pgtype.Int4
	PosterURL       pgtype.Text
	GenreIDs        []int32
}

func (u *MovieUsecase) CreateMovie(ctx context.Context, userID int64, input CreateMovieInput) (int64, error) {
	const op = "MovieUsecase.CreateMovie"
	isAdmin, err := u.authClient.IsAdmin(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("%s, %w", op, err)

	}
	if !isAdmin {
		return 0, ErrPermissionDenied
	}

	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s:TX begin failed , %w", op, err)
	}
	defer tx.Rollback(ctx)
	qtx := u.q.WithTx(tx)

	movieID, err := qtx.CreateMovie(ctx, db.CreateMovieParams{
		Title:           input.Title,
		OriginalTitle:   input.OriginalTitle,
		Description:     input.Description,
		Director:        input.Director,
		DurationMinutes: input.DurationMinutes,
		PosterUrl:       input.PosterURL,
	})
	if err != nil {
		return 0, fmt.Errorf("%s: create movie failed, %w", op, err)
	}
	for _, genreID := range input.GenreIDs {
		if err := qtx.AddGenreToMovie(ctx, db.AddGenreToMovieParams{
			MovieID: movieID,
			GenreID: genreID,
		}); err != nil {
			return 0, fmt.Errorf("%s: add genre failed, %w", op, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("%s: commit tx failed, %w", op, err)
	}
	u.log.Info("movie created successfully", slog.Int64("movie_id", movieID), slog.Int64("by_user", userID))
	return movieID, nil
}

func (u *MovieUsecase) GetMovieByID(ctx context.Context, id int64) (db.GetMovieByIDRow, error) {
	const op = "MovieUsecase.GetMovieByID"
	movie, err := u.q.GetMovieByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.GetMovieByIDRow{}, ErrMovieNotFound
		}
		return db.GetMovieByIDRow{}, err
	}
	go func(mID int64) {
		_ = u.q.IncrementMovieViews(context.Background(), mID)
	}(id)
	return movie, nil
}
func (u *MovieUsecase) ListMovies(ctx context.Context, genreID int32, page, limit int32) ([]db.ListMoviesRow, error) {
	const op = "MovieUsecase.ListMovies"

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	movies, err := u.q.ListMovies(ctx, db.ListMoviesParams{
		Column1: genreID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return movies, nil
}
