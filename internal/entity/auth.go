package entity

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenClaims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"userId"`
	Role   RoleType  `json:"role"`
}
