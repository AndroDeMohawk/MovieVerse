package movie

import (
	"context"

	moviev1 "github.com/AndroDeMohawk/movie-proto/gen/go/movie"
	"google.golang.org/grpc"
)

type serverAPI struct {
	moviev1.UnimplementedMovieServiceServer
	// TODO: Usecase/Service слой
}

func Register(gRPC *grpc.Server) {
	moviev1.RegisterMovieServiceServer(gRPC, &serverAPI{})
}

func (s *serverAPI) GetMovie(
	ctx context.Context,
	req *moviev1.GetMovieRequest,
) (*moviev1.GetMovieResponse, error) {
	// TODO: Интеграция бизнес-логики (Usecase)
	return &moviev1.GetMovieResponse{}, nil
}

func (s *serverAPI) ListMovies(
	ctx context.Context,
	req *moviev1.ListMoviesRequest,
) (*moviev1.ListMoviesResponse, error) {
	// TODO: Интеграция бизнес-логики (Usecase)
	return &moviev1.ListMoviesResponse{}, nil
}

func (s *serverAPI) CreateMovie(
	ctx context.Context,
	req *moviev1.CreateMovieRequest,
) (*moviev1.CreateMovieResponse, error) {
	// TODO: Интеграция бизнес-логики (Usecase)
	return &moviev1.CreateMovieResponse{}, nil
}
