package products_test

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

	"AvitoPVZ/internal/handlers/products"
	"AvitoPVZ/internal/models"
)

// --- Мок ProductUseCase ---
type mockProductUseCase struct {
	mock.Mock
}

func (m *mockProductUseCase) CreateProduct(ctx context.Context, pvzID uuid.UUID, productType models.TypeProduct) (models.Product, error) {
	args := m.Called(ctx, pvzID, productType)
	return args.Get(0).(models.Product), args.Error(1)
}

// --- Сьют ---
type ProductHandlerSuite struct {
	suite.Suite
	app    *fiber.App
	mockUC *mockProductUseCase
}

// --- Setup ---
func (s *ProductHandlerSuite) SetupTest() {
	s.app = fiber.New()
	s.mockUC = new(mockProductUseCase)

	handler := products.NewProductHandler(s.mockUC)

	s.app.Post("/products", func(c *fiber.Ctx) error {
		c.Locals("Role", s.T().Context().Value("role"))
		return handler.CreateProduct(c)
	})
}

// --- Успешный кейс ---
func (s *ProductHandlerSuite) TestCreateProduct_Success() {
	// Arrange
	pvzID := uuid.New()
	req := products.ReqProducts{
		Type:  string(models.TypeElectronics),
		PvzID: pvzID.String(),
	}
	body, _ := json.Marshal(req)

	expectedProduct := models.Product{
		ID:          uuid.New(),
		DateTime:    time.Now(),
		Type:        models.TypeElectronics,
		ReceptionID: 101,
	}

	s.mockUC.On("CreateProduct", mock.Anything, pvzID, models.TypeElectronics).
		Return(expectedProduct, nil)

	reqHttp := httptest.NewRequest("POST", "/products", bytes.NewReader(body))
	reqHttp.Header.Set("Content-Type", "application/json")
	reqHttp = reqHttp.WithContext(context.WithValue(reqHttp.Context(), "role", models.RoleEmployee))

	// Act
	resp, err := s.app.Test(reqHttp)

	// Assert
	s.Require().NoError(err)
	s.Equal(fiber.StatusCreated, resp.StatusCode)
	s.mockUC.AssertExpectations(s.T())
}

// --- Ошибка: не сотрудник PVZ ---
func (s *ProductHandlerSuite) TestCreateProduct_Forbidden() {
	// Arrange
	req := products.ReqProducts{
		Type:  string(models.TypeElectronics),
		PvzID: uuid.New().String(),
	}
	body, _ := json.Marshal(req)

	reqHttp := httptest.NewRequest("POST", "/products", bytes.NewReader(body))
	reqHttp.Header.Set("Content-Type", "application/json")
	reqHttp = reqHttp.WithContext(context.WithValue(reqHttp.Context(), "role", models.RoleAdmin)) // неправильная роль

	// Act
	resp, err := s.app.Test(reqHttp)

	// Assert
	s.Require().NoError(err)
	s.Equal(fiber.StatusForbidden, resp.StatusCode)
}

// --- Ошибка: невалидный body ---
func (s *ProductHandlerSuite) TestCreateProduct_InvalidBody() {
	// Arrange
	reqHttp := httptest.NewRequest("POST", "/products", bytes.NewBufferString(`invalid-json`))
	reqHttp.Header.Set("Content-Type", "application/json")
	reqHttp = reqHttp.WithContext(context.WithValue(reqHttp.Context(), "role", models.RoleEmployee))

	// Act
	resp, err := s.app.Test(reqHttp)

	// Assert
	s.Require().NoError(err)
	s.Equal(fiber.StatusBadRequest, resp.StatusCode)
}

// --- Ошибка: невалидный тип продукта ---
func (s *ProductHandlerSuite) TestCreateProduct_InvalidType() {
	// Arrange
	req := products.ReqProducts{
		Type:  "nonexistent_type",
		PvzID: uuid.New().String(),
	}
	body, _ := json.Marshal(req)

	reqHttp := httptest.NewRequest("POST", "/products", bytes.NewReader(body))
	reqHttp.Header.Set("Content-Type", "application/json")
	reqHttp = reqHttp.WithContext(context.WithValue(reqHttp.Context(), "role", models.RoleEmployee))

	// Act
	resp, err := s.app.Test(reqHttp)

	// Assert
	s.Require().NoError(err)
	s.Equal(fiber.StatusBadRequest, resp.StatusCode)
}

// --- Ошибка: use case вернул ошибку ---
func (s *ProductHandlerSuite) TestCreateProduct_UseCaseError() {
	// Arrange
	pvzID := uuid.New()
	req := products.ReqProducts{
		Type:  string(models.TypeElectronics),
		PvzID: pvzID.String(),
	}
	body, _ := json.Marshal(req)

	s.mockUC.On("CreateProduct", mock.Anything, pvzID, models.TypeElectronics).
		Return(models.Product{}, errors.New("some creation error"))

	reqHttp := httptest.NewRequest("POST", "/products", bytes.NewReader(body))
	reqHttp.Header.Set("Content-Type", "application/json")
	reqHttp = reqHttp.WithContext(context.WithValue(reqHttp.Context(), "role", models.RoleEmployee))

	// Act
	resp, err := s.app.Test(reqHttp)

	// Assert
	s.Require().NoError(err)
	s.Equal(fiber.StatusBadRequest, resp.StatusCode)
	s.mockUC.AssertExpectations(s.T())
}

// --- Запуск сьюта ---
func TestProductHandlerSuite(t *testing.T) {
	suite.Run(t, new(ProductHandlerSuite))
}
