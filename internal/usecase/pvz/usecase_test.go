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

type PVZUseCaseSuite struct {
	suite.Suite
	repo *mockPVZRepository
	uc   *pvz.UseCase
}

func (s *PVZUseCaseSuite) SetupTest() {
	s.repo = new(mockPVZRepository)
	s.uc = pvz.NewPVZUseCase(s.repo)
}

func (s *PVZUseCaseSuite) Test_CreatePVZ_Success() {
	city := models.CityMoscow
	expected := models.PVZ{
		ID:               uuid.New().String(),
		City:             string(city),
		RegistrationDate: time.Now(),
	}
	s.repo.On("Create", mock.Anything, city).Return(expected, nil)

	result, err := s.uc.CreatePVZ(context.Background(), city)

	s.Require().NoError(err)
	s.Equal(expected, result)
	s.repo.AssertExpectations(s.T())
}

func (s *PVZUseCaseSuite) Test_CreatePVZ_Error() {
	city := models.CityKazan
	expectedErr := errors.New("db failure")
	s.repo.On("Create", mock.Anything, city).Return(models.PVZ{}, expectedErr)

	result, err := s.uc.CreatePVZ(context.Background(), city)

	s.Require().Error(err)
	s.Equal(models.PVZ{}, result)
	s.ErrorContains(err, "failed to create PVZ")
	s.repo.AssertExpectations(s.T())
}

func (s *PVZUseCaseSuite) Test_GetPVZData_Success() {
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()
	page := 1
	limit := 10
	expectedData := []models.PVZData{
		{
			PVZ: models.PVZ{
				ID:   uuid.New().String(),
				City: string(models.CityMoscow),
			},
			Receptions: []models.ReceptionData{},
		},
	}
	s.repo.On("GetPVZData", mock.Anything, &start, &end, page, limit).Return(expectedData, nil)

	result, err := s.uc.GetPVZData(context.Background(), &start, &end, page, limit)

	s.Require().NoError(err)
	s.Equal(expectedData, result)
	s.repo.AssertExpectations(s.T())
}

func TestPVZUseCaseSuite(t *testing.T) {
	suite.Run(t, new(PVZUseCaseSuite))
}
