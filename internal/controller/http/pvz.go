package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/spanwalla/pvz/internal/controller/http/dto"
	"github.com/spanwalla/pvz/internal/controller/http/mw"
	"github.com/spanwalla/pvz/internal/entity"
	"github.com/spanwalla/pvz/internal/service"
)

type pvzPostRequest struct {
	City string `json:"city" validate:"required,oneof=Москва Санкт-Петербург Казань"`
}

type pvzGetRequest struct {
	StartDate *time.Time `query:"startDate"`
	EndDate   *time.Time `query:"endDate"`
	Page      *int       `query:"page" validate:"omitnil,gte=1"`
	Limit     *int       `query:"limit" validate:"omitnil,gte=1,lte=30"`
}

type receptionResult struct {
	Reception dto.Reception `json:"reception"`
	Products  []dto.Product `json:"products"`
}

type pvzGetResponse struct {
	Pvz        dto.PVZ           `json:"pvz"`
	Receptions []receptionResult `json:"receptions"`
}

type closeLastReceptionRequest struct {
	PointID uuid.UUID `param:"pvzId" validate:"required,uuid"`
}

type deleteLastProductRequest struct {
	PointID uuid.UUID `param:"pvzId" validate:"required,uuid"`
}

type pvzRoutes struct {
	pointService service.Point
}

func newPvzRoutes(g *echo.Group, pointService service.Point, authMW *mw.Auth) {
	r := &pvzRoutes{pointService: pointService}

	g.POST("", r.postRoot, authMW.CheckRole(entity.RoleTypeModerator))
	g.GET("", r.getRoot)
	g.POST("/:pvzId/close_last_reception", r.closeLastReception, authMW.CheckRole(entity.RoleTypeEmployee))
	g.POST("/:pvzId/delete_last_product", r.deleteLastProduct, authMW.CheckRole(entity.RoleTypeEmployee))
}

func (r *pvzRoutes) postRoot(c echo.Context) error {
	var req pvzPostRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	point, err := r.pointService.Create(c.Request().Context(), req.City)
	if err != nil {
		if errors.Is(err, service.ErrCityNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, dto.PVZ{
		Id:               &point.ID,
		RegistrationDate: &point.CreatedAt,
		City:             dto.PVZCity(point.City),
	})
}

func (r *pvzRoutes) getRoot(c echo.Context) error {
	var req pvzGetRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	points, err := r.pointService.GetExtended(c.Request().Context(), req.StartDate, req.EndDate, req.Page, req.Limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	response := make([]pvzGetResponse, 0, len(points))
	for _, point := range points {
		receptions := make([]receptionResult, 0, len(point.Receptions))
		for _, reception := range point.Receptions {
			products := make([]dto.Product, 0, len(reception.Products))
			for _, product := range reception.Products {
				products = append(products, dto.Product{
					Id:          &product.ID,
					ReceptionId: product.ReceptionID,
					DateTime:    &product.CreatedAt,
					Type:        dto.ProductType(product.Type),
				})
			}

			receptions = append(receptions, receptionResult{
				Reception: dto.Reception{
					Id:       &reception.Reception.ID,
					DateTime: reception.Reception.CreatedAt,
					PvzId:    reception.Reception.PointID,
					Status:   dto.ReceptionStatus(reception.Reception.Status),
				},
				Products: products,
			})
		}

		response = append(response, pvzGetResponse{
			Pvz: dto.PVZ{
				Id:               &point.Point.ID,
				RegistrationDate: &point.Point.CreatedAt,
				City:             dto.PVZCity(point.Point.City),
			},
			Receptions: receptions,
		})
	}

	return c.JSON(http.StatusOK, response)
}

func (r *pvzRoutes) closeLastReception(c echo.Context) error {
	var req closeLastReceptionRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	reception, err := r.pointService.CloseLastReception(c.Request().Context(), req.PointID)
	if err != nil {
		if errors.Is(err, service.ErrActiveReceptionNotFound) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, dto.Reception{
		Id:       &reception.ID,
		DateTime: reception.CreatedAt,
		PvzId:    reception.PointID,
		Status:   dto.ReceptionStatus(reception.Status),
	})
}

func (r *pvzRoutes) deleteLastProduct(c echo.Context) error {
	var req deleteLastProductRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := r.pointService.DeleteLastProduct(c.Request().Context(), req.PointID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrActiveReceptionNotFound):
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrProductNotFound):
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrProductAlreadyDeleted):
			break
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
	}

	return c.NoContent(http.StatusOK)
}
