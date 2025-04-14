package post

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/suite"

	"AvitoPVZ/internal/models"
)

type mockPVZUseCase struct {
	pvz models.PVZ
	err error
}

func (m *mockPVZUseCase) CreatePVZ(ctx context.Context, city models.PVZCity) (models.PVZ, error) {
	return m.pvz, m.err
}

type CreatePVZHandlerTestSuite struct {
	suite.Suite
	app     *fiber.App
	uc      *mockPVZUseCase
	handler *CreatePVZHandler
}

func (suite *CreatePVZHandlerTestSuite) SetupTest() {
	suite.app = fiber.New()

	suite.app.Use(func(c *fiber.Ctx) error {
		if role := c.Get("X-Role"); role != "" {
			c.Locals("Role", models.UserRole(role))
		}
		return c.Next()
	})

	suite.uc = &mockPVZUseCase{}
	suite.handler = NewCreatePVZHandler(suite.uc)

	suite.app.Post("/pvz", suite.handler.Handle)
}

func (suite *CreatePVZHandlerTestSuite) TestAccessDenied() {
	reqBody := `{"city": "Москва"}`
	req := httptest.NewRequest("POST", "/pvz", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", "employee") // неверная роль

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	suite.Equal(http.StatusForbidden, resp.StatusCode)
	var body models.ErrorResp
	err = json.NewDecoder(resp.Body).Decode(&body)
	suite.Require().NoError(err)
	suite.Equal("access denied", body.Message)
}

func (suite *CreatePVZHandlerTestSuite) TestBadBody() {
	req := httptest.NewRequest("POST", "/pvz", strings.NewReader("invalid-json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(models.RoleModerator))

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
	var body models.ErrorResp
	err = json.NewDecoder(resp.Body).Decode(&body)
	suite.Require().NoError(err)
	suite.Equal("bad request", body.Message)
}

func (suite *CreatePVZHandlerTestSuite) TestValidationErrorMissingCity() {
	req := httptest.NewRequest("POST", "/pvz", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(models.RoleModerator))

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
	var body models.ErrorResp
	err = json.NewDecoder(resp.Body).Decode(&body)
	suite.Require().NoError(err)
	suite.Contains(body.Message, "bad request:")
}

func (suite *CreatePVZHandlerTestSuite) TestValidationErrorInvalidCity() {
	req := httptest.NewRequest("POST", "/pvz", strings.NewReader(`{"city": "NewYork"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(models.RoleModerator))

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
	var body models.ErrorResp
	err = json.NewDecoder(resp.Body).Decode(&body)
	suite.Require().NoError(err)
	suite.Contains(body.Message, "bad request:")
	suite.Contains(body.Message, "city is not allowed")
}

func (suite *CreatePVZHandlerTestSuite) TestUseCaseError() {
	reqBody := `{"city": "Москва"}`
	req := httptest.NewRequest("POST", "/pvz", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(models.RoleModerator))

	suite.uc.err = errors.New("use case error")

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
	var body models.ErrorResp
	err = json.NewDecoder(resp.Body).Decode(&body)
	suite.Require().NoError(err)
	suite.Contains(body.Message, "create pvz failed")
}

func (suite *CreatePVZHandlerTestSuite) TestSuccess() {
	reqBody := `{"city": "Москва"}`
	req := httptest.NewRequest("POST", "/pvz", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(models.RoleModerator))

	fixedTime := time.Date(2025, 4, 13, 12, 0, 0, 0, time.UTC)
	pvz := models.PVZ{
		ID:               uuid.New().String(),
		RegistrationDate: fixedTime,
		City:             "Moscow",
	}
	suite.uc.err = nil
	suite.uc.pvz = pvz

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)
	suite.Equal(http.StatusCreated, resp.StatusCode)

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	suite.Require().NoError(err)
	suite.Equal(pvz.ID, body["id"])
	suite.Equal(pvz.City, body["city"])

	expectedRegDate := fixedTime.Format(time.RFC3339Nano)
	suite.Equal(expectedRegDate, body["registrationDate"])
}

func TestCreatePVZHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(CreatePVZHandlerTestSuite))
}
