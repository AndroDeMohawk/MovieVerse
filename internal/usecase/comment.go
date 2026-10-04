package usecase

import (
	"context"
	"fmt"

	db "github.com/AndroDeMohawk/MovieVerse/internal/repository/sqlc"
)

type Comment struct {
	q *db.Queries
}

func NewComment(q *db.Queries) *Comment {
	return &Comment{q: q}
}

func (u *Comment) CreateComment(ctx context.Context, movieID, profileID int64, content string) (db.Comment, error) {
	const op = "usecase.comment.CreateComment"

	comment, err := u.q.CreateComment(ctx, db.CreateCommentParams{
		MovieID:   movieID,
		ProfileID: profileID,
		Content:   content,
	})
	if err != nil {
		return db.Comment{}, fmt.Errorf("%s: %w", op, err)
	}

	return comment, nil
}

func (u *Comment) ListComments(ctx context.Context, movieID int64, page, limit int32) ([]db.ListCommentsByMovieIDRow, error) {
	const op = "usecase.comment.ListComments"

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	comments, err := u.q.ListCommentsByMovieID(ctx, db.ListCommentsByMovieIDParams{
		MovieID: movieID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return comments, nil
}
