package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/spanwalla/pvz/internal/entity"
	"github.com/spanwalla/pvz/pkg/postgres"
)

type ReceptionRepository struct {
	*postgres.Postgres
}

func NewReceptionRepository(pg *postgres.Postgres) *ReceptionRepository {
	return &ReceptionRepository{pg}
}

func (r *ReceptionRepository) Create(ctx context.Context, pointID uuid.UUID) (entity.Reception, error) {
	sql, args, _ := r.Builder.
		Insert("receptions").
		Columns("point_id").
		Values(pointID).
		Suffix("RETURNING id, created_at, status").
		ToSql()

	reception := entity.Reception{PointID: pointID}
	err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).QueryRow(ctx, sql, args...).Scan(
		&reception.ID,
		&reception.CreatedAt,
		&reception.Status,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return entity.Reception{}, ErrAlreadyExists
			}
		}

		return entity.Reception{}, fmt.Errorf("ReceptionRepository.Create - QueryRow: %w", err)
	}

	return reception, nil
}

func (r *ReceptionRepository) GetActiveID(ctx context.Context, pointID uuid.UUID) (uuid.UUID, error) {
	sql, args, _ := r.Builder.
		Select("id").
		From("receptions").
		Where("status = ?", entity.ReceptionStatusInProgress).
		Where("point_id = ?", pointID).
		Limit(1).
		ToSql()

	var receptionID uuid.UUID

	err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).QueryRow(ctx, sql, args...).Scan(&receptionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrNotFound
		}

		return uuid.Nil, fmt.Errorf("ReceptionRepository.GetActiveID - QueryRow: %w", err)
	}

	return receptionID, nil
}

func (r *ReceptionRepository) Close(ctx context.Context, receptionID uuid.UUID) (entity.Reception, error) {
	sql, args, _ := r.Builder.
		Update("receptions").
		Set("status", entity.ReceptionStatusClosed).
		Where("id = ?", receptionID).
		Suffix("RETURNING point_id, created_at").
		ToSql()

	reception := entity.Reception{ID: receptionID, Status: entity.ReceptionStatusClosed}
	err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).QueryRow(ctx, sql, args...).Scan(
		&reception.PointID,
		&reception.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Reception{}, ErrNotFound
		}

		return entity.Reception{}, fmt.Errorf("ReceptionRepository.Close - QueryRow: %w", err)
	}

	return reception, nil
}
