package profile

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
	moviev1.UnimplementedProfileServiceServer
	profile *usecase.Profile
}

func Register(gRPC *grpc.Server, profile *usecase.Profile) {
	moviev1.RegisterProfileServiceServer(gRPC, &serverAPI{profile: profile})
}

func (s *serverAPI) AddFavorite(
	ctx context.Context,
	req *moviev1.AddFavoriteRequest,
) (*moviev1.AddFavoriteResponse, error) {
	userID, ok := interceptor.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.GetMovieId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid movie_id")
	}

	err := s.profile.AddFavorite(ctx, userID, req.GetMovieId())
	if err != nil {
		if errors.Is(err, usecase.ErrMovieNotFound) {
			return nil, status.Error(codes.NotFound, "movie not found")
		}
		return nil, status.Error(codes.Internal, "failed to add favorite")
	}

	return &moviev1.AddFavoriteResponse{Success: true}, nil
}

func (s *serverAPI) RemoveFavorite(
	ctx context.Context,
	req *moviev1.RemoveFavoriteRequest,
) (*moviev1.RemoveFavoriteResponse, error) {
	userID, ok := interceptor.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.GetMovieId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid movie_id")
	}

	err := s.profile.RemoveFavorite(ctx, userID, req.GetMovieId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to remove favorite")
	}

	return &moviev1.RemoveFavoriteResponse{Success: true}, nil
}

func (s *serverAPI) ListFavorites(
	ctx context.Context,
	req *moviev1.ListFavoritesRequest,
) (*moviev1.ListFavoritesResponse, error) {
	userID, ok := interceptor.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	favorites, err := s.profile.ListFavorites(ctx, userID, req.GetPage(), req.GetLimit())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch favorites")
	}

	pbMovies := make([]*moviev1.FavoriteMovie, 0, len(favorites))
	for _, f := range favorites {
		ratingVal, _ := f.AverageRating.Float64Value()

		pbMovies = append(pbMovies, &moviev1.FavoriteMovie{
			Id:              f.ID,
			Title:           f.Title,
			OriginalTitle:   f.OriginalTitle,
			Description:     f.Description,
			ReleaseDate:     f.ReleaseDate.Time.Unix(),
			DurationMinutes: f.DurationMinutes,
			PosterUrl:       f.PosterUrl,
			Rating:          ratingVal.Float64,
			ViewsCount:      f.ViewsCount,
			AddedAt:         f.AddedAt.Unix(),
		})
	}

	return &moviev1.ListFavoritesResponse{Movies: pbMovies}, nil
}

func (s *serverAPI) IsFavorite(
	ctx context.Context,
	req *moviev1.IsFavoriteRequest,
) (*moviev1.IsFavoriteResponse, error) {
	userID, ok := interceptor.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if req.GetMovieId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid movie_id")
	}

	isFav, err := s.profile.IsFavorite(ctx, userID, req.GetMovieId())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to check favorite status")
	}

	return &moviev1.IsFavoriteResponse{IsFavorite: isFav}, nil
}
