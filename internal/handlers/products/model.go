package products

import (
	"AvitoPVZ/internal/models"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type reqProducts struct {
	Type  string `json:"type" validate:"required"`
	PvzID string `json:"pvzId" validate:"required,uuid"`
}

func (u *reqProducts) validate() (models.TypeProduct, uuid.UUID, error) {
	validate := validator.New()
	if err := validate.Struct(u); err != nil {
		return "", uuid.UUID{}, fmt.Errorf("%s: %w", models.ErrValidation, err)
	}

	if !models.IsTypeProduct(u.Type) {
		return "", uuid.UUID{}, models.ErrValidation
	}

	pvzID, err := uuid.Parse(u.PvzID)
	if err != nil {
		return "", uuid.UUID{}, fmt.Errorf("%s: %w", models.ErrValidation, err)
	}

	return models.TypeProduct(u.Type), pvzID, nil
}
