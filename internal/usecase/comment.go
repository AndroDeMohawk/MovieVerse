package usecase

import (
	"context"
	"fmt"

	db "github.com/AndroDeMohawk/MovieVerse/internal/repository/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type CommentUsecase struct {
	q *db.Queries
}

func NewCommentUsecase(q *db.Queries) *CommentUsecase {
	return &CommentUsecase{q: q}
}

func (u *CommentUsecase) CreateComment(ctx context.Context, movieID, profileID pgtype.Int8, content string) (db.Comment, error) {
	const op = "CommentUsecase.CreateComment"

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

func (u *CommentUsecase) ListComments(ctx context.Context, movieID pgtype.Int8, page, limit int32) ([]db.ListCommentsByMovieIDRow, error) {
	const op = "usecase.CommentUsecase.ListComments"

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
