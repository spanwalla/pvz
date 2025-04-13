package http

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"

	"github.com/spanwalla/pvz/internal/controller/http/dto"
	"github.com/spanwalla/pvz/internal/controller/http/mw"
	"github.com/spanwalla/pvz/internal/service"
)

//go:generate go tool oapi-codegen --config=dto.cfg.yaml ../../../api/swagger.yaml

type Server struct {
}

func ConfigureRouter(handler *echo.Echo, services *service.Services) {
	swagger, err := dto.GetSwagger()
	if err != nil {
		panic(fmt.Errorf("app - ConfigureRouter - GetSwagger: %w", err))
	}
	swagger.Servers = nil

	handler.Use(middleware.CORS())
	handler.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}", "method":"${method}","uri":"${uri}", "status":${status},"error":"${error}"}` + "\n",
		Output: setLogsFile(),
	}))

	handler.Use(middleware.Recover())
	handler.Use(echoprometheus.NewMiddleware("app"))

	handler.GET("/health", func(c echo.Context) error { return c.NoContent(http.StatusOK) })

	authGroup := handler.Group("")
	newAuthRoutes(authGroup, services.Auth)

	authMW := mw.NewAuth(services.Auth)

	pvzGroup := handler.Group("/pvz", authMW.UserIdentity())
	newPvzRoutes(pvzGroup, services.Point, authMW)

	receptionsGroup := handler.Group("/receptions", authMW.UserIdentity())
	newReceptionRoutes(receptionsGroup, services.Reception, authMW)

	productsGroup := handler.Group("/products", authMW.UserIdentity())
	newProductRoutes(productsGroup, services.Product, authMW)
}

func setLogsFile() *os.File {
	file, err := os.OpenFile("/logs/requests.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		log.Fatalf("http - setLogsFile - os.OpenFile: %v", err)
	}

	return file
}
