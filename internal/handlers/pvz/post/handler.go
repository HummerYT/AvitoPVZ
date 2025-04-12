package post

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"AvitoPVZ/internal/models"
)

type PVZUseCase interface {
	CreatePVZ(ctx context.Context, city models.PVZCity) (models.PVZ, error)
}

type CreatePVZHandler struct {
	UC PVZUseCase
}

func NewCreatePVZHandler(uc PVZUseCase) *CreatePVZHandler {
	return &CreatePVZHandler{UC: uc}
}

func (h *CreatePVZHandler) Handle(ctx *fiber.Ctx) error {
	userRole, ok := ctx.Locals("Role").(models.UserRole)
	if !ok || !models.IsUserRole(userRole) || userRole != models.RoleModerator {
		return ctx.Status(http.StatusForbidden).JSON(models.ErrorResp{
			Message: "access denied",
		})
	}

	var req pvzPostReq
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: "bad request",
		})
	}

	city, err := req.validate()
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: fmt.Sprintf("bad request: %s", err.Error()),
		})
	}

	newPVZ, err := h.UC.CreatePVZ(ctx.Context(), city)
	if err != nil {
		return ctx.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: fmt.Sprintf("create pvz failed: %s", err.Error()),
		})
	}

	return ctx.Status(http.StatusCreated).JSON(fiber.Map{
		"id":               newPVZ.ID,
		"registrationDate": newPVZ.RegistrationDate,
		"city":             newPVZ.City,
	})
}
