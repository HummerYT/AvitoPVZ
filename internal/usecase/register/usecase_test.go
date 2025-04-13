package register_test

import (
	"AvitoPVZ/internal/models"
	"AvitoPVZ/internal/usecase/register"
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// --- Мок для insert-интерфейса ---
type mockInsert struct {
	mock.Mock
}

func (m *mockInsert) InsertUser(ctx context.Context, user models.User) (string, error) {
	args := m.Called(ctx, user)
	return args.String(0), args.Error(1)
}

// --- Тестовый suite ---
type RegisterUseCaseTestSuite struct {
	suite.Suite
	insert *mockInsert
	uc     *register.UseCase
}

// --- Setup ---
func (s *RegisterUseCaseTestSuite) SetupTest() {
	s.insert = new(mockInsert)
	s.uc = register.NewUseCase(s.insert)

	// Мокаем хеширование пароля
	s.uc.CreateHashPassword = func(password string) (string, error) {
		return "hashed_" + password, nil
	}
}

// --- Успешная регистрация ---
func (s *RegisterUseCaseTestSuite) Test_RegisterUser_Success() {
	// Arrange
	user := models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: "simplepassword",
		Role:     models.RoleModerator,
	}
	expectedID := "user-id-123"

	userWithHashedPassword := user
	userWithHashedPassword.Password = "hashed_" + user.Password

	s.insert.On("InsertUser", mock.Anything, userWithHashedPassword).Return(expectedID, nil)

	// Act
	result, err := s.uc.RegisterUser(context.Background(), user)

	// Assert
	s.Require().NoError(err)
	s.Equal(expectedID, result)
	s.insert.AssertExpectations(s.T())
}

// --- Ошибка при хешировании пароля ---
func (s *RegisterUseCaseTestSuite) Test_RegisterUser_HashError() {
	// Arrange
	user := models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: "failhash",
		Role:     models.RoleModerator,
	}

	s.uc.CreateHashPassword = func(password string) (string, error) {
		return "", errors.New("hashing failed")
	}

	// Act
	result, err := s.uc.RegisterUser(context.Background(), user)

	// Assert
	s.Require().Error(err)
	s.Empty(result)
	s.Contains(err.Error(), "failed generate password")
}

// --- Ошибка при вставке пользователя ---
func (s *RegisterUseCaseTestSuite) Test_RegisterUser_InsertError() {
	// Arrange
	user := models.User{
		ID:       uuid.New(),
		Email:    "user@example.com",
		Password: "mypassword",
		Role:     models.RoleModerator,
	}
	userWithHashedPassword := user
	userWithHashedPassword.Password = "hashed_" + user.Password

	s.insert.On("InsertUser", mock.Anything, userWithHashedPassword).Return("", errors.New("insert error"))

	// Act
	result, err := s.uc.RegisterUser(context.Background(), user)

	// Assert
	s.Require().Error(err)
	s.Empty(result)
	s.Contains(err.Error(), "failed of create user")
	s.insert.AssertExpectations(s.T())
}

// --- Запуск suite ---
func TestRegisterUseCaseTestSuite(t *testing.T) {
	suite.Run(t, new(RegisterUseCaseTestSuite))
}
