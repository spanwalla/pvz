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
    ErrCannotCreateProduct = errors.New("cannot create product")
)

type ProductService struct {
    productRepo     repository.Product
    receptionRepo   repository.Reception
    productsCreated metrics.Counter
}

func NewProductService(productRepo repository.Product, receptionRepo repository.Reception, productsCreated metrics.Counter) *ProductService {
    return &ProductService{
        productRepo:     productRepo,
        receptionRepo:   receptionRepo,
        productsCreated: productsCreated,
    }
}

func (s *ProductService) Create(ctx context.Context, pointID uuid.UUID, productType entity.ProductType) (entity.Product, error) {
    receptionID, err := s.receptionRepo.GetActiveID(ctx, pointID)
    if err != nil {
        if errors.Is(err, repository.ErrNotFound) {
            return entity.Product{}, ErrActiveReceptionNotFound
        }

        log.Errorf("ProductService.Create - s.receptionRepo.GetActiveID: %v", err)
        return entity.Product{}, ErrCannotCreateProduct
    }

    log.Debugf("ProductService.Create - receptionID: %v", receptionID)

    product, err := s.productRepo.Create(ctx, receptionID, productType)
    if err != nil {
        log.Errorf("ProductService.Create - s.productRepo.Create: %v", err)
        return entity.Product{}, ErrCannotCreateProduct
    }

    s.productsCreated.Inc()
    return product, nil
}
