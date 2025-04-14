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

	"AvitoPVZ/internal/handlers/pvz/delete_last_product"
	"AvitoPVZ/internal/models"
)

type mockProductUseCase struct {
	mock.Mock
}

func (m *mockProductUseCase) DeleteLastProduct(ctx context.Context, pvzID string) error {
	args := m.Called(ctx, pvzID)
	return args.Error(0)
}

type DeleteLastProductSuite struct {
	suite.Suite
	app  *fiber.App
	mock *mockProductUseCase
}

func (s *DeleteLastProductSuite) SetupTest() {
	s.app = fiber.New()
	s.mock = new(mockProductUseCase)

	handler := delete_last_product.NewProductHandler(s.mock)

	s.app.Use(func(c *fiber.Ctx) error {
		if role := c.Get("X-Role"); role != "" {
			c.Locals("Role", models.UserRole(role))
		}
		return c.Next()
	})

	s.app.Delete("/product/:pvzId/delete", handler.DeleteLastProduct)
}

func (s *DeleteLastProductSuite) TestDeleteLastProduct_Success() {
	pvzID := uuid.New().String()

	s.mock.On("DeleteLastProduct", mock.Anything, pvzID).
		Return(nil)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/product/%s/delete", pvzID), nil)
	req.Header.Set("X-Role", string(models.RoleEmployee))

	resp, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(fiber.StatusOK, resp.StatusCode)
	s.mock.AssertExpectations(s.T())
}

func (s *DeleteLastProductSuite) TestDeleteLastProduct_Forbidden() {
	pvzID := uuid.New().String()
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/product/%s/delete", pvzID), nil)
	req.Header.Set("X-Role", string(models.RoleModerator))

	resp, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(fiber.StatusForbidden, resp.StatusCode)
}

func (s *DeleteLastProductSuite) TestDeleteLastProduct_InvalidUUID() {
	invalidPvzID := "not-a-uuid"
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/product/%s/delete", invalidPvzID), nil)
	req.Header.Set("X-Role", string(models.RoleEmployee))

	resp, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(fiber.StatusBadRequest, resp.StatusCode)
}

func (s *DeleteLastProductSuite) TestDeleteLastProduct_UseCaseError() {
	pvzID := uuid.New().String()

	s.mock.On("DeleteLastProduct", mock.Anything, pvzID).
		Return(errors.New("deletion failed"))

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/product/%s/delete", pvzID), nil)
	req.Header.Set("X-Role", string(models.RoleEmployee))

	resp, err := s.app.Test(req)

	s.Require().NoError(err)
	s.Equal(fiber.StatusBadRequest, resp.StatusCode)
	s.mock.AssertExpectations(s.T())
}

func TestDeleteLastProductSuite(t *testing.T) {
	suite.Run(t, new(DeleteLastProductSuite))
}
