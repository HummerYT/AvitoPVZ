package models

import "github.com/google/uuid"

type UserRole string

const (
	RoleEmployee  UserRole = "employee"
	RoleModerator UserRole = "moderator"
)

type User struct {
	ID       uuid.UUID
	Email    string
	Password string
	Role     UserRole
}

func IsUserRole(role UserRole) bool {
	return role == RoleEmployee || role == RoleModerator
}
