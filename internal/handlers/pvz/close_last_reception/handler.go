package close_last_reception

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"AvitoPVZ/internal/models"
)

type ReceptionUseCase interface {
	CloseLastReception(ctx context.Context, pvzID string) (models.Reception, error)
}

type ReceptionHandler struct {
	UC ReceptionUseCase
}

func NewReceptionHandler(uc ReceptionUseCase) *ReceptionHandler {
	return &ReceptionHandler{UC: uc}
}

func (h *ReceptionHandler) CloseLastReception(c *fiber.Ctx) error {
	userRole, ok := c.Locals("Role").(models.UserRole)
	if !ok || userRole != models.RoleEmployee {
		return c.Status(http.StatusForbidden).JSON(models.ErrorResp{
			Message: "Access denied (only PVZ employee)",
		})
	}

	req := CloseReceptionRequest{
		PvzID: c.Params("pvzId"),
	}
	if err := validateCloseReceptionRequest(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: "PvzID is invalid",
		})
	}

	closedRec, err := h.UC.CloseLastReception(c.Context(), req.PvzID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"id":       closedRec.ID,
		"dateTime": closedRec.DateTime,
		"pvzId":    closedRec.PvzID,
		"status":   closedRec.Status,
	})
}
