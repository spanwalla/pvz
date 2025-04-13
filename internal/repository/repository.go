package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/spanwalla/pvz/internal/dto"
	"github.com/spanwalla/pvz/internal/entity"
	"github.com/spanwalla/pvz/pkg/postgres"
)

//go:generate go tool mockgen -source=repository.go -destination=mocks/mock_repository.go -package=mocks

type Point interface {
	Create(ctx context.Context, city string) (entity.Point, error)
	GetAll(ctx context.Context) ([]entity.Point, error)
	GetExtended(ctx context.Context, start, end *time.Time, offset, limit int) ([]dto.PointOutput, error)
}

type Product interface {
	Create(ctx context.Context, receptionID uuid.UUID, productType entity.ProductType) (entity.Product, error)
	GetLatestID(ctx context.Context, receptionID uuid.UUID) (uuid.UUID, error)
	DeleteByID(ctx context.Context, productID uuid.UUID) error
}

type Reception interface {
	Create(ctx context.Context, pointID uuid.UUID) (entity.Reception, error)
	GetActiveID(ctx context.Context, pointID uuid.UUID) (uuid.UUID, error)
	Close(ctx context.Context, receptionID uuid.UUID) (entity.Reception, error)
}

type User interface {
	Create(ctx context.Context, email, password string, role entity.RoleType) (entity.User, error)
	GetByEmail(ctx context.Context, email string) (entity.User, error)
}

type Repositories struct {
	Point
	Product
	Reception
	User
}

func New(pg *postgres.Postgres) *Repositories {
	return &Repositories{
		Point:     NewPointRepository(pg),
		Product:   NewProductRepository(pg),
		Reception: NewReceptionRepository(pg),
		User:      NewUserRepository(pg),
	}
}
