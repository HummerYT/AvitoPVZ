package receptions_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"AvitoPVZ/internal/handlers/receptions"
	"AvitoPVZ/internal/models"
)

type mockReceptionUseCase struct {
	mock.Mock
}

func (m *mockReceptionUseCase) CreateReception(ctx context.Context, pvzID uuid.UUID) (models.Reception, error) {
	args := m.Called(ctx, pvzID)
	return args.Get(0).(models.Reception), args.Error(1)
}

type ReceptionHandlerSuite struct {
	suite.Suite
	app  *fiber.App
	mock *mockReceptionUseCase
}

func (s *ReceptionHandlerSuite) SetupTest() {
	s.app = fiber.New()
	s.mock = new(mockReceptionUseCase)
	handler := receptions.NewReceptionHandler(s.mock)

	s.app.Use(func(c *fiber.Ctx) error {
		if role := c.Get("X-Role"); role != "" {
			c.Locals("Role", models.UserRole(role))
		}
		return c.Next()
	})

	s.app.Post("/reception", handler.CreateReception)
}

func (s *ReceptionHandlerSuite) Test_CreateReception_Success() {
	pvzID := uuid.New()
	id := uuid.New()
	expected := models.Reception{
		ID:       id,
		PvzID:    pvzID,
		DateTime: time.Now(),
		Status:   models.StatusInProgress,
	}
	s.mock.On("CreateReception", mock.Anything, pvzID).Return(expected, nil)

	body := map[string]string{"pvzId": pvzID.String()}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/reception", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(models.RoleEmployee))

	resp, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(201, resp.StatusCode)

	var result map[string]interface{}
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&result))
	s.Equal(expected.ID.String(), result["id"])
	s.Equal(expected.PvzID.String(), result["pvzId"])
	s.Equal(string(expected.Status), result["status"])

	s.mock.AssertExpectations(s.T())
}

func (s *ReceptionHandlerSuite) Test_CreateReception_Forbidden() {
	req := httptest.NewRequest("POST", "/reception", nil)
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleModerator))

	resp, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(403, resp.StatusCode)
}

func (s *ReceptionHandlerSuite) Test_CreateReception_BadJSON() {
	req := httptest.NewRequest("POST", "/reception", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(models.RoleEmployee))

	resp, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(400, resp.StatusCode)
}

func (s *ReceptionHandlerSuite) Test_CreateReception_EmptyPvzID() {
	body := map[string]string{"pvzId": ""}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/reception", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(models.RoleEmployee))

	resp, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(400, resp.StatusCode)
}

func (s *ReceptionHandlerSuite) Test_CreateReception_InvalidUUID() {
	body := map[string]string{"pvzId": "not-a-uuid"}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/reception", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(models.RoleEmployee))

	resp, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(400, resp.StatusCode)
}

func (s *ReceptionHandlerSuite) Test_CreateReception_UseCaseError() {
	pvzID := uuid.New()
	s.mock.On("CreateReception", mock.Anything, pvzID).Return(models.Reception{}, errors.New("fail"))

	body := map[string]string{"pvzId": pvzID.String()}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/reception", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(models.RoleEmployee))

	resp, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(400, resp.StatusCode)
	s.mock.AssertExpectations(s.T())
}

func TestReceptionHandlerSuite(t *testing.T) {
	suite.Run(t, new(ReceptionHandlerSuite))
}
