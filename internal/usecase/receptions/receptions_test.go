package receptions_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"AvitoPVZ/internal/models"
	"AvitoPVZ/internal/usecase/receptions"
)

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

type ReceptionUseCaseTestSuite struct {
	suite.Suite
	repo *mockReceptionRepo
	uc   *receptions.ReceptionUseCase
}

func (s *ReceptionUseCaseTestSuite) SetupTest() {
	s.repo = new(mockReceptionRepo)
	s.uc = receptions.NewReceptionUseCase(s.repo)
}

func (s *ReceptionUseCaseTestSuite) Test_CreateReception_Success() {
	pvzID := uuid.New()
	expected := models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PvzID:    pvzID,
		Status:   models.StatusInProgress,
	}
	s.repo.On("CreateReceptionTransactional", mock.Anything, pvzID).Return(expected, nil)

	result, err := s.uc.CreateReception(context.Background(), pvzID)

	s.Require().NoError(err)
	s.Equal(expected, result)
	s.repo.AssertExpectations(s.T())
}

func (s *ReceptionUseCaseTestSuite) Test_CreateReception_Error() {
	pvzID := uuid.New()
	expectedErr := errors.New("db error")
	s.repo.On("CreateReceptionTransactional", mock.Anything, pvzID).Return(models.Reception{}, expectedErr)

	result, err := s.uc.CreateReception(context.Background(), pvzID)

	s.Require().Error(err)
	s.Equal(models.Reception{}, result)
	s.Equal(expectedErr, err)
	s.repo.AssertExpectations(s.T())
}

func (s *ReceptionUseCaseTestSuite) Test_CloseLastReception_Success() {
	pvzID := uuid.New().String()
	expected := models.Reception{
		ID:       uuid.New(),
		DateTime: time.Now(),
		PvzID:    uuid.MustParse(pvzID),
		Status:   models.StatusClose,
	}
	s.repo.On("CloseLastReceptionTransactional", mock.Anything, pvzID).Return(expected, nil)

	result, err := s.uc.CloseLastReception(context.Background(), pvzID)

	s.Require().NoError(err)
	s.Equal(expected, result)
	s.repo.AssertExpectations(s.T())
}

func (s *ReceptionUseCaseTestSuite) Test_CloseLastReception_Error() {
	pvzID := uuid.New().String()
	expectedErr := errors.New("no open reception found")
	s.repo.On("CloseLastReceptionTransactional", mock.Anything, pvzID).Return(models.Reception{}, expectedErr)

	result, err := s.uc.CloseLastReception(context.Background(), pvzID)

	s.Require().Error(err)
	s.Equal(models.Reception{}, result)
	s.Equal(expectedErr, err)
	s.repo.AssertExpectations(s.T())
}

func TestReceptionUseCaseTestSuite(t *testing.T) {
	suite.Run(t, new(ReceptionUseCaseTestSuite))
}
