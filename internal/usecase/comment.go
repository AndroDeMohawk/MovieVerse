package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/AndroDeMohawk/MovieVerse/internal/infrastructure/kafka"
	db "github.com/AndroDeMohawk/MovieVerse/internal/repository/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrCommentNotFound  = errors.New("comment not found")
	ErrEmptyCommentText = errors.New("comment text cannot be empty")
)

type Comment struct {
	q            *db.Queries
	pool         *pgxpool.Pool
	producer     *kafka.Producer
	commentTopic string
	log          *slog.Logger
}

func NewComment(
	q *db.Queries,
	pool *pgxpool.Pool,
	producer *kafka.Producer,
	commentTopic string,
	log *slog.Logger,
) *Comment {
	return &Comment{
		q:            q,
		pool:         pool,
		producer:     producer,
		commentTopic: commentTopic,
		log:          log,
	}
}

func (u *Comment) CreateComment(
	ctx context.Context,
	movieID, userID int64,
	text string,
) (db.Comment, error) {
	const op = "usecase.comment.CreateComment"

	text = strings.TrimSpace(text)
	if text == "" {
		return db.Comment{}, ErrEmptyCommentText
	}

	comment, err := u.q.CreateComment(ctx, db.CreateCommentParams{
		MovieID: movieID,
		UserID:  userID,
		Text:    text,
	})
	if err != nil {
		return db.Comment{}, fmt.Errorf("%s: %w", op, err)
	}

	event := CommentCreatedEvent{
		CommentID: comment.ID,
		MovieID:   comment.MovieID,
		UserID:    comment.UserID,
		CreatedAt: comment.CreatedAt.Unix(),
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		u.log.Error("failed to marshal comment event", slog.String("error", err.Error()))
		return comment, nil
	}

	key := []byte(strconv.FormatInt(movieID, 10))
	if err := u.producer.SendMessage(ctx, u.commentTopic, key, eventBytes); err != nil {
		u.log.Error("failed to publish comment event to kafka",
			slog.Int64("comment_id", comment.ID),
			slog.String("error", err.Error()),
		)
	}

	u.log.Info("comment created successfully",
		slog.Int64("comment_id", comment.ID),
		slog.Int64("movie_id", movieID),
		slog.Int64("user_id", userID),
	)

	return comment, nil
}

func (u *Comment) ListMovieComments(
	ctx context.Context,
	movieID int64,
	page, limit int32,
) ([]db.Comment, error) {
	const op = "usecase.comment.ListMovieComments"

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	comments, err := u.q.ListMovieComments(ctx, db.ListMovieCommentsParams{
		MovieID: movieID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return comments, nil
}

func (u *Comment) DeleteComment(ctx context.Context, commentID, userID int64) error {
	const op = "usecase.comment.DeleteComment"

	err := u.q.DeleteComment(ctx, db.DeleteCommentParams{
		ID:     commentID,
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	u.log.Info("comment deleted", slog.Int64("comment_id", commentID), slog.Int64("user_id", userID))
	return nil
}
