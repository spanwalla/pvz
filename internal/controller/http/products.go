package http

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/spanwalla/pvz/internal/controller/http/dto"
	"github.com/spanwalla/pvz/internal/controller/http/mw"
	"github.com/spanwalla/pvz/internal/entity"
	"github.com/spanwalla/pvz/internal/service"
)

type productRequest struct {
	Type    entity.ProductType `json:"type" validate:"required,oneof=электроника одежда обувь"`
	PointID uuid.UUID          `json:"pvzId" validate:"required,uuid"`
}

type productRoutes struct {
	productService service.Product
}

func newProductRoutes(g *echo.Group, productService service.Product, authMW *mw.Auth) {
	r := &productRoutes{productService}

	g.POST("", r.root, authMW.CheckRole(entity.RoleTypeEmployee))
}

func (r *productRoutes) root(c echo.Context) error {
	var req productRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	product, err := r.productService.Create(c.Request().Context(), req.PointID, req.Type)
	if err != nil {
		if errors.Is(err, service.ErrActiveReceptionNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.Product{
		Id:          &product.ID,
		DateTime:    &product.CreatedAt,
		Type:        dto.ProductType(product.Type),
		ReceptionId: product.ReceptionID,
	})
}
