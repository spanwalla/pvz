package grpc

import (
	"context"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/spanwalla/pvz/internal/controller/grpc/pvz_v1"
	"github.com/spanwalla/pvz/internal/service"
)

type PVZHandler struct {
	pointService service.Point
	pvz_v1.UnimplementedPVZServiceServer
}

func NewPVZHandler(pointService service.Point) *PVZHandler {
	return &PVZHandler{
		pointService: pointService,
	}
}

func (h *PVZHandler) GetPVZList(ctx context.Context, req *pvz_v1.GetPVZListRequest) (*pvz_v1.GetPVZListResponse, error) {
	points, err := h.pointService.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]*pvz_v1.PVZ, len(points))
	for i, point := range points {
		out[i] = &pvz_v1.PVZ{
			Id:               point.ID.String(),
			RegistrationDate: timestamppb.New(point.CreatedAt),
			City:             point.City,
		}
	}
	return &pvz_v1.GetPVZListResponse{Pvzs: out}, nil
}
