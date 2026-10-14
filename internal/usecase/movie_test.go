package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	db "github.com/AndroDeMohawk/MovieVerse/internal/repository/sqlc"
	"github.com/AndroDeMohawk/MovieVerse/internal/usecase"
	"github.com/jackc/pgx/v5/pgtype"
)

type MockMovieRepository struct {
	mock.Mock
}

func (m *MockMovieRepository) GetMovieByID(ctx context.Context, movieID int64) (db.GetMovieByIDRow, error) {
	args := m.Called(ctx, movieID)
	var row db.GetMovieByIDRow
	if args.Get(0) != nil {
		row = args.Get(0).(db.GetMovieByIDRow)
	}
	return row, args.Error(1)
}

type MockMovieQueries struct {
	mock.Mock
}

func (m *MockMovieQueries) CreateMovie(ctx context.Context, arg db.CreateMovieParams) (int64, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMovieQueries) AddGenreToMovie(ctx context.Context, arg db.AddGenreToMovieParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func (m *MockMovieQueries) ListMovies(ctx context.Context, arg db.ListMoviesParams) ([]db.ListMoviesRow, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.ListMoviesRow), args.Error(1)
}

func (m *MockMovieQueries) SearchMovies(ctx context.Context, arg db.SearchMoviesParams) ([]db.SearchMoviesRow, error) {
	args := m.Called(ctx, arg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]db.SearchMoviesRow), args.Error(1)
}

func (m *MockMovieQueries) IncrementMovieViewsBy(ctx context.Context, arg db.IncrementMovieViewsByParams) error {
	args := m.Called(ctx, arg)
	return args.Error(0)
}

func TestMovieUsecase_SearchMovies(t *testing.T) {
	t.Run("success search with genres and pagination", func(t *testing.T) {
		mockRepo := new(MockMovieRepository)
		mockQueries := new(MockMovieQueries)

		expectedMovies := []db.SearchMoviesRow{
			{
				ID:            1,
				Title:         "Довод",
				Director:      "Кристофер Нолан",
				ReleaseDate:   pgtype.Date{Time: time.Date(2020, 8, 26, 0, 0, 0, 0, time.UTC), Valid: true},
				AverageRating: 8.5,
				ViewsCount:    1200,
				CreatedAt:     time.Date(2020, 8, 27, 10, 0, 0, 0, time.UTC),
				Genres:        []string{"Sci-Fi", "Thriller"},
			},
		}

		query := "Нолан"
		genreIDs := []int32{1, 2}
		page := int32(1)
		limit := int32(10)

		mockQueries.On("SearchMovies", mock.Anything, db.SearchMoviesParams{
			Query:    query,
			GenreIds: genreIDs,
			Limit:    limit,
			Offset:   0,
		}).Return(expectedMovies, nil)

		uc := usecase.NewMovie(mockRepo, nil, mockQueries, nil, nil, nil)

		movies, err := uc.SearchMovies(context.Background(), query, page, limit, genreIDs)

		assert.NoError(t, err)
		assert.Equal(t, expectedMovies, movies)

		mockQueries.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error handling", func(t *testing.T) {
		mockRepo := new(MockMovieRepository)
		mockQueries := new(MockMovieQueries)

		query := "test"
		genreIDs := []int32{}
		page := int32(1)
		limit := int32(10)

		mockQueries.On("SearchMovies", mock.Anything, db.SearchMoviesParams{
			Query:    query,
			GenreIds: genreIDs,
			Limit:    limit,
			Offset:   0,
		}).Return(nil, errors.New("database failure"))

		uc := usecase.NewMovie(mockRepo, nil, mockQueries, nil, nil, nil)

		movies, err := uc.SearchMovies(context.Background(), query, page, limit, genreIDs)

		assert.Error(t, err)
		assert.Nil(t, movies)
		assert.ErrorContains(t, err, "database failure")
		assert.ErrorContains(t, err, "usecase.movie.SearchMovies")

		mockQueries.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}
