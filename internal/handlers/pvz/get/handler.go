package get

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"AvitoPVZ/internal/models"
)

type PVZDataUseCase interface {
	GetPVZData(ctx context.Context, startDate, endDate *time.Time, page, limit int) ([]models.PVZData, error)
}

type PVZDataHandler struct {
	UC PVZDataUseCase
}

func NewPVZDataHandler(uc PVZDataUseCase) *PVZDataHandler {
	return &PVZDataHandler{UC: uc}
}

func (h *PVZDataHandler) GetPVZData(c *fiber.Ctx) error {
	userRole, ok := c.Locals("Role").(models.UserRole)
	if !ok || !models.IsUserRole(userRole) {
		return c.Status(http.StatusForbidden).JSON(models.ErrorResp{
			Message: "access denied",
		})
	}

	req := PVZDataRequest{
		StartDate: c.Query("startDate", ""),
		EndDate:   c.Query("endDate", ""),
		Page:      1,
		Limit:     10,
	}
	if pageStr := c.Query("page", ""); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			req.Page = p
		}
	}
	if limitStr := c.Query("limit", ""); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = l
		}
	}

	if err := validateGetPVZDataRequest(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: "Неверные параметры запроса: " + err.Error(),
		})
	}

	var startDatePtr, endDatePtr *time.Time
	if req.StartDate != "" {
		if t, err := time.Parse(time.RFC3339, req.StartDate); err == nil {
			startDatePtr = &t
		} else {
			return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{
				Message: "Неверный формат startDate",
			})
		}
	}
	if req.EndDate != "" {
		if t, err := time.Parse(time.RFC3339, req.EndDate); err == nil {
			endDatePtr = &t
		} else {
			return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{
				Message: "Неверный формат endDate",
			})
		}
	}

	data, err := h.UC.GetPVZData(c.Context(), startDatePtr, endDatePtr, req.Page, req.Limit)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.ErrorResp{
			Message: err.Error(),
		})
	}

	var result []fiber.Map
	for _, item := range data {
		var recs []fiber.Map
		for _, recData := range item.Receptions {
			recs = append(recs, fiber.Map{
				"reception": recData.Reception,
				"products":  recData.Products,
			})
		}

		result = append(result, fiber.Map{
			"pvz":        item.PVZ,
			"receptions": recs,
		})
	}

	return c.Status(http.StatusOK).JSON(result)
}
