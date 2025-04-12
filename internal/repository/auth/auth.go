package auth

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"

	"AvitoPVZ/internal/models"
)

type pool interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Repository struct {
	pool pool
}

func NewInsertRepo(pool pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) GetUserEmail(ctx context.Context, email string) (models.User, error) {
	var dbUser models.User
	query := `SELECT id, email, password, role FROM users WHERE email = $1 LIMIT 1`

	row := r.pool.QueryRow(ctx, query, email)
	err := row.Scan(&dbUser.ID, &dbUser.Email, &dbUser.Password, &dbUser.Role)
	if err != nil {
		return dbUser, fmt.Errorf("failed to scan user: %w", err)
	}

	return dbUser, nil
}

func (r *Repository) InsertUser(ctx context.Context, user models.User) (string, error) {
	var userID string
	query := `
        INSERT INTO users (id, email, password, role)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `
	err := r.pool.QueryRow(ctx, query, user.ID, user.Email, user.Password, user.Role).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("failed to insert user: %w", err)
	}

	return userID, nil
}
