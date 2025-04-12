package register

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	passwordValidator "github.com/wagslane/go-password-validator"

	"AvitoPVZ/internal/models"
)

type userAuthIn struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,max=255"`
	Role     string `json:"role" validate:"required"`
}

func (u *userAuthIn) validate() (models.User, error) {
	validate := validator.New()
	if err := validate.Struct(u); err != nil {
		return models.User{}, fmt.Errorf("%s: %w", models.ErrValidation, err)
	}

	if err := passwordValidator.Validate(u.Password, float64(models.MinEntropyBits)); err != nil {
		return models.User{}, fmt.Errorf("password is too simple: %w", err)
	}

	if !models.IsUserRole(models.UserRole(u.Role)) {
		return models.User{}, fmt.Errorf("%s is not a valid role", u.Role)
	}

	return models.User{
		ID:       uuid.New(),
		Email:    u.Email,
		Password: u.Password,
		Role:     models.UserRole(u.Role),
	}, nil
}
