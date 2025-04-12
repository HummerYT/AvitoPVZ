package login

import (
	"AvitoPVZ/internal/utils"
	"context"
	"errors"
	"fmt"

	"AvitoPVZ/internal/models"
)

var ErrIncorrectPassword = errors.New("incorrect password")

type db interface {
	GetUserEmail(ctx context.Context, email string) (models.User, error)
}

type UseCase struct {
	db                     db
	CompareHashAndPassword func(hash string, password string) (bool, error)
}

func NewUseCase(db db) *UseCase {
	useCase := &UseCase{
		db: db,
	}

	if useCase.CompareHashAndPassword == nil {
		useCase.CompareHashAndPassword = utils.CompareHashAndPassword
	}

	return useCase
}

func (c *UseCase) LoginUser(ctx context.Context, user models.User) (models.User, error) {
	dbUser, err := c.db.GetUserEmail(ctx, user.Email)
	if err != nil {
		return models.User{}, fmt.Errorf("failed get user by username: %w", err)
	}

	_, err = c.CompareHashAndPassword(dbUser.Password, user.Password)
	if err != nil {
		return models.User{}, ErrIncorrectPassword
	}

	return dbUser, nil

}
