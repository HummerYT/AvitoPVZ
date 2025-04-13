package login_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"AvitoPVZ/internal/handlers/login"
	"AvitoPVZ/internal/models"
)

// --- Мок login интерфейса ---
type loginMock struct {
	mock.Mock
}

func (m *loginMock) LoginUser(ctx context.Context, user models.User) (models.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(models.User), args.Error(1)
}

// --- Тестовая структура ---
type LoginHandlerSuite struct {
	suite.Suite
	app    *fiber.App
	mock   *loginMock
	router *login.Handler
}

// --- Setup ---
func (s *LoginHandlerSuite) SetupTest() {
	s.app = fiber.New()
	s.mock = new(loginMock)
	s.router = login.NewHandler(s.mock)

	s.app.Post("/login", s.router.Register, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"user_id": c.Locals("UserID"),
			"role":    c.Locals("Role"),
		})
	})
}

// --- Тест успешного логина ---
func (s *LoginHandlerSuite) TestRegister_Success() {
	// Arrange
	reqBody := login.UserLoginIn{
		Email:    "test@example.com",
		Password: "StrongPassword123!",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	id := uuid.New()
	expectedUser := models.User{
		ID:       id,
		Email:    reqBody.Email,
		Password: reqBody.Password,
		Role:     "admin",
	}

	s.mock.On("LoginUser", mock.Anything, mock.MatchedBy(func(u models.User) bool {
		return u.Email == reqBody.Email && u.Password == reqBody.Password
	})).Return(expectedUser, nil)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(fiber.StatusOK, resp.StatusCode)
	s.mock.AssertExpectations(s.T())
}

// --- Тест с невалидным JSON ---
func (s *LoginHandlerSuite) TestRegister_InvalidJSON() {
	// Arrange
	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(`invalid-json`))
	req.Header.Set("Content-Type", "application/json")

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(fiber.StatusBadRequest, resp.StatusCode)
}

// --- Тест с ошибкой валидации (простой пароль) ---
func (s *LoginHandlerSuite) TestRegister_ValidationError() {
	// Arrange
	reqBody := login.UserLoginIn{
		Email:    "test@example.com",
		Password: "123", // слабый пароль
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(fiber.StatusBadRequest, resp.StatusCode)
}

// --- Тест когда LoginUser возвращает ошибку ---
func (s *LoginHandlerSuite) TestRegister_LoginError() {
	// Arrange
	reqBody := login.UserLoginIn{
		Email:    "test@example.com",
		Password: "StrongPassword123!",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	s.mock.On("LoginUser", mock.Anything, mock.Anything).Return(models.User{}, errors.New("login failed"))

	req := httptest.NewRequest("POST", "/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(fiber.StatusBadRequest, resp.StatusCode)
	s.mock.AssertExpectations(s.T())
}

// --- Запуск ---
func TestLoginHandlerSuite(t *testing.T) {
	suite.Run(t, new(LoginHandlerSuite))
}
