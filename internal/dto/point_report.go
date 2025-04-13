package dto

import (
	"time"

	"github.com/google/uuid"

	"github.com/spanwalla/pvz/internal/entity"
)

type Point struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	City      string    `json:"city"`
}

type Product struct {
	ID          uuid.UUID          `json:"id"`
	ReceptionID uuid.UUID          `json:"receptionId"`
	CreatedAt   time.Time          `json:"createdAt"`
	Type        entity.ProductType `json:"type"`
}

type Reception struct {
	ID        uuid.UUID              `json:"id"`
	PointID   uuid.UUID              `json:"pointId"`
	CreatedAt time.Time              `json:"createdAt"`
	Status    entity.ReceptionStatus `json:"status"`
}

type PointOutput struct {
	Point      Point             `json:"point"`
	Receptions []ReceptionResult `json:"receptions"`
}

type ReceptionResult struct {
	Reception Reception `json:"reception"`
	Products  []Product `json:"products"`
}
