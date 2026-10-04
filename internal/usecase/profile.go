package usecase

import (
	"context"
	"errors"
	"fmt"

	db "github.com/AndroDeMohawk/MovieVerse/internal/repository/sqlc"
	"github.com/jackc/pgx/v5"
)

var ErrProfileNotFound = errors.New("profile not found")

type Profile struct {
	q *db.Queries
}

func NewProfile(q *db.Queries) *Profile {
	return &Profile{q: q}
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
