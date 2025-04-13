package login

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"

	"AvitoPVZ/internal/models"
)

type login interface {
	LoginUser(ctx context.Context, user models.User) (models.User, error)
}

type Handler struct {
	login login
}

func NewHandler(loginLogin login) *Handler {
	return &Handler{login: loginLogin}
}

func (h *Handler) Register(ctx *fiber.Ctx) error {
	var req UserLoginIn

	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(models.ErrorResp{
			Message: "invalid request body",
		})
	}

	email, password, err := req.validate()
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(models.ErrorResp{
			Message: fmt.Sprintf("invalid request body: %s", err.Error()),
		})
	}

	user, err := h.login.LoginUser(ctx.Context(), models.User{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(models.ErrorResp{
			Message: fmt.Sprintf("login failed: %s", err.Error()),
		})
	}

	ctx.Locals("UserID", user.ID)
	ctx.Locals("Role", user.Role)

	return ctx.Next()
}
