package usecase

import (
	"context"
	"errors"
	"fmt"

	db "github.com/AndroDeMohawk/MovieVerse/internal/repository/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var ErrProfileNotFound = errors.New("profile not found")

type ProfileUsecase struct {
	q *db.Queries
}

func NewProfileUsecase(q *db.Queries) *ProfileUsecase {
	return &ProfileUsecase{q: q}
}

func (u *ProfileUsecase) GetProfileByAuthID(ctx context.Context, authUserID int64) (db.Profile, error) {
	const op = "ProfileUsecase.GetProfileByAuthID"

	profile, err := u.q.GetProfileByAuthID(ctx, authUserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.Profile{}, ErrProfileNotFound
		}
		return db.Profile{}, fmt.Errorf("%s: %w", op, err)
	}

	return profile, nil
}

func (u *ProfileUsecase) CreateProfile(ctx context.Context, authUserID int64, username string, avatarURL, bio pgtype.Text) (db.Profile, error) {
	const op = "ProfileUsecase.CreateProfile"

	profile, err := u.q.CreateProfile(ctx, db.CreateProfileParams{
		AuthUserID: authUserID,
		Username:   username,
		AvatarUrl:  avatarURL,
		Bio:        bio,
	})
	if err != nil {
		return db.Profile{}, fmt.Errorf("%s: %w", op, err)
	}

	return profile, nil
}

type UpdateProfileInput struct {
	Username  pgtype.Text
	AvatarURL pgtype.Text
	Bio       pgtype.Text
}

func (u *ProfileUsecase) UpdateProfile(ctx context.Context, authUserID int64, input UpdateProfileInput) (db.Profile, error) {
	const op = "ProfileUsecase.UpdateProfile"

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
