package pvz_test

import (
	"AvitoPVZ/internal/models"
	"AvitoPVZ/internal/usecase/pvz"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// --- Мок репозитория ---
type mockPVZRepository struct {
	mock.Mock
}

func (m *mockPVZRepository) Create(ctx context.Context, city models.PVZCity) (models.PVZ, error) {
	args := m.Called(ctx, city)
	return args.Get(0).(models.PVZ), args.Error(1)
}

func (m *mockPVZRepository) GetPVZData(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]models.PVZData, error) {
	args := m.Called(ctx, startDate, endDate, page, limit)
	return args.Get(0).([]models.PVZData), args.Error(1)
}

// --- Тестовая структура ---
type PVZUseCaseSuite struct {
	suite.Suite
	repo *mockPVZRepository
	uc   *pvz.UseCase
}

// --- Setup ---
func (s *PVZUseCaseSuite) SetupTest() {
	s.repo = new(mockPVZRepository)
	s.uc = pvz.NewPVZUseCase(s.repo)
}

// --- Тест: Успешное создание ПВЗ ---
func (s *PVZUseCaseSuite) Test_CreatePVZ_Success() {
	// Arrange
	city := models.CityMoscow
	expected := models.PVZ{
		ID:               uuid.New(),
		City:             city,
		RegistrationDate: time.Now(),
	}
	s.repo.On("Create", mock.Anything, city).Return(expected, nil)

	// Act
	result, err := s.uc.CreatePVZ(context.Background(), city)

	// Assert
	s.Require().NoError(err)
	s.Equal(expected, result)
	s.repo.AssertExpectations(s.T())
}

// --- Тест: Ошибка при создании ПВЗ ---
func (s *PVZUseCaseSuite) Test_CreatePVZ_Error() {
	// Arrange
	city := models.CityKazan
	expectedErr := errors.New("db failure")
	s.repo.On("Create", mock.Anything, city).Return(models.PVZ{}, expectedErr)

	// Act
	result, err := s.uc.CreatePVZ(context.Background(), city)

	// Assert
	s.Require().Error(err)
	s.Equal(models.PVZ{}, result)
	s.ErrorContains(err, "failed to create PVZ")
	s.repo.AssertExpectations(s.T())
}

// --- Тест: Успешное получение PVZData ---
func (s *PVZUseCaseSuite) Test_GetPVZData_Success() {
	// Arrange
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()
	page := 1
	limit := 10
	expectedData := []models.PVZData{
		{
			PVZ: models.PVZ{
				ID:   uuid.New(),
				City: models.CitySamara,
			},
			Receptions: []models.ReceptionWithProducts{},
		},
	}
	s.repo.On("GetPVZData", mock.Anything, &start, &end, page, limit).Return(expectedData, nil)

	// Act
	result, err := s.uc.GetPVZData(context.Background(), &start, &end, page, limit)

	// Assert
	s.Require().NoError(err)
	s.Equal(expectedData, result)
	s.repo.AssertExpectations(s.T())
}

// --- Тест: Ошибка при получении PVZData ---
func (s *PVZUseCaseSuite) Test_GetPVZData_Error() {
	// Arrange
	start := time.Now().Add(-48 * time.Hour)
	end := time.Now()
	page := 2
	limit := 5
	expectedErr := errors.New("repository error")
	s.repo.On("GetPVZData", mock.Anything, &start, &end, page, limit).Return(nil, expectedErr)

	// Act
	result, err := s.uc.GetPVZData(context.Background(), &start, &end, page, limit)

	// Assert
	s.Require().Error(err)
	s.Nil(result)
	s.Equal(expectedErr, err)
	s.repo.AssertExpectations(s.T())
}

// --- Запуск ---
func TestPVZUseCaseSuite(t *testing.T) {
	suite.Run(t, new(PVZUseCaseSuite))
}
