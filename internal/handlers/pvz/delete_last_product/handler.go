package delete_last_product

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"AvitoPVZ/internal/models"
)

type ProductUseCase interface {
	DeleteLastProduct(ctx context.Context, pvzID string) error
}

type ProductHandler struct {
	UC ProductUseCase
}

func NewProductHandler(uc ProductUseCase) *ProductHandler {
	return &ProductHandler{UC: uc}
}

func (h *ProductHandler) DeleteLastProduct(c *fiber.Ctx) error {
	userRole, ok := c.Locals("Role").(models.UserRole)
	if !ok || userRole != models.RoleEmployee {
		return c.Status(http.StatusForbidden).JSON(models.ErrorResp{
			Message: "Access denied (only PVZ employee)",
		})
	}

	req := DeleteProductRequest{
		PvzID: c.Params("pvzId"),
	}

	if err := validateDeleteProductRequest(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: "PvzID is invalid",
		})
	}

	if err := h.UC.DeleteLastProduct(c.Context(), req.PvzID); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"description": "Товар удален",
	})
}
