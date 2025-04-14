package get_test

import (
	"AvitoPVZ/internal/handlers/pvz/get"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"AvitoPVZ/internal/models"
)

type mockPVZDataUseCase struct {
	data []models.PVZData
	err  error
}

func (m *mockPVZDataUseCase) GetPVZData(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]models.PVZData, error) {
	return m.data, m.err
}

type PVZDataHandlerTestSuite struct {
	suite.Suite
	app     *fiber.App
	uc      *mockPVZDataUseCase
	handler *get.PVZDataHandler
}

func (suite *PVZDataHandlerTestSuite) SetupTest() {
	suite.app = fiber.New()

	suite.app.Use(func(c *fiber.Ctx) error {
		if role := c.Get("X-Role"); role != "" {
			c.Locals("Role", models.UserRole(role))
		}
		return c.Next()
	})

	suite.uc = &mockPVZDataUseCase{}
	suite.handler = get.NewPVZDataHandler(suite.uc)

	suite.app.Get("/pvzdata", suite.handler.GetPVZData)
}

func (suite *PVZDataHandlerTestSuite) TestAccessDenied() {
	req := httptest.NewRequest("GET", "/pvzdata", nil)

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	suite.Equal(http.StatusForbidden, resp.StatusCode)
	var body models.ErrorResp
	err = json.NewDecoder(resp.Body).Decode(&body)
	suite.Require().NoError(err)
	suite.Equal("access denied", body.Message)
}

func (suite *PVZDataHandlerTestSuite) TestValidationError() {
	req := httptest.NewRequest("GET", "/pvzdata?page=-1&limit=10", nil)
	req.Header.Set("X-Role", string(models.RoleEmployee)) // предполагаем, что "employee" проходит проверку models.IsUserRole

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
	var body models.ErrorResp
	err = json.NewDecoder(resp.Body).Decode(&body)
	suite.Require().NoError(err)
	suite.Contains(body.Message, "Неверные параметры запроса:")
}

func (suite *PVZDataHandlerTestSuite) TestInvalidStartDate() {
	req := httptest.NewRequest("GET", "/pvzdata?startDate=invalid-date", nil)
	req.Header.Set("X-Role", "employee")

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
	var body models.ErrorResp
	err = json.NewDecoder(resp.Body).Decode(&body)
	suite.Require().NoError(err)
	suite.Equal(
		"Неверные параметры запроса: Key: 'PVZDataRequest.StartDate' Error:Field validation for 'StartDate' failed on the 'datetime' tag",
		body.Message,
	)
}

func (suite *PVZDataHandlerTestSuite) TestInvalidEndDate() {
	fixedTime := time.Date(2025, 4, 13, 10, 30, 0, 0, time.UTC)
	validDate := fixedTime.Format(time.RFC3339)
	req := httptest.NewRequest("GET", "/pvzdata?startDate="+validDate+"&endDate=bad-date", nil)
	req.Header.Set("X-Role", "employee")

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
	var body models.ErrorResp
	err = json.NewDecoder(resp.Body).Decode(&body)
	suite.Require().NoError(err)
	suite.True(strings.Contains(body.Message, "Неверные параметры запроса"))
}

func (suite *PVZDataHandlerTestSuite) TestUseCaseError() {
	fixedTime := time.Date(2025, 4, 13, 10, 30, 0, 0, time.UTC)
	validDate := fixedTime.Format(time.RFC3339)
	req := httptest.NewRequest("GET", "/pvzdata?startDate="+validDate+"&page="+strconv.Itoa(2)+"&limit="+strconv.Itoa(5), nil)
	req.Header.Set("X-Role", "employee")

	suite.uc.err = errors.New("use case error")

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
	var body models.ErrorResp
	err = json.NewDecoder(resp.Body).Decode(&body)
	suite.Require().NoError(err)
	suite.Equal("use case error", body.Message)
}

func (suite *PVZDataHandlerTestSuite) TestSuccess() {
	fixedTime := time.Date(2025, 4, 13, 10, 30, 0, 0, time.UTC)
	validDate := fixedTime.Format(time.RFC3339)
	req := httptest.NewRequest("GET", "/pvzdata?startDate="+validDate+"&endDate="+validDate+"&page=3&limit=7", nil)
	req.Header.Set("X-Role", "employee")

	pvzData := models.PVZData{
		PVZ: models.PVZ{
			ID:               "pvz-1",
			RegistrationDate: fixedTime,
			City:             "Moscow",
		},
		Receptions: []models.ReceptionData{
			{
				Reception: models.Reception{
					ID:       uuid.New(),
					DateTime: fixedTime,
					PvzID:    uuid.New(),
					Status:   "closed",
				},
				Products: []models.Product{
					{
						ID:          uuid.New(),
						DateTime:    fixedTime,
						Type:        "electronic",
						ReceptionID: uuid.New(),
					},
					{
						ID:          uuid.New(),
						DateTime:    fixedTime,
						Type:        "clothes",
						ReceptionID: uuid.New(),
					},
				},
			},
		},
	}
	suite.uc.err = nil
	suite.uc.data = []models.PVZData{pvzData}

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, resp.StatusCode)

	var result []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	suite.Require().NoError(err)
	suite.Len(result, 1)

	pvzMap, ok := result[0]["pvz"].(map[string]interface{})
	suite.True(ok)
	suite.Equal("pvz-1", pvzMap["id"])
	suite.Equal("Moscow", pvzMap["city"])
	expectedRegDate := fixedTime.Format(time.RFC3339)
	suite.Equal(expectedRegDate, pvzMap["registrationDate"])

	receptions, ok := result[0]["receptions"].([]interface{})
	suite.True(ok)
	suite.Len(receptions, 1)

	recMap, ok := receptions[0].(map[string]interface{})
	suite.True(ok)
	receptionData, ok := recMap["reception"].(map[string]interface{})
	suite.True(ok)
	expectedRecDate := fixedTime.Format(time.RFC3339)
	suite.Equal(expectedRecDate, receptionData["dateTime"])

	products, ok := recMap["products"].([]interface{})
	suite.True(ok)
	suite.Len(products, 2)
}

func TestPVZDataHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(PVZDataHandlerTestSuite))
}
