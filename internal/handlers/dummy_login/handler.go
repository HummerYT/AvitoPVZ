package dummy_login

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"AvitoPVZ/internal/models"
)

func DummyLoginHandler(c *fiber.Ctx) error {
	var req struct {
		Role string `json:"role"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResp{
			Message: "Invalid request body",
		})
	}

	if !models.IsUserRole(models.UserRole(req.Role)) {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResp{
			Message: "Role must be 'employee' or 'moderator'",
		})
	}

	userID := uuid.New()

	c.Locals("UserID", userID)
	c.Locals("Role", models.UserRole(req.Role))

	return c.Next()
}
