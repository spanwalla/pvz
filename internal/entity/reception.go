package entity

import (
	"time"

	"github.com/google/uuid"
)

type Reception struct {
	ID        uuid.UUID       `db:"id"`
	PointID   uuid.UUID       `db:"point_id"`
	CreatedAt time.Time       `db:"created_at"`
	Status    ReceptionStatus `db:"status"`
}

type ReceptionStatus string

const (
	ReceptionStatusInProgress = "in_progress"
	ReceptionStatusClosed     = "close"
)
