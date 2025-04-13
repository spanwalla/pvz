package entity

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID   `db:"id"`
	ReceptionID uuid.UUID   `db:"reception_id"`
	CreatedAt   time.Time   `db:"created_at"`
	Type        ProductType `db:"type"`
}

type ProductType string

const (
	ProductTypeElectronics ProductType = "электроника"
	ProductTypeClothes     ProductType = "одежда"
	ProductTypeShoes       ProductType = "обувь"
)
