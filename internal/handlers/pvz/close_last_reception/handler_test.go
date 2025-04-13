package close_last_reception_test

import (
	"context"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"AvitoPVZ/internal/handlers/close_last_reception"
	"AvitoPVZ/internal/models"
)

// --- Мок UseCase ---
type mockReceptionUseCase struct {
	mock.Mock
}

func (m *mockReceptionUseCase) CloseLastReception(ctx context.Context, pvzID string) (models.Reception, error) {
	args := m.Called(ctx, pvzID)
	return args.Get(0).(models.Reception), args.Error(1)
}

// --- Тестовый сьют ---
type CloseLastReceptionSuite struct {
	suite.Suite
	app  *fiber.App
	mock *mockReceptionUseCase
}

// --- Setup ---
func (s *CloseLastReceptionSuite) SetupTest() {
	s.app = fiber.New()
	s.mock = new(mockReceptionUseCase)

	handler := close_last_reception.NewReceptionHandler(s.mock)

	s.app.Post("/reception/:pvzId/close", func(c *fiber.Ctx) error {
		c.Locals("Role", s.T().Context().Value("role"))
		return handler.CloseLastReception(c)
	})
}

// --- Успешное закрытие приёмки ---
func (s *CloseLastReceptionSuite) TestCloseLastReception_Success() {
	// Arrange
	pvzID := uuid.New().String()

	expectedReception := models.Reception{
		ID:       1001,
		DateTime: time.Now(),
		PvzID:    pvzID,
		Status:   "CLOSED",
	}

	s.mock.On("CloseLastReception", mock.Anything, pvzID).
		Return(expectedReception, nil)

	req := httptest.NewRequest("POST", fmt.Sprintf("/reception/%s/close", pvzID), nil)
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleEmployee))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(fiber.StatusOK, resp.StatusCode)
	s.mock.AssertExpectations(s.T())
}

// --- Ошибка: не сотрудник PVZ ---
func (s *CloseLastReceptionSuite) TestCloseLastReception_Forbidden() {
	// Arrange
	pvzID := uuid.New().String()
	req := httptest.NewRequest("POST", fmt.Sprintf("/reception/%s/close", pvzID), nil)
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleAdmin))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(fiber.StatusForbidden, resp.StatusCode)
}

// --- Ошибка: невалидный UUID ---
func (s *CloseLastReceptionSuite) TestCloseLastReception_InvalidUUID() {
	// Arrange
	invalidPvzID := "not-a-uuid"
	req := httptest.NewRequest("POST", fmt.Sprintf("/reception/%s/close", invalidPvzID), nil)
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleEmployee))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(fiber.StatusBadRequest, resp.StatusCode)
}

// --- Ошибка: usecase вернул ошибку ---
func (s *CloseLastReceptionSuite) TestCloseLastReception_UseCaseError() {
	// Arrange
	pvzID := uuid.New().String()
	s.mock.On("CloseLastReception", mock.Anything, pvzID).
		Return(models.Reception{}, errors.New("something went wrong"))

	req := httptest.NewRequest("POST", fmt.Sprintf("/reception/%s/close", pvzID), nil)
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleEmployee))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(fiber.StatusBadRequest, resp.StatusCode)
	s.mock.AssertExpectations(s.T())
}

// --- Запуск тестов ---
func TestCloseLastReceptionSuite(t *testing.T) {
	suite.Run(t, new(CloseLastReceptionSuite))
}
