package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	db "github.com/AndroDeMohawk/MovieVerse/internal/repository/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrProfileNotFound = errors.New("profile not found")

type Profile struct {
	q    *db.Queries
	pool *pgxpool.Pool
	log  *slog.Logger
}

func NewProfile(q *db.Queries, pool *pgxpool.Pool, log *slog.Logger) *Profile {
	return &Profile{
		q:    q,
		pool: pool,
		log:  log,
	}
}

func (u *Profile) GetProfileByAuthID(ctx context.Context, authUserID int64) (db.Profile, error) {
	const op = "usecase.profile.GetProfileByAuthID"

	profile, err := u.q.GetProfileByAuthID(ctx, authUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Profile{}, ErrProfileNotFound
		}
		return db.Profile{}, fmt.Errorf("%s: %w", op, err)
	}

	return profile, nil
}

func (u *Profile) CreateProfile(ctx context.Context, authUserID int64, username, avatarURL, bio string) (db.Profile, error) {
	const op = "usecase.profile.CreateProfile"

	profile, err := u.q.CreateProfile(ctx, db.CreateProfileParams{
		AuthUserID: authUserID,
		Username:   username,
		AvatarUrl:  &avatarURL,
		Bio:        &bio,
	})
	if err != nil {
		return db.Profile{}, fmt.Errorf("%s: %w", op, err)
	}

	return profile, nil
}

type UpdateProfileInput struct {
	Username  *string
	AvatarURL *string
	Bio       *string
}

func (u *Profile) UpdateProfile(ctx context.Context, authUserID int64, input UpdateProfileInput) (db.Profile, error) {
	const op = "usecase.profile.UpdateProfile"

	profile, err := u.q.UpdateProfile(ctx, db.UpdateProfileParams{
		AuthUserID: authUserID,
		Username:   input.Username,
		AvatarUrl:  input.AvatarURL,
		Bio:        input.Bio,
	})
	if err != nil {
		return db.Profile{}, fmt.Errorf("%s: %w", op, err)
	}

	return profile, nil
}

func (u *Profile) AddFavorite(ctx context.Context, userID, movieID int64) error {
	const op = "usecase.profile.AddFavorite"

	_, err := u.q.GetMovieByID(ctx, movieID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, ErrMovieNotFound)
	}

	err = u.q.AddFavorite(ctx, db.AddFavoriteParams{
		UserID:  userID,
		MovieID: movieID,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	u.log.Info("movie added to favorites", slog.Int64("user_id", userID), slog.Int64("movie_id", movieID))
	return nil
}

func (u *Profile) RemoveFavorite(ctx context.Context, userID, movieID int64) error {
	const op = "usecase.profile.RemoveFavorite"

	err := u.q.RemoveFavorite(ctx, db.RemoveFavoriteParams{
		UserID:  userID,
		MovieID: movieID,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	u.log.Info("movie removed from favorites", slog.Int64("user_id", userID), slog.Int64("movie_id", movieID))
	return nil
}

func (u *Profile) ListFavorites(ctx context.Context, userID int64, page, limit int32) ([]db.ListFavoritesRow, error) {
	const op = "usecase.profile.ListFavorites"

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	favorites, err := u.q.ListFavorites(ctx, db.ListFavoritesParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return favorites, nil
}

func (u *Profile) IsFavorite(ctx context.Context, userID, movieID int64) (bool, error) {
	const op = "usecase.profile.IsFavorite"

	isFav, err := u.q.IsFavorite(ctx, db.IsFavoriteParams{
		UserID:  userID,
		MovieID: movieID,
	})
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isFav, nil
}
