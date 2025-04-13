package mw

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"

	"github.com/spanwalla/pvz/internal/entity"
	"github.com/spanwalla/pvz/internal/service"
)

var (
	ErrInvalidAuthHeader = errors.New("invalid auth header")
	ErrNoRights          = errors.New("no rights")
)

const (
	UserIDKey = "userID"
	RoleKey   = "role"
)

type Auth struct {
	authService service.Auth
}

func NewAuth(authService service.Auth) *Auth {
	return &Auth{
		authService: authService,
	}
}

// UserIdentity - middleware to check authorization
//
// - Parse `Authorization` header (expected format `Bearer <JWT token>`)
//
// - Sets {"userID": <uuid.UUID>, "role": <entity.RoleType>} in `c echo.Context`
func (m *Auth) UserIdentity() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, err := bearerToken(c.Request())
			if err != nil {
				log.Errorf("AuthMW.UserIdentity - bearerToken: %v", err)
				return echo.NewHTTPError(http.StatusUnauthorized, ErrInvalidAuthHeader.Error())
			}

			claims, err := m.authService.ParseToken(token)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
			}

			c.Set(UserIDKey, claims.UserID)
			c.Set(RoleKey, claims.Role)

			return next(c)
		}
	}
}

// CheckRole - middleware to check user role
//
// Expect that value from `c echo.Context` by "role" key is `entity.RoleTypeModerator`
func (m *Auth) CheckRole(required entity.RoleType) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, ok := c.Get(RoleKey).(entity.RoleType)
			if !ok || role != required {
				return echo.NewHTTPError(http.StatusForbidden, ErrNoRights.Error())
			}

			return next(c)
		}
	}
}

func bearerToken(req *http.Request) (string, error) {
	const prefix = "Bearer "

	header := req.Header.Get(echo.HeaderAuthorization)

	if len(header) == 0 {
		return "", ErrInvalidAuthHeader
	}

	if len(header) > len(prefix) && strings.EqualFold(header[:len(prefix)], prefix) {
		return header[len(prefix):], nil
	}

	return "", ErrInvalidAuthHeader
}
