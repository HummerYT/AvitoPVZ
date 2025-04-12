package receptions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"AvitoPVZ/internal/models"
)

type ReceptionUseCase interface {
	CreateReception(ctx context.Context, pvzID uuid.UUID) (models.Reception, error)
}

type ReceptionHandler struct {
	UC ReceptionUseCase
}

func NewReceptionHandler(uc ReceptionUseCase) *ReceptionHandler {
	return &ReceptionHandler{UC: uc}
}

func (h *ReceptionHandler) CreateReception(c *fiber.Ctx) error {
	userRole, ok := c.Locals("Role").(models.UserRole)
	if !ok || !models.IsUserRole(userRole) || userRole != models.RoleEmployee {
		return c.Status(http.StatusForbidden).JSON(models.ErrorResp{
			Message: "Access denied (only for PVZ employee)",
		})
	}

	var req struct {
		PvzID string `json:"pvzId"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: "Bad request",
		})
	}
	if req.PvzID == "" {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: "pvzId is required",
		})
	}

	PvzUUID, err := uuid.Parse(req.PvzID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: fmt.Sprintf("Invalid Pvz ID: %s", req.PvzID),
		})
	}

	reception, err := h.UC.CreateReception(c.Context(), PvzUUID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"id":       reception.ID,
		"dateTime": reception.DateTime,
		"pvzId":    reception.PvzID,
		"status":   reception.Status,
	})
}
