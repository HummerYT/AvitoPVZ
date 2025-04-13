package dummy_login_test

import (
	"AvitoPVZ/internal/handlers/dummy_login"
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"AvitoPVZ/internal/models"
)

type DummyLoginTestSuite struct {
	suite.Suite
	app *fiber.App
}

func (suite *DummyLoginTestSuite) SetupTest() {
	// Инициализируем Fiber-приложение и регистрируем маршрут.
	// В цепочке вызовов первый middleware — DummyLoginHandler,
	// а второй (финальный) возвращает установленные значения в JSON-формате.
	suite.app = fiber.New()
	suite.app.Post("/login", dummy_login.DummyLoginHandler, func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"userID": c.Locals("UserID"),
			"role":   c.Locals("Role"),
		})
	})
}

func (suite *DummyLoginTestSuite) TestInvalidJSON() {
	// Arrange: создаём запрос с некорректным JSON
	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	// Act: отправляем запрос через тестовое приложение
	resp, err := suite.app.Test(req)
	suite.NoError(err)

	// Assert: ожидаем статус 400 и сообщение об ошибке в теле ответа
	suite.Equal(fiber.StatusBadRequest, resp.StatusCode)

	var errResp models.ErrorResp
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	suite.NoError(err)
	suite.Equal("Invalid request body", errResp.Message)
}

func (suite *DummyLoginTestSuite) TestInvalidRole() {
	// Arrange: создаём запрос с корректным JSON, но неверной ролью "admin"
	body := `{"role": "admin"}`
	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	// Act
	resp, err := suite.app.Test(req)
	suite.NoError(err)

	// Assert
	suite.Equal(fiber.StatusBadRequest, resp.StatusCode)
	var errResp models.ErrorResp
	err = json.NewDecoder(resp.Body).Decode(&errResp)
	suite.NoError(err)
	suite.Equal("Role must be 'employee' or 'moderator'", errResp.Message)
}

func (suite *DummyLoginTestSuite) TestValidLogin() {
	// Arrange: создаём запрос с корректным JSON, роль "employee"
	body := `{"role": "employee"}`
	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	// Act
	resp, err := suite.app.Test(req)
	suite.NoError(err)
	suite.Equal(fiber.StatusOK, resp.StatusCode)

	// Assert: проверяем, что ответ содержит корректные значения
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
