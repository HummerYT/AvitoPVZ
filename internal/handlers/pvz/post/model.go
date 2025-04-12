package post

import (
	"AvitoPVZ/internal/models"
	"fmt"
	"github.com/go-playground/validator/v10"
)

type pvzPostReq struct {
	City string `json:"city" validate:"required"`
}

func (u *pvzPostReq) validate() (models.PVZCity, error) {
	validate := validator.New()
	if err := validate.Struct(u); err != nil {
		return "", fmt.Errorf("%s: %w", models.ErrValidation, err)
	}

	if !models.IsPVZCity(models.PVZCity(u.City)) {
		return "", fmt.Errorf("%s: %w", models.ErrValidation, "city is not allowed")
	}

	return models.PVZCity(u.City), nil
}
