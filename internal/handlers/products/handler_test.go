package products_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"

	"AvitoPVZ/internal/handlers/products"
	"AvitoPVZ/internal/models"
)

type mockProductUseCase struct {
	product models.Product
	err     error
}

func (m *mockProductUseCase) CreateProduct(_ context.Context, _ uuid.UUID, _ models.TypeProduct) (models.Product, error) {
	return m.product, m.err
}

type ProductHandlerTestSuite struct {
	suite.Suite
	app     *fiber.App
	useCase *mockProductUseCase
	handler *products.ProductHandler
}

func (suite *ProductHandlerTestSuite) SetupTest() {
	suite.app = fiber.New()

	suite.app.Use(func(c *fiber.Ctx) error {
		if role := c.Get("X-Role"); role != "" {
			c.Locals("Role", models.UserRole(role))
		}
		return c.Next()
	})

	suite.useCase = &mockProductUseCase{}
	suite.handler = products.NewProductHandler(suite.useCase)

	suite.app.Post("/products", suite.handler.CreateProduct)
}

func (suite *ProductHandlerTestSuite) TestAccessDenied() {
	reqBody := `{"pvzId": "123e4567-e89b-12d3-a456-426614174000", "type": "одежда"}`
	req := httptest.NewRequest("POST", "/products", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	suite.Equal(http.StatusForbidden, resp.StatusCode)
	var body models.ErrorResp
	err = json.NewDecoder(resp.Body).Decode(&body)
	suite.Require().NoError(err)
	suite.Equal("Access denied (only PVZ employee)", body.Message)
}

func (suite *ProductHandlerTestSuite) TestBadBody() {
	reqBody := `invalid-json`
	req := httptest.NewRequest("POST", "/products", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(models.RoleEmployee))

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
	var body models.ErrorResp
	err = json.NewDecoder(resp.Body).Decode(&body)
	suite.Require().NoError(err)
	suite.Equal("Bad Request", body.Message)
}

func (suite *ProductHandlerTestSuite) TestValidationError() {
	reqBody := `{"pvzId": "invalid-uuid", "type": "some_type"}`
	req := httptest.NewRequest("POST", "/products", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(models.RoleEmployee))

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
	var body models.ErrorResp
	err = json.NewDecoder(resp.Body).Decode(&body)
	suite.Require().NoError(err)
	suite.NotEmpty(body.Message)
}

func (suite *ProductHandlerTestSuite) TestUseCaseError() {
	validUUID := "123e4567-e89b-12d3-a456-426614174000"
	reqBody := `{"pvzId": "` + validUUID + `", "type": "одежда"}`
	req := httptest.NewRequest("POST", "/products", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(models.RoleEmployee))

	suite.useCase.err = errors.New("use case error")

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	suite.Equal(http.StatusBadRequest, resp.StatusCode)
	var body models.ErrorResp
	err = json.NewDecoder(resp.Body).Decode(&body)
	suite.Require().NoError(err)
	suite.Equal("use case error", body.Message)
}

func (suite *ProductHandlerTestSuite) TestSuccess() {
	validUUID := "123e4567-e89b-12d3-a456-426614174000"
	pvzID, err := uuid.Parse(validUUID)
	suite.Require().NoError(err)

	product := models.Product{
		ID:          uuid.New(),
		DateTime:    time.Now(),
		Type:        models.TypeClothes,
		ReceptionID: pvzID,
	}
	suite.useCase.err = nil
	suite.useCase.product = product

	reqBody := `{"pvzId": "` + validUUID + `", "type": "одежда"}`
	req := httptest.NewRequest("POST", "/products", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Role", string(models.RoleEmployee))

	resp, err := suite.app.Test(req)
	suite.Require().NoError(err)

	suite.Equal(http.StatusCreated, resp.StatusCode)
	var payload map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&payload)
	suite.Require().NoError(err)

	suite.Equal(product.ID.String(), payload["id"])
	suite.Equal(string(product.Type), payload["type"])
	suite.Equal(product.ReceptionID.String(), payload["receptionId"])
	expectedDateTime := product.DateTime.Format("2006-01-02 15:04:05")
	suite.Equal(expectedDateTime, payload["dateTime"])
}

func TestProductHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ProductHandlerTestSuite))
}
