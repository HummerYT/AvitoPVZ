package delete_last_product_test

import (
	"context"
	"errors"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"AvitoPVZ/internal/handlers/delete_last_product"
	"AvitoPVZ/internal/models"
)

// --- Мок UseCase ---
type mockProductUseCase struct {
	mock.Mock
}

func (m *mockProductUseCase) DeleteLastProduct(ctx context.Context, pvzID string) error {
	args := m.Called(ctx, pvzID)
	return args.Error(0)
}

// --- Сьют ---
type DeleteLastProductSuite struct {
	suite.Suite
	app  *fiber.App
	mock *mockProductUseCase
}

// --- Setup ---
func (s *DeleteLastProductSuite) SetupTest() {
	s.app = fiber.New()
	s.mock = new(mockProductUseCase)

	handler := delete_last_product.NewProductHandler(s.mock)

	s.app.Delete("/product/:pvzId/delete", func(c *fiber.Ctx) error {
		c.Locals("Role", s.T().Context().Value("role"))
		return handler.DeleteLastProduct(c)
	})
}

// --- Успешное удаление ---
func (s *DeleteLastProductSuite) TestDeleteLastProduct_Success() {
	// Arrange
	pvzID := uuid.New().String()

	s.mock.On("DeleteLastProduct", mock.Anything, pvzID).
		Return(nil)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/product/%s/delete", pvzID), nil)
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleEmployee))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(fiber.StatusOK, resp.StatusCode)
	s.mock.AssertExpectations(s.T())
}

// --- Доступ запрещён ---
func (s *DeleteLastProductSuite) TestDeleteLastProduct_Forbidden() {
	// Arrange
	pvzID := uuid.New().String()
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/product/%s/delete", pvzID), nil)
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleAdmin)) // не сотрудник

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(fiber.StatusForbidden, resp.StatusCode)
}

// --- Невалидный UUID ---
func (s *DeleteLastProductSuite) TestDeleteLastProduct_InvalidUUID() {
	// Arrange
	invalidPvzID := "not-a-uuid"
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/product/%s/delete", invalidPvzID), nil)
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleEmployee))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(fiber.StatusBadRequest, resp.StatusCode)
}

// --- UseCase вернул ошибку ---
func (s *DeleteLastProductSuite) TestDeleteLastProduct_UseCaseError() {
	// Arrange
	pvzID := uuid.New().String()

	s.mock.On("DeleteLastProduct", mock.Anything, pvzID).
		Return(errors.New("deletion failed"))

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/product/%s/delete", pvzID), nil)
	req = req.WithContext(context.WithValue(req.Context(), "role", models.RoleEmployee))

	// Act
	resp, err := s.app.Test(req)

	// Assert
	s.Require().NoError(err)
	s.Equal(fiber.StatusBadRequest, resp.StatusCode)
	s.mock.AssertExpectations(s.T())
}

// --- Запуск тестов ---
func TestDeleteLastProductSuite(t *testing.T) {
	suite.Run(t, new(DeleteLastProductSuite))
}
