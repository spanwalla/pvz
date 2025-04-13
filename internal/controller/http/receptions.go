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

type receptionRequest struct {
	PointID uuid.UUID `json:"pvzId" validate:"required,uuid"`
}

type receptionRoutes struct {
	receptionService service.Reception
}

func newReceptionRoutes(g *echo.Group, receptionService service.Reception, authMW *mw.Auth) {
	r := &receptionRoutes{receptionService}

	g.POST("", r.root, authMW.CheckRole(entity.RoleTypeEmployee))
}

func (r *receptionRoutes) root(c echo.Context) error {
	var req receptionRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	reception, err := r.receptionService.Create(c.Request().Context(), req.PointID)
	if err != nil {
		if errors.Is(err, service.ErrReceptionAlreadyOpened) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.Reception{
		Id:       &reception.ID,
		DateTime: reception.CreatedAt,
		PvzId:    reception.PointID,
		Status:   dto.ReceptionStatus(reception.Status),
	})
}
