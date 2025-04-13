package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/spanwalla/pvz/internal/dto"
	"github.com/spanwalla/pvz/internal/entity"
	"github.com/spanwalla/pvz/internal/metrics"
	"github.com/spanwalla/pvz/internal/repository"
)

var (
	ErrCityNotFound            = errors.New("city not found")
	ErrCannotCreatePoint       = errors.New("cannot create point")
	ErrActiveReceptionNotFound = errors.New("active reception not found")
	ErrProductNotFound         = errors.New("product not found")
	ErrCannotCloseReception    = errors.New("cannot close reception")
	ErrCannotDeleteLastProduct = errors.New("cannot delete last product")
	ErrProductAlreadyDeleted   = errors.New("product already deleted")
	ErrCannotGetPoints         = errors.New("cannot get points")
)

type PointService struct {
	pointRepo     repository.Point
	productRepo   repository.Product
	receptionRepo repository.Reception
	pointsCreated metrics.Counter
}

func NewPointService(pointRepo repository.Point, productRepo repository.Product, receptionRepo repository.Reception, pointsCreated metrics.Counter) *PointService {
	return &PointService{
		pointRepo:     pointRepo,
		productRepo:   productRepo,
		receptionRepo: receptionRepo,
		pointsCreated: pointsCreated,
	}
}

func (s *PointService) Create(ctx context.Context, city string) (entity.Point, error) {
	point, err := s.pointRepo.Create(ctx, city)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.Point{}, ErrCityNotFound
		}

		log.Errorf("PointService.Create - s.pointRepo.Create: %v", err)
		return entity.Point{}, ErrCannotCreatePoint
	}

	s.pointsCreated.Inc()
	return point, nil
}

func (s *PointService) GetAll(ctx context.Context) ([]entity.Point, error) {
	points, err := s.pointRepo.GetAll(ctx)
	if err != nil {
		log.Errorf("PointService.GetAll - s.pointRepo.GetAll: %v", err)
		return []entity.Point{}, ErrCannotGetPoints
	}

	return points, nil
}

func (s *PointService) GetExtended(ctx context.Context, start, end *time.Time, pagePtr, limitPtr *int) ([]dto.PointOutput, error) {
	limit := DefaultLimit
	if limitPtr != nil && *limitPtr > 0 {
		limit = *limitPtr
	}

	page := DefaultPage
	if pagePtr != nil && *pagePtr > 0 {
		page = *pagePtr
	}

	offset := (page - 1) * limit

	points, err := s.pointRepo.GetExtended(ctx, start, end, offset, limit)
	if err != nil {
		log.Errorf("PointService.GetExtended - s.pointRepo.GetExtended: %v", err)
		return []dto.PointOutput{}, ErrCannotGetPoints
	}

	return points, nil
}

func (s *PointService) CloseLastReception(ctx context.Context, pointID uuid.UUID) (entity.Reception, error) {
	receptionID, err := s.receptionRepo.GetActiveID(ctx, pointID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return entity.Reception{}, ErrActiveReceptionNotFound
		}

		log.Errorf("PointService.CloseLastReception - s.receptionRepo.GetActiveID: %v", err)
		return entity.Reception{}, ErrCannotCloseReception
	}

	log.Debugf("PointService.CloseLastReception - receptionID: %v", receptionID)

	reception, err := s.receptionRepo.Close(ctx, receptionID)
	if err != nil {
		log.Errorf("PointService.CloseLastReception - s.receptionRepo.Close: %v", err)
		return entity.Reception{}, ErrCannotCloseReception
	}

	return reception, nil
}

func (s *PointService) DeleteLastProduct(ctx context.Context, pointID uuid.UUID) error {
	receptionID, err := s.receptionRepo.GetActiveID(ctx, pointID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrActiveReceptionNotFound
		}

		log.Errorf("PointService.DeleteLastProduct - s.receptionRepo.GetActiveID: %v", err)
		return ErrCannotDeleteLastProduct
	}

	log.Debugf("PointService.DeleteLastProduct - receptionID: %v", receptionID)

	productID, err := s.productRepo.GetLatestID(ctx, receptionID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrProductNotFound
		}

		log.Errorf("PointService.DeleteLastProduct - s.productRepo.GetLatestID: %v", err)
		return ErrCannotDeleteLastProduct
	}

	log.Debugf("PointService.DeleteLastProduct - productID: %v", productID)

	err = s.productRepo.DeleteByID(ctx, productID)
	if err != nil {
		if errors.Is(err, repository.ErrNoRowsDeleted) {
			return ErrProductAlreadyDeleted
		}

		log.Errorf("PointService.DeleteLastProduct - s.productRepo.DeleteByID: %v", err)
		return ErrCannotDeleteLastProduct
	}

	return nil
}
