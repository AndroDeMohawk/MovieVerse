package comment

import (
	"context"
	"errors"

	"github.com/AndroDeMohawk/MovieVerse/internal/transport/grpc/interceptor"
	"github.com/AndroDeMohawk/MovieVerse/internal/usecase"
	moviev1 "github.com/AndroDeMohawk/movie-proto/gen/go/movie"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	moviev1.UnimplementedCommentServiceServer
	comment *usecase.Comment
}

func Register(gRPC *grpc.Server, comment *usecase.Comment) {
	moviev1.RegisterCommentServiceServer(gRPC, &serverAPI{comment: comment})
}

func (s *serverAPI) CreateComment(
	ctx context.Context,
	req *moviev1.CreateCommentRequest,
) (*moviev1.CreateCommentResponse, error) {
	userID, ok := interceptor.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.GetMovieId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid movie_id")
	}

	c, err := s.comment.CreateComment(ctx, req.GetMovieId(), userID, req.GetText())
	if err != nil {
		if errors.Is(err, usecase.ErrEmptyCommentText) {
			return nil, status.Error(codes.InvalidArgument, "comment text cannot be empty")
		}
		return nil, status.Error(codes.Internal, "failed to create comment")
	}

	return &moviev1.CreateCommentResponse{
		Comment: &moviev1.Comment{
			Id:        c.ID,
			MovieId:   c.MovieID,
			UserId:    c.UserID,
			Text:      c.Text,
			CreatedAt: c.CreatedAt.Unix(),
			UpdatedAt: c.UpdatedAt.Unix(),
		},
	}, nil
}

func (s *serverAPI) DeleteComment(
	ctx context.Context,
	req *moviev1.DeleteCommentRequest,
) (*moviev1.DeleteCommentResponse, error) {
	userID, ok := interceptor.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.GetCommentId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid comment_id")
	}

	err := s.comment.DeleteComment(ctx, req.GetCommentId(), userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to delete comment")
	}

	return &moviev1.DeleteCommentResponse{Success: true}, nil
}

func (s *serverAPI) ListMovieComments(
	ctx context.Context,
	req *moviev1.ListMovieCommentsRequest,
) (*moviev1.ListMovieCommentsResponse, error) {
	if req.GetMovieId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid movie_id")
	}

	comments, err := s.comment.ListMovieComments(ctx, req.GetMovieId(), req.GetPage(), req.GetLimit())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch comments")
	}

	pbComments := make([]*moviev1.Comment, 0, len(comments))
	for _, c := range comments {
		pbComments = append(pbComments, &moviev1.Comment{
			Id:        c.ID,
			MovieId:   c.MovieID,
			UserId:    c.UserID,
			Text:      c.Text,
			CreatedAt: c.CreatedAt.Unix(),
			UpdatedAt: c.UpdatedAt.Unix(),
		})
	}

	return &moviev1.ListMovieCommentsResponse{Comments: pbComments}, nil
}
