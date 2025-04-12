package register

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"

	"AvitoPVZ/internal/models"
)

type register interface {
	RegisterUser(ctx context.Context, user models.User) (string, error)
}

type Handler struct {
	register register
}

func NewHandler(register register) *Handler {
	return &Handler{
		register: register,
	}
}

func (h *Handler) Register(ctx *fiber.Ctx) error {
	var req userAuthIn

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(models.ErrorResp{
			Message: "invalid request body",
		})
	}

	user, err := req.validate()
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(models.ErrorResp{
			Message: fmt.Sprintf("invalid request body: %s", err.Error()),
		})
	}

	userID, err := h.register.RegisterUser(ctx.Context(), user)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(models.ErrorResp{
			Message: fmt.Sprintf("register user failed: %s", err.Error()),
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":    userID,
		"email": user.Email,
		"role":  user.Role,
	})
}
