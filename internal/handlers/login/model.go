package login

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	passwordValidator "github.com/wagslane/go-password-validator"

	"AvitoPVZ/internal/models"
)

type userLoginIn struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,max=255"`
}

func (u *userLoginIn) validate() (string, string, error) {
	validate := validator.New()
	if err := validate.Struct(u); err != nil {
		return "", "", fmt.Errorf("%s: %w", models.ErrValidation, err)
	}

	if err := passwordValidator.Validate(u.Password, float64(models.MinEntropyBits)); err != nil {
		return "", "", fmt.Errorf("password is too simple: %w", err)
	}

	return u.Email, u.Password, nil
}
