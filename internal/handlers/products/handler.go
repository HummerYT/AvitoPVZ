package products

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"AvitoPVZ/internal/models"
)

type ProductUseCase interface {
	CreateProduct(ctx context.Context, pvzID uuid.UUID, productType models.TypeProduct) (models.Product, error)
}

type ProductHandler struct {
	UC ProductUseCase
}

func NewProductHandler(uc ProductUseCase) *ProductHandler {
	return &ProductHandler{UC: uc}
}

func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	userRole, ok := c.Locals("Role").(models.UserRole)
	if !ok || userRole != models.RoleEmployee {
		return c.Status(http.StatusForbidden).JSON(models.ErrorResp{
			Message: "Access denied (only PVZ employee)",
		})
	}

	var req ReqProducts
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: "Bad Request",
		})
	}

	typeProduct, pvzID, err := req.validate()
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: err.Error(),
		})
	}

	product, err := h.UC.CreateProduct(c.Context(), pvzID, typeProduct)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"id":          product.ID.String(),
		"dateTime":    product.DateTime.Format("2006-01-02 15:04:05"),
		"type":        product.Type,
		"receptionId": product.ReceptionID,
	})
}
