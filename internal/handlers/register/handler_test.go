package register_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"AvitoPVZ/internal/handlers/register"
	"AvitoPVZ/internal/models"
)

type mockRegister struct {
	mock.Mock
}

func (m *mockRegister) RegisterUser(ctx context.Context, user models.User) (string, error) {
	args := m.Called(ctx, user)
	return args.String(0), args.Error(1)
}

type RegisterHandlerSuite struct {
	suite.Suite
	app     *fiber.App
	mockReg *mockRegister
}

func (s *RegisterHandlerSuite) SetupTest() {
	s.app = fiber.New()
	s.mockReg = new(mockRegister)
	handler := register.NewHandler(s.mockReg)

	s.app.Post("/register", handler.Register)
}

func (s *RegisterHandlerSuite) Test_Register_Success() {
	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "StrongPassword123!",
		"role":     string(models.RoleEmployee),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	expectedUser := models.User{
		Email:    reqBody["email"],
		Password: reqBody["password"],
		Role:     models.UserRole(reqBody["role"]),
	}
	expectedID := uuid.New().String()

	s.mockReg.On("RegisterUser", mock.Anything, mock.MatchedBy(func(u models.User) bool {
		return u.Email == expectedUser.Email && u.Password == expectedUser.Password && u.Role == expectedUser.Role
	})).Return(expectedID, nil)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(fiber.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&result))
	s.Equal(expectedID, result["id"])
	s.Equal(reqBody["email"], result["email"])
	s.Equal(reqBody["role"], result["role"])

	s.mockReg.AssertExpectations(s.T())
}

func (s *RegisterHandlerSuite) Test_Register_InvalidJSON() {
	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(fiber.StatusBadRequest, resp.StatusCode)
}

func (s *RegisterHandlerSuite) Test_Register_WeakPassword() {
	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "123", // слишком слабый
		"role":     string(models.RoleEmployee),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(fiber.StatusBadRequest, resp.StatusCode)
}

func (s *RegisterHandlerSuite) Test_Register_InvalidRole() {
	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "StrongPassword123!",
		"role":     "unknown",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(fiber.StatusBadRequest, resp.StatusCode)
}

func (s *RegisterHandlerSuite) Test_Register_UseCaseError() {
	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "StrongPassword123!",
		"role":     string(models.RoleModerator),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	expectedUser := models.User{
		Email:    reqBody["email"],
		Password: reqBody["password"],
		Role:     models.UserRole(reqBody["role"]),
	}

	s.mockReg.On("RegisterUser", mock.Anything, mock.MatchedBy(func(u models.User) bool {
		return u.Email == expectedUser.Email
	})).Return("", errors.New("registration failed"))

	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(fiber.StatusBadRequest, resp.StatusCode)
	s.mockReg.AssertExpectations(s.T())
}

func TestRegisterHandlerSuite(t *testing.T) {
	suite.Run(t, new(RegisterHandlerSuite))
}
