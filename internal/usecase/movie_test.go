package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	db "github.com/AndroDeMohawk/MovieVerse/internal/repository/sqlc"
	"github.com/AndroDeMohawk/MovieVerse/internal/usecase"
)

// 1. Создаем мок для репозитория фильмов
type MockMovieRepository struct {
	mock.Mock
}

func (m *MockMovieRepository) SearchMovies(ctx context.Context, query string, genreIDs []int32, limit, offset int32) ([]db.Movie, int64, error) {
	args := m.Called(ctx, query, genreIDs, limit, offset)

	var movies []db.Movie
	if args.Get(0) != nil {
		movies = args.Get(0).([]db.Movie)
	}

	return movies, args.Get(1).(int64), args.Error(2)
}

func (m *MockMovieRepository) CreateMovie(ctx context.Context, input db.CreateMovieParams, userID int64) (int64, error) {
	args := m.Called(ctx, input, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMovieRepository) GetMovieByID(ctx context.Context, movieID int64) (db.GetMovieByIDRow, error) {
	args := m.Called(ctx, movieID)
	var row db.GetMovieByIDRow
	if args.Get(0) != nil {
		row = args.Get(0).(db.GetMovieByIDRow)
	}
	return row, args.Error(1)
}

// 2. Тестируем UseCase поиска
func TestMovieUsecase_SearchMovies(t *testing.T) {
	t.Run("success search with genres and pagination", func(t *testing.T) {
		// Arrange (Подготовка)
		mockRepo := new(MockMovieRepository)

		expectedMovies := []db.Movie{
			{ID: 1, Title: "Довод", Director: "Кристофер Нолан"},
		}
		expectedTotal := int64(1)

		query := "Нолан"
		genreIDs := []int32{1, 2}
		page := int32(1)
		limit := int32(10)
		expectedOffset := int32(0) // (page - 1) * limit

		// Ожидаем вызов метода репозитория с конкретными аргументами
		mockRepo.On("SearchMovies", mock.Anything, query, genreIDs, limit, expectedOffset).
			Return(expectedMovies, expectedTotal, nil)

		uc := usecase.NewMovie(mockRepo, nil, nil, nil, nil, nil)

		// Act (Выполнение)
		movies, err := uc.SearchMovies(context.Background(), query, page, limit, genreIDs)

		// Assert (Проверки)
		assert.NoError(t, err)
		assert.Equal(t, expectedMovies, movies)

		// Проверяем, что все замоканные вызовы реально произошли
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository error handling", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockMovieRepository)

		query := "test"
		genreIDs := []int32{}
		page := int32(1)
		limit := int32(10)
		expectedOffset := int32(0)

		mockRepo.On("SearchMovies", mock.Anything, query, genreIDs, limit, expectedOffset).
			Return(nil, int64(0), errors.New("database failure"))

		uc := usecase.NewMovie(mockRepo, nil, nil, nil, nil, nil)

		// Act
		movies, err := uc.SearchMovies(context.Background(), query, page, limit, genreIDs)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, movies)
		assert.Equal(t, "database failure", err.Error())

		mockRepo.AssertExpectations(t)
	})
}
