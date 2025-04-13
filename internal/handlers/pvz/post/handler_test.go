package post_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"AvitoPVZ/internal/handlers/post"
	"AvitoPVZ/internal/models"
)

// --- Мок ---
type mockPVZUseCase struct {
	mock.Mock
}

func (m *mockPVZUseCase) CreatePVZ(ctx context.Context, city models.PVZCity) (models.PVZ, error) {
	args := m.Called(ctx, city)
	return args.Get(0).(models.PVZ), args.Error(1)
}

// --- Тестовая структура ---
type CreatePVZSuite struct {
	suite.Suite
	app  *fiber.App
	mock *mockPVZUseCase
}

// --- Setup ---
func (s *CreatePVZSuite) SetupTest() {
	s.app = fiber.New()
	s.mock = new(mockPVZUseCase)
	handler := post.NewCreatePVZHandler(s.mock)

	s.app.Post("/pvz", func(c *fiber.Ctx) error {
		c.Locals("Role", s.T().Context().Value("role"))
		return handler.Handle(c)
	})
}

// --- Успешный кейс ---
func (s *CreatePVZSuite) TestHandle_Success() {
	// Arrange
	body := `{"city":"Москва"}`
	city := models.PVZCity("Москва")
	expectedPVZ := models.PVZ{
		ID:               "pvz-1",
		City:             city,
		RegistrationDate: time.Now(),
	}

	s.mock.On("CreatePVZ", mock.Anything, city).Return(expectedPVZ, nil)

	req := httptest.NewRequest("POST", "/pvz", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleModerator))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(201, resp.StatusCode)

	var response map[string]any
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&response))
	s.Equal(expectedPVZ.ID, response["id"])
	s.Equal(string(expectedPVZ.City), response["city"])

	s.mock.AssertExpectations(s.T())
}

// --- Ошибка доступа ---
func (s *CreatePVZSuite) TestHandle_Forbidden() {
	// Arrange
	req := httptest.NewRequest("POST", "/pvz", nil)
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleEmployee))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(403, resp.StatusCode)
}

// --- Ошибка парсинга JSON ---
func (s *CreatePVZSuite) TestHandle_BadRequest_InvalidJSON() {
	// Arrange
	req := httptest.NewRequest("POST", "/pvz", bytes.NewBufferString(`invalid-json`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleModerator))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(400, resp.StatusCode)
}

// --- Ошибка валидации (пустой город) ---
func (s *CreatePVZSuite) TestHandle_BadRequest_ValidationError() {
	// Arrange
	body := `{"city":""}`
	req := httptest.NewRequest("POST", "/pvz", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleModerator))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(400, resp.StatusCode)
}

// --- Ошибка валидации (город не из списка) ---
func (s *CreatePVZSuite) TestHandle_BadRequest_InvalidCity() {
	// Arrange
	body := `{"city":"Готэм"}`
	req := httptest.NewRequest("POST", "/pvz", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleModerator))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(400, resp.StatusCode)
}

// --- Ошибка из UseCase ---
func (s *CreatePVZSuite) TestHandle_UseCaseError() {
	// Arrange
	city := models.PVZCity("Москва")
	body := `{"city":"Москва"}`
	s.mock.On("CreatePVZ", mock.Anything, city).
		Return(models.PVZ{}, errors.New("db is down"))

	req := httptest.NewRequest("POST", "/pvz", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleModerator))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(400, resp.StatusCode)
	s.mock.AssertExpectations(s.T())
}

// --- Запуск ---
func TestCreatePVZSuite(t *testing.T) {
	suite.Run(t, new(CreatePVZSuite))
}
