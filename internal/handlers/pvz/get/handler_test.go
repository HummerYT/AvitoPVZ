package get_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"AvitoPVZ/internal/handlers/get"
	"AvitoPVZ/internal/models"
)

// --- Мок ---
type mockPVZUseCase struct {
	mock.Mock
}

func (m *mockPVZUseCase) GetPVZData(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]models.PVZData, error) {
	args := m.Called(ctx, startDate, endDate, page, limit)
	return args.Get(0).([]models.PVZData), args.Error(1)
}

// --- Тестовая структура ---
type GetPVZDataSuite struct {
	suite.Suite
	app  *fiber.App
	mock *mockPVZUseCase
}

// --- Setup ---
func (s *GetPVZDataSuite) SetupTest() {
	s.app = fiber.New()
	s.mock = new(mockPVZUseCase)
	handler := get.NewPVZDataHandler(s.mock)

	s.app.Get("/pvz", func(c *fiber.Ctx) error {
		c.Locals("Role", s.T().Context().Value("role"))
		return handler.GetPVZData(c)
	})
}

// --- Успешный кейс ---
func (s *GetPVZDataSuite) TestGetPVZData_Success() {
	// Arrange
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()
	page := 2
	limit := 5

	expectedData := []models.PVZData{
		{
			PVZ: models.PVZ{ID: "pvz1", Address: "Test Address"},
			Receptions: []models.ReceptionData{
				{Reception: models.Reception{ID: "rec1"}, Products: []models.Product{}},
			},
		},
	}

	s.mock.On("GetPVZData", mock.Anything, &start, &end, page, limit).Return(expectedData, nil)

	req := httptest.NewRequest("GET", fmt.Sprintf("/pvz?startDate=%s&endDate=%s&page=%d&limit=%d",
		start.Format(time.RFC3339), end.Format(time.RFC3339), page, limit), nil)
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleAdmin))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(200, resp.StatusCode)

	var body []map[string]any
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))
	s.Len(body, 1)
	s.mock.AssertExpectations(s.T())
}

// --- Ошибка доступа ---
func (s *GetPVZDataSuite) TestGetPVZData_Forbidden() {
	// Arrange
	req := httptest.NewRequest("GET", "/pvz", nil)
	req = req.WithContext(context.WithValue(req.Context(), "role", "unauthorized"))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(403, resp.StatusCode)
}

// --- Неверный формат даты ---
func (s *GetPVZDataSuite) TestGetPVZData_InvalidDateFormat() {
	// Arrange
	req := httptest.NewRequest("GET", "/pvz?startDate=invalid-date", nil)
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleAdmin))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(400, resp.StatusCode)
}

// --- Ошибка от UseCase ---
func (s *GetPVZDataSuite) TestGetPVZData_UseCaseError() {
	// Arrange
	start := time.Now()
	end := time.Now().Add(2 * time.Hour)

	s.mock.On("GetPVZData", mock.Anything, &start, &end, 1, 10).
		Return(nil, errors.New("something went wrong"))

	req := httptest.NewRequest("GET", fmt.Sprintf("/pvz?startDate=%s&endDate=%s",
		start.Format(time.RFC3339), end.Format(time.RFC3339)), nil)
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleAdmin))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(400, resp.StatusCode)
	s.mock.AssertExpectations(s.T())
}

// --- Запуск ---
func TestGetPVZDataSuite(t *testing.T) {
	suite.Run(t, new(GetPVZDataSuite))
}
