package login_test

import (
	"AvitoPVZ/internal/models"
	"AvitoPVZ/internal/usecase/auth/login"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// --- Мок БД ---
type mockDB struct {
	mock.Mock
}

func (m *mockDB) GetUserEmail(ctx context.Context, email string) (models.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(models.User), args.Error(1)
}

// --- Тестовая структура ---
type LoginUseCaseSuite struct {
	suite.Suite
	mockDB *mockDB
	uc     *login.UseCase
}

// --- Setup ---
func (s *LoginUseCaseSuite) SetupTest() {
	s.mockDB = new(mockDB)
	s.uc = login.NewUseCase(s.mockDB)

	// Заменяем реальную функцию сравнения пароля на стаб
	s.uc.CompareHashAndPassword = func(hash string, password string) (bool, error) {
		if hash == "hashed_pass" && password == "plain_pass" {
			return true, nil
		}
		return false, errors.New("incorrect password")
	}
}

// --- Успешный логин ---
func (s *LoginUseCaseSuite) Test_LoginUser_Success() {
	// Arrange
	email := "user@example.com"
	inputUser := models.User{
		Email:    email,
		Password: "plain_pass",
	}
	expectedUser := models.User{
		ID:       uuid.New(),
		Email:    email,
		Password: "hashed_pass",
		Role:     models.RoleClient,
	}

	s.mockDB.On("GetUserEmail", mock.Anything, email).Return(expectedUser, nil)

	// Act
	result, err := s.uc.LoginUser(context.Background(), inputUser)

	// Assert
	s.Require().NoError(err)
	s.Equal(expectedUser.ID, result.ID)
	s.Equal(expectedUser.Email, result.Email)
	s.Equal(expectedUser.Role, result.Role)
	s.mockDB.AssertExpectations(s.T())
}

// --- Ошибка: пользователь не найден ---
func (s *LoginUseCaseSuite) Test_LoginUser_UserNotFound() {
	// Arrange
	email := "missing@example.com"
	inputUser := models.User{
		Email:    email,
		Password: "plain_pass",
	}

	s.mockDB.On("GetUserEmail", mock.Anything, email).Return(models.User{}, errors.New("user not found"))

	// Act
	_, err := s.uc.LoginUser(context.Background(), inputUser)

	// Assert
	s.Require().Error(err)
	s.Contains(err.Error(), "failed get user by username")
	s.mockDB.AssertExpectations(s.T())
}

// --- Ошибка: неправильный пароль ---
func (s *LoginUseCaseSuite) Test_LoginUser_IncorrectPassword() {
	// Arrange
	email := "user@example.com"
	inputUser := models.User{
		Email:    email,
		Password: "wrong_pass",
	}
	dbUser := models.User{
		ID:       uuid.New(),
		Email:    email,
		Password: "hashed_pass", // ожидаем, что будет невалиден
		Role:     models.RoleEmployee,
	}

	s.mockDB.On("GetUserEmail", mock.Anything, email).Return(dbUser, nil)

	// Act
	_, err := s.uc.LoginUser(context.Background(), inputUser)

	// Assert
	s.Require().Error(err)
	s.Equal(login.ErrIncorrectPassword, err)
	s.mockDB.AssertExpectations(s.T())
}

// --- Запуск ---
func TestLoginUseCaseSuite(t *testing.T) {
	suite.Run(t, new(LoginUseCaseSuite))
}
