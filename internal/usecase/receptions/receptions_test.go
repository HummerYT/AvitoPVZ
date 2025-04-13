package receptions_test

import (
	"AvitoPVZ/internal/models"
	"AvitoPVZ/internal/usecase/receptions"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// --- Мок для репозитория ---
type mockReceptionRepo struct {
	mock.Mock
}

func (m *mockReceptionRepo) CreateReceptionTransactional(ctx context.Context, pvzID uuid.UUID) (models.Reception, error) {
	args := m.Called(ctx, pvzID)
	return args.Get(0).(models.Reception), args.Error(1)
}

func (m *mockReceptionRepo) CloseLastReceptionTransactional(ctx context.Context, pvzID string) (models.Reception, error) {
	args := m.Called(ctx, pvzID)
	return args.Get(0).(models.Reception), args.Error(1)
}

// --- Структура тестов ---
type ReceptionUseCaseTestSuite struct {
	suite.Suite
	repo *mockReceptionRepo
	uc   *receptions.ReceptionUseCase
}

// --- Setup ---
func (s *ReceptionUseCaseTestSuite) SetupTest() {
	s.repo = new(mockReceptionRepo)
	s.uc = receptions.NewReceptionUseCase(s.repo)
}

// --- Тест: успешное создание приёма ---
func (s *ReceptionUseCaseTestSuite) Test_CreateReception_Success() {
	// Arrange
	pvzID := uuid.New()
	expected := models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PvzID:    pvzID,
		Status:   models.ReceptionStatusOpened,
	}
	s.repo.On("CreateReceptionTransactional", mock.Anything, pvzID).Return(expected, nil)

	// Act
	result, err := s.uc.CreateReception(context.Background(), pvzID)

	// Assert
	s.Require().NoError(err)
	s.Equal(expected, result)
	s.repo.AssertExpectations(s.T())
}

// --- Тест: ошибка при создании приёма ---
func (s *ReceptionUseCaseTestSuite) Test_CreateReception_Error() {
	// Arrange
	pvzID := uuid.New()
	expectedErr := errors.New("db error")
	s.repo.On("CreateReceptionTransactional", mock.Anything, pvzID).Return(models.Reception{}, expectedErr)

	// Act
	result, err := s.uc.CreateReception(context.Background(), pvzID)

	// Assert
	s.Require().Error(err)
	s.Equal(models.Reception{}, result)
	s.Equal(expectedErr, err)
	s.repo.AssertExpectations(s.T())
}

// --- Тест: успешное закрытие приёма ---
func (s *ReceptionUseCaseTestSuite) Test_CloseLastReception_Success() {
	// Arrange
	pvzID := uuid.New().String()
	expected := models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PvzID:    uuid.MustParse(pvzID),
		Status:   models.ReceptionStatusClosed,
	}
	s.repo.On("CloseLastReceptionTransactional", mock.Anything, pvzID).Return(expected, nil)

	// Act
	result, err := s.uc.CloseLastReception(context.Background(), pvzID)

	// Assert
	s.Require().NoError(err)
	s.Equal(expected, result)
	s.repo.AssertExpectations(s.T())
}

// --- Тест: ошибка при закрытии приёма ---
func (s *ReceptionUseCaseTestSuite) Test_CloseLastReception_Error() {
	// Arrange
	pvzID := uuid.New().String()
	expectedErr := errors.New("no open reception found")
	s.repo.On("CloseLastReceptionTransactional", mock.Anything, pvzID).Return(models.Reception{}, expectedErr)

	// Act
	result, err := s.uc.CloseLastReception(context.Background(), pvzID)

	// Assert
	s.Require().Error(err)
	s.Equal(models.Reception{}, result)
	s.Equal(expectedErr, err)
	s.repo.AssertExpectations(s.T())
}

// --- Запуск тестов ---
func TestReceptionUseCaseTestSuite(t *testing.T) {
	suite.Run(t, new(ReceptionUseCaseTestSuite))
}
