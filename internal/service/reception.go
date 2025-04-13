package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"github.com/spanwalla/pvz/internal/entity"
	"github.com/spanwalla/pvz/internal/metrics"
	"github.com/spanwalla/pvz/internal/repository"
)

var (
	ErrReceptionAlreadyOpened = errors.New("reception already opened")
	ErrCannotCreateReception  = errors.New("cannot create reception")
)

type ReceptionService struct {
	receptionRepo     repository.Reception
	receptionsCreated metrics.Counter
}

func NewReceptionService(receptionRepo repository.Reception, receptionsCreated metrics.Counter) *ReceptionService {
	return &ReceptionService{
		receptionRepo:     receptionRepo,
		receptionsCreated: receptionsCreated,
	}
}

func (s *ReceptionService) Create(ctx context.Context, pointID uuid.UUID) (entity.Reception, error) {
	reception, err := s.receptionRepo.Create(ctx, pointID)
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return entity.Reception{}, ErrReceptionAlreadyOpened
		}

		log.Errorf("ReceptionService.Create - s.receptionRepo.Create: %v", err)
		return entity.Reception{}, ErrCannotCreateReception
	}

	s.receptionsCreated.Inc()
	return reception, nil
}
