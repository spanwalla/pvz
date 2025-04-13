package entity

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID `db:"id"`
	Email    string    `db:"email"`
	Password string    `db:"password"`
	Role     RoleType  `db:"role"`
}

type RoleType string

const (
	RoleTypeModerator RoleType = "moderator"
	RoleTypeEmployee  RoleType = "employee"
)
