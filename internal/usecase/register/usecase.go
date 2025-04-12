package register

import (
	"AvitoPVZ/internal/utils"
	"context"
	"fmt"

	"AvitoPVZ/internal/models"
)

type insert interface {
	InsertUser(ctx context.Context, user models.User) (string, error)
}

type UseCase struct {
	insert             insert
	CreateHashPassword func(password string) (string, error)
}

func NewUseCase(insert insert) *UseCase {
	useCase := &UseCase{
		insert: insert,
	}

	if useCase.CreateHashPassword == nil {
		useCase.CreateHashPassword = utils.CreateHashPassword
	}

	return useCase
}

func (c *UseCase) RegisterUser(ctx context.Context, user models.User) (string, error) {
	hashPassword, err := c.CreateHashPassword(user.Password)
	if err != nil {
		return "", fmt.Errorf("failed generate password: %w", err)
	}

	user.Password = hashPassword

	userID, err := c.insert.InsertUser(ctx, user)
	if err != nil {
		return "", fmt.Errorf("failed of create user: %w", err)
	}

	return userID, nil
}
