package movie

import (
	"context"
	"errors"
	"time"

	db "github.com/AndroDeMohawk/MovieVerse/internal/repository/sqlc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/AndroDeMohawk/MovieVerse/internal/transport/grpc/interceptor"
	"github.com/AndroDeMohawk/MovieVerse/internal/usecase"
	moviev1 "github.com/AndroDeMohawk/movie-proto/gen/go/movie"
)

type serverAPI struct {
	moviev1.UnimplementedMovieServiceServer
	movieUsecase *usecase.Movie
}

func Register(gRPC *grpc.Server, movieUsecase *usecase.Movie) {
	moviev1.RegisterMovieServiceServer(gRPC, &serverAPI{
		movieUsecase: movieUsecase,
	})
}

func (s *serverAPI) CreateMovie(
	ctx context.Context,
	req *moviev1.CreateMovieRequest,
) (*moviev1.CreateMovieResponse, error) {
	userID, ok := interceptor.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user is not authenticated")
	}

	if req.GetTitle() == "" {
		return nil, status.Error(codes.InvalidArgument, "title is required")
	}

	input := usecase.CreateMovieInput{
		Title:           req.GetTitle(),
		OriginalTitle:   stringPtr(req.GetOriginalTitle()),
		Description:     stringPtr(req.GetDescription()),
		Director:        stringPtr(req.GetDirector()),
		ReleaseDate:     unixToTimePtr(req.GetReleaseDate()),
		DurationMinutes: int32Ptr(req.GetDurationMinutes()),
		PosterURL:       stringPtr(req.GetPosterUrl()),
		GenreIDs:        req.GetGenreIds(),
	}

	movieID, err := s.movieUsecase.CreateMovie(ctx, userID, input)
	if err != nil {
		if errors.Is(err, usecase.ErrPermissionDenied) {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, "failed to create movie")
	}

	return &moviev1.CreateMovieResponse{
		Id: movieID,
	}, nil
}

func (s *serverAPI) GetMovie(
	ctx context.Context,
	req *moviev1.GetMovieRequest,
) (*moviev1.GetMovieResponse, error) {
	movieID := req.GetId()
	if movieID <= 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid movie_id")
	}

	movie, err := s.movieUsecase.GetMovie(ctx, movieID)
	if err != nil {
		if errors.Is(err, usecase.ErrMovieNotFound) {
			return nil, status.Error(codes.NotFound, "movie not found")
		}
		return nil, status.Error(codes.Internal, "failed to fetch movie")
	}

	return &moviev1.GetMovieResponse{
		Movie: mapGetMovieByIDRowToProto(movie),
	}, nil
}

func (s *serverAPI) ListMovies(
	ctx context.Context,
	req *moviev1.ListMoviesRequest,
) (*moviev1.ListMoviesResponse, error) {
	var genreID *int32
	if req.GenreId != nil {
		g := req.GetGenreId()
		genreID = &g
	}

	movies, err := s.movieUsecase.ListMovies(ctx, genreID, req.GetPage(), req.GetLimit())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list movies")
	}

	protoMovies := make([]*moviev1.Movie, 0, len(movies))
	for _, m := range movies {
		protoMovies = append(protoMovies, mapListMoviesRowToProto(m))
	}

	return &moviev1.ListMoviesResponse{
		Movies: protoMovies,
	}, nil
}

//============================================================================
//MAPPERS & HELPERS
//============================================================================

func mapGetMovieByIDRowToProto(m db.GetMovieByIDRow) *moviev1.Movie {
	return &moviev1.Movie{
		Id:              m.ID,
		Title:           m.Title,
		OriginalTitle:   derefStr(m.OriginalTitle),
		Description:     derefStr(m.Description),
		Director:        derefStr(m.Director),
		DurationMinutes: derefInt32(m.DurationMinutes),
		PosterUrl:       derefStr(m.PosterUrl),
		AverageRating:   m.AverageRating,
		ViewsCount:      m.ViewsCount,
		Genres:          m.Genres,
		CreatedAt:       m.CreatedAt.Unix(),
	}
}

func mapListMoviesRowToProto(m db.ListMoviesRow) *moviev1.Movie {
	return &moviev1.Movie{
		Id:              m.ID,
		Title:           m.Title,
		OriginalTitle:   derefStr(m.OriginalTitle),
		Description:     derefStr(m.Description),
		Director:        derefStr(m.Director),
		DurationMinutes: derefInt32(m.DurationMinutes),
		PosterUrl:       derefStr(m.PosterUrl),
		AverageRating:   m.AverageRating,
		ViewsCount:      m.ViewsCount,
		Genres:          m.Genres,
		CreatedAt:       m.CreatedAt.Unix(),
	}
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func int32Ptr(i int32) *int32 {
	if i <= 0 {
		return nil
	}
	return &i
}

func unixToTimePtr(sec int64) time.Time {
	if sec <= 0 {
		return time.Time{}
	}
	t := time.Unix(sec, 0)
	return t
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}
