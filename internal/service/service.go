package service

import (
	"context"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/google/uuid"
	"github.com/jonboulle/clockwork"

	"github.com/spanwalla/pvz/internal/dto"
	"github.com/spanwalla/pvz/internal/entity"
	"github.com/spanwalla/pvz/internal/metrics"
	"github.com/spanwalla/pvz/internal/repository"
	"github.com/spanwalla/pvz/pkg/hasher"
)

//go:generate go tool mockgen -source=service.go -destination=mocks/mock_service.go -package=mocks

const (
	DefaultPage  = 1
	DefaultLimit = 10
)

type RegisterOutput struct {
	ID    uuid.UUID
	Email string
	Role  entity.RoleType
}

type Auth interface {
	DummyLogin(ctx context.Context, role entity.RoleType) (string, error)
	Login(ctx context.Context, email, password string) (string, error)
	Register(ctx context.Context, email, password string, role entity.RoleType) (RegisterOutput, error)
	ParseToken(token string) (*entity.TokenClaims, error)
}

type Point interface {
	Create(ctx context.Context, city string) (entity.Point, error)
	GetAll(ctx context.Context) ([]entity.Point, error)
	GetExtended(ctx context.Context, start, end *time.Time, pagePtr, limitPtr *int) ([]dto.PointOutput, error)
	CloseLastReception(ctx context.Context, pointID uuid.UUID) (entity.Reception, error)
	DeleteLastProduct(ctx context.Context, pointID uuid.UUID) error
}

type Product interface {
	Create(ctx context.Context, pointID uuid.UUID, productType entity.ProductType) (entity.Product, error)
}

type Reception interface {
	Create(ctx context.Context, pointID uuid.UUID) (entity.Reception, error)
}

type Services struct {
	Auth
	Point
	Product
	Reception
}

type Dependencies struct {
	Repos          *repository.Repositories
	Counters       *metrics.Counters
	Transaction    *manager.Manager
	PasswordHasher hasher.PasswordHasher
	Clock          clockwork.Clock
	SecretKey      string
	TokenTTL       time.Duration
}

func New(deps Dependencies) *Services {
	return &Services{
		Auth:      NewAuthService(deps.Repos.User, deps.PasswordHasher, deps.Clock, deps.SecretKey, deps.TokenTTL),
		Point:     NewPointService(deps.Repos.Point, deps.Repos.Product, deps.Repos.Reception, deps.Counters.PointsCreated),
		Product:   NewProductService(deps.Repos.Product, deps.Repos.Reception, deps.Counters.ProductsCreated),
		Reception: NewReceptionService(deps.Repos.Reception, deps.Counters.ReceptionsCreated),
	}
}
