package auth

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"testing"

	"AvitoPVZ/internal/models"
	"AvitoPVZ/internal/repository/auth/mocks"
)

type fakeRow struct {
	scanFunc func(dest ...interface{}) error
}

func (f fakeRow) Scan(dest ...interface{}) error {
	return f.scanFunc(dest...)
}

type RepositorySuite struct {
	suite.Suite
	ctrl *gomock.Controller
	pool *mocks.Mockpool
	repo *Repository
}

func (s *RepositorySuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.pool = mocks.NewMockpool(s.ctrl)
	s.repo = NewInsertRepo(s.pool)
}

func (s *RepositorySuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *RepositorySuite) TestGetUserEmail_Success() {
	expectedUser := models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Password: "secret",
		Role:     "user",
	}
	ctx := context.Background()
	query := `SELECT id, email, password, role FROM users WHERE email = $1 LIMIT 1`

	s.pool.EXPECT().
		QueryRow(ctx, query, expectedUser.Email).
		Return(fakeRow{
			scanFunc: func(dest ...interface{}) error {
				*(dest[0].(*uuid.UUID)) = expectedUser.ID
				*(dest[1].(*string)) = expectedUser.Email
				*(dest[2].(*string)) = expectedUser.Password
				*(dest[3].(*models.UserRole)) = expectedUser.Role
				return nil
			},
		})

	user, err := s.repo.GetUserEmail(ctx, expectedUser.Email)
	s.NoError(err)
	s.Equal(expectedUser, user)
}

func (s *RepositorySuite) TestGetUserEmail_Error() {
	ctx := context.Background()
	email := "missing@example.com"
	expectedError := fmt.Errorf("scan error")
	query := `SELECT id, email, password, role FROM users WHERE email = $1 LIMIT 1`

	s.pool.EXPECT().
		QueryRow(ctx, query, email).
		Return(fakeRow{
			scanFunc: func(dest ...interface{}) error {
				return expectedError
			},
		})

	_, err := s.repo.GetUserEmail(ctx, email)
	s.Error(err)
	s.Contains(err.Error(), "failed to scan user")
}

func (s *RepositorySuite) TestInsertUser_Success() {
	newUser := models.User{
		ID:       uuid.New(),
		Email:    "insert@example.com",
		Password: "password",
		Role:     "admin",
	}
	ctx := context.Background()
	expectedQuery := `
        INSERT INTO users (id, email, password, role)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `

	s.pool.EXPECT().
		QueryRow(ctx, expectedQuery, newUser.ID, newUser.Email, newUser.Password, newUser.Role).
		Return(fakeRow{
			scanFunc: func(dest ...interface{}) error {
				*(dest[0].(*string)) = newUser.ID.String()
				return nil
			},
		})

	id, err := s.repo.InsertUser(ctx, newUser)
	s.NoError(err)
	s.Equal(newUser.ID.String(), id)
}

func (s *RepositorySuite) TestInsertUser_Error() {
	newUser := models.User{
		ID:       uuid.New(),
		Email:    "fail@example.com",
		Password: "password",
		Role:     "user",
	}
	ctx := context.Background()
	expectedQuery := `
        INSERT INTO users (id, email, password, role)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `
	expectedError := fmt.Errorf("insert error")

	s.pool.EXPECT().
		QueryRow(ctx, expectedQuery, newUser.ID, newUser.Email, newUser.Password, newUser.Role).
		Return(fakeRow{
			scanFunc: func(dest ...interface{}) error {
				return expectedError
			},
		})

	id, err := s.repo.InsertUser(ctx, newUser)
	s.Error(err)
	s.Empty(id)
	s.Contains(err.Error(), "failed to insert user")
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositorySuite))
}
