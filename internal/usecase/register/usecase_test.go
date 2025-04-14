package register_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"AvitoPVZ/internal/models"
	"AvitoPVZ/internal/usecase/register"
)

type mockInsert struct {
	mock.Mock
}

func (m *mockInsert) InsertUser(ctx context.Context, user models.User) (string, error) {
	args := m.Called(ctx, user)
	return args.String(0), args.Error(1)
}

type RegisterUseCaseTestSuite struct {
	suite.Suite
	insert *mockInsert
	uc     *register.UseCase
}

func (s *RegisterUseCaseTestSuite) SetupTest() {
	s.insert = new(mockInsert)
	s.uc = register.NewUseCase(s.insert)

	s.uc.CreateHashPassword = func(password string) (string, error) {
		return "hashed_" + password, nil
	}
}

func (s *RegisterUseCaseTestSuite) Test_RegisterUser_Success() {
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

	result, err := s.uc.RegisterUser(context.Background(), user)

	s.Require().NoError(err)
	s.Equal(expectedID, result)
	s.insert.AssertExpectations(s.T())
}

func (s *RegisterUseCaseTestSuite) Test_RegisterUser_HashError() {
	user := models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: "failhash",
		Role:     models.RoleModerator,
	}

	s.uc.CreateHashPassword = func(password string) (string, error) {
		return "", errors.New("hashing failed")
	}

	result, err := s.uc.RegisterUser(context.Background(), user)

	s.Require().Error(err)
	s.Empty(result)
	s.Contains(err.Error(), "failed generate password")
}

func (s *RegisterUseCaseTestSuite) Test_RegisterUser_InsertError() {
	user := models.User{
		ID:       uuid.New(),
		Email:    "user@example.com",
		Password: "mypassword",
		Role:     models.RoleModerator,
	}
	userWithHashedPassword := user
	userWithHashedPassword.Password = "hashed_" + user.Password

	s.insert.On("InsertUser", mock.Anything, userWithHashedPassword).Return("", errors.New("insert error"))

	result, err := s.uc.RegisterUser(context.Background(), user)

	s.Require().Error(err)
	s.Empty(result)
	s.Contains(err.Error(), "failed of create user")
	s.insert.AssertExpectations(s.T())
}

func TestRegisterUseCaseTestSuite(t *testing.T) {
	suite.Run(t, new(RegisterUseCaseTestSuite))
}
