package http

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/spanwalla/pvz/internal/controller/http/dto"
	"github.com/spanwalla/pvz/internal/entity"
	"github.com/spanwalla/pvz/internal/service"
)

type authRoutes struct {
	authService service.Auth
}

type dummyLoginRequest struct {
	Role entity.RoleType `json:"role" validate:"required,oneof=employee moderator"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

type registerRequest struct {
	Email    string          `json:"email" validate:"required,email,max=255"`
	Password string          `json:"password" validate:"required,min=8,max=64"`
	Role     entity.RoleType `json:"role" validate:"required,oneof=employee moderator"`
}

func newAuthRoutes(g *echo.Group, authService service.Auth) {
	r := &authRoutes{authService}

	g.POST("/dummyLogin", r.dummyLogin)
	g.POST("/register", r.register)
	g.POST("/login", r.login)
}

func (r *authRoutes) dummyLogin(c echo.Context) error {
	var req dummyLoginRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	token, err := r.authService.DummyLogin(c.Request().Context(), req.Role)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, token)
}

func (r *authRoutes) register(c echo.Context) error {
	var req registerRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	user, err := r.authService.Register(c.Request().Context(), req.Email, req.Password, req.Role)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.User{
		Id:    &user.ID,
		Email: openapi_types.Email(user.Email),
		Role:  dto.UserRole(user.Role),
	})
}

func (r *authRoutes) login(c echo.Context) error {
	var req loginRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	token, err := r.authService.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) || errors.Is(err, service.ErrWrongPassword) {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		}

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, token)
}
