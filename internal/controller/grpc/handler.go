package grpc

import (
	"google.golang.org/grpc"

	"github.com/spanwalla/pvz/internal/controller/grpc/pvz_v1"
	"github.com/spanwalla/pvz/internal/service"
)

func ConfigureHandler(server *grpc.Server, services *service.Services) {
	pvz_v1.RegisterPVZServiceServer(server, NewPVZHandler(services.Point))
}
