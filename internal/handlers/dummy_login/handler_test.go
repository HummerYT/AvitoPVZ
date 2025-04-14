package dummy_login_test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"AvitoPVZ/internal/handlers/dummy_login"
	"AvitoPVZ/internal/models"
)

type DummyLoginTestSuite struct {
	suite.Suite
	app *fiber.App
}

func (suite *DummyLoginTestSuite) SetupTest() {
	suite.app = fiber.New()
	suite.app.Post("/login", dummy_login.DummyLoginHandler, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"userID": c.Locals("UserID"),
			"role":   c.Locals("Role"),
		})
	})
}

func (suite *DummyLoginTestSuite) TestInvalidJSON() {
	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)
	suite.NoError(err)
	suite.Equal(fiber.StatusBadRequest, resp.StatusCode)

	var errResp models.ErrorResp
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	suite.NoError(err)
	suite.Equal("Invalid request body", errResp.Message)
}

func (suite *DummyLoginTestSuite) TestInvalidRole() {
	body := `{"role": "admin"}`
	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := suite.app.Test(req)
	suite.NoError(err)
	suite.Equal(fiber.StatusBadRequest, resp.StatusCode)
	var errResp models.ErrorResp
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	suite.NoError(err)
	suite.Equal("Role must be 'employee' or 'moderator'", errResp.Message)
}

func (suite *DummyLoginTestSuite) TestValidLogin() {

	body := `{"role": "employee"}`
	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)
	suite.NoError(err)
	suite.Equal(fiber.StatusOK, resp.StatusCode)

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	suite.NoError(err)

	suite.Equal("employee", data["role"])
	userID, ok := data["userID"].(string)
	suite.True(ok, "UserID должен быть строкой")
	_, err = uuid.Parse(userID)
	suite.NoError(err, "UserID должен быть валидным UUID")
}

func TestDummyLoginTestSuite(t *testing.T) {
	suite.Run(t, new(DummyLoginTestSuite))
}
