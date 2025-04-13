package receptions_test

import (
	"AvitoPVZ/internal/handlers/receptions"
	"AvitoPVZ/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"net/http/httptest"
	"testing"
	"time"
)

// --- Мок UseCase ---
type mockReceptionUseCase struct {
	mock.Mock
}

func (m *mockReceptionUseCase) CreateReception(ctx context.Context, pvzID uuid.UUID) (models.Reception, error) {
	args := m.Called(ctx, pvzID)
	return args.Get(0).(models.Reception), args.Error(1)
}

// --- Тестовая структура ---
type ReceptionHandlerSuite struct {
	suite.Suite
	app  *fiber.App
	mock *mockReceptionUseCase
}

// --- Setup ---
func (s *ReceptionHandlerSuite) SetupTest() {
	s.app = fiber.New()
	s.mock = new(mockReceptionUseCase)
	handler := receptions.NewReceptionHandler(s.mock)

	s.app.Post("/reception", func(c *fiber.Ctx) error {
		c.Locals("Role", s.T().Context().Value("role"))
		return handler.CreateReception(c)
	})
}

// --- Успешный кейс ---
func (s *ReceptionHandlerSuite) Test_CreateReception_Success() {
	// Arrange
	pvzID := uuid.New()
	expected := models.Reception{
		ID:       "rec-1",
		PvzID:    pvzID,
		DateTime: time.Now(),
		Status:   "open",
	}
	s.mock.On("CreateReception", mock.Anything, pvzID).Return(expected, nil)

	body := map[string]string{"pvzId": pvzID.String()}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/reception", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleEmployee))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(201, resp.StatusCode)

	var result map[string]interface{}
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&result))
	s.Equal(expected.ID, result["id"])
	s.Equal(expected.PvzID.String(), result["pvzId"])
	s.Equal(expected.Status, result["status"])

	s.mock.AssertExpectations(s.T())
}

// --- Ошибка: Нет доступа ---
func (s *ReceptionHandlerSuite) Test_CreateReception_Forbidden() {
	// Arrange
	req := httptest.NewRequest("POST", "/reception", nil)
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleClient))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(403, resp.StatusCode)
}

// --- Ошибка: Невалидный JSON ---
func (s *ReceptionHandlerSuite) Test_CreateReception_BadJSON() {
	// Arrange
	req := httptest.NewRequest("POST", "/reception", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleEmployee))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(400, resp.StatusCode)
}

// --- Ошибка: Пустой pvzId ---
func (s *ReceptionHandlerSuite) Test_CreateReception_EmptyPvzID() {
	// Arrange
	body := map[string]string{"pvzId": ""}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/reception", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleEmployee))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(400, resp.StatusCode)
}

// --- Ошибка: Невалидный UUID ---
func (s *ReceptionHandlerSuite) Test_CreateReception_InvalidUUID() {
	// Arrange
	body := map[string]string{"pvzId": "not-a-uuid"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/reception", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleEmployee))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(400, resp.StatusCode)
}

// --- Ошибка от UseCase ---
func (s *ReceptionHandlerSuite) Test_CreateReception_UseCaseError() {
	// Arrange
	pvzID := uuid.New()
	s.mock.On("CreateReception", mock.Anything, pvzID).Return(models.Reception{}, errors.New("fail"))

	body := map[string]string{"pvzId": pvzID.String()}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/reception", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleEmployee))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(400, resp.StatusCode)
	s.mock.AssertExpectations(s.T())
}

// --- Запуск ---
func TestReceptionHandlerSuite(t *testing.T) {
	suite.Run(t, new(ReceptionHandlerSuite))
}
