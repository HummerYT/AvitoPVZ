package close_last_reception

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/suite"

	"AvitoPVZ/internal/models"
)

type mockReceptionUseCase struct {
	response models.Reception
	err      error
}

func (m *mockReceptionUseCase) CloseLastReception(_ context.Context, _ string) (models.Reception, error) {
	return m.response, m.err
}

type ReceptionHandlerTestSuite struct {
	suite.Suite
	app     *fiber.App
	useCase *mockReceptionUseCase
	handler *ReceptionHandler
}

func (suite *ReceptionHandlerTestSuite) SetupTest() {
	suite.app = fiber.New()

	suite.app.Use(func(c *fiber.Ctx) error {
		if role := c.Get("X-Role"); role != "" {
			c.Locals("Role", models.UserRole(role))
		}
		return c.Next()
	})

	suite.useCase = &mockReceptionUseCase{}

	suite.handler = NewReceptionHandler(suite.useCase)

	suite.app.Post("/close/:pvzId", suite.handler.CloseLastReception)
}

func (suite *ReceptionHandlerTestSuite) TestAccessDenied() {
	validUUID := "123e4567-e89b-12d3-a456-426614174000"
	req := httptest.NewRequest("POST", "/close/"+validUUID, nil)

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	suite.Equal(http.StatusForbidden, resp.StatusCode)

	var body models.ErrorResp
	err = json.NewDecoder(resp.Body).Decode(&body)
	suite.Require().NoError(err)
	suite.Equal("Access denied (only PVZ employee)", body.Message)
}

func (suite *ReceptionHandlerTestSuite) TestInvalidPvzID() {
	invalidID := "not-a-valid-uuid"
	req := httptest.NewRequest("POST", "/close/"+invalidID, nil)
	req.Header.Set("X-Role", string(models.RoleEmployee))

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
	var body models.ErrorResp
	err = json.NewDecoder(resp.Body).Decode(&body)
	suite.Require().NoError(err)
	suite.Equal("PvzID is invalid", body.Message)
}

func (suite *ReceptionHandlerTestSuite) TestUseCaseError() {
	validUUID := "123e4567-e89b-12d3-a456-426614174000"
	suite.useCase.err = errors.New("use case error")
	req := httptest.NewRequest("POST", "/close/"+validUUID, nil)
	req.Header.Set("X-Role", string(models.RoleEmployee))
	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
	var body models.ErrorResp
	err = json.NewDecoder(resp.Body).Decode(&body)
	suite.Require().NoError(err)
	suite.Equal("use case error", body.Message)
}

func (suite *ReceptionHandlerTestSuite) TestSuccess() {
	validUUID := uuid.New()

	id := uuid.New()
	expectedReception := models.Reception{
		ID:       id,
		DateTime: time.Now(),
		PvzID:    validUUID,
		Status:   models.StatusClose,
	}
	suite.useCase.err = nil
	suite.useCase.response = expectedReception

	req := httptest.NewRequest("POST", "/close/"+validUUID.String(), nil)
	req.Header.Set("X-Role", string(models.RoleEmployee))

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	suite.Equal(http.StatusOK, resp.StatusCode)
	var payload map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&payload)
	suite.Require().NoError(err)

	suite.Equal(expectedReception.ID.String(), payload["id"])
	suite.Equal(expectedReception.DateTime.Format(time.RFC3339Nano), payload["dateTime"])
	suite.Equal(expectedReception.PvzID.String(), payload["pvzId"])
	suite.Equal(string(expectedReception.Status), payload["status"])
}

func TestReceptionHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ReceptionHandlerTestSuite))
}
