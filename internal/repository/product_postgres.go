package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/spanwalla/pvz/internal/entity"
	"github.com/spanwalla/pvz/pkg/postgres"
)

type ProductRepository struct {
	*postgres.Postgres
}

func NewProductRepository(pg *postgres.Postgres) *ProductRepository {
	return &ProductRepository{pg}
}

func (r *ProductRepository) Create(ctx context.Context, receptionID uuid.UUID, productType entity.ProductType) (entity.Product, error) {
	sql, args, _ := r.Builder.
		Insert("products").
		Columns("reception_id, type").
		Values(receptionID, productType).
		Suffix("RETURNING id, created_at").
		ToSql()

	product := entity.Product{ReceptionID: receptionID, Type: productType}
	err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).QueryRow(ctx, sql, args...).Scan(
		&product.ID,
		&product.CreatedAt,
	)
	if err != nil {
		return entity.Product{}, fmt.Errorf("ProductRepository.Create - QueryRow: %w", err)
	}

	return product, nil
}

func (r *ProductRepository) GetLatestID(ctx context.Context, receptionID uuid.UUID) (uuid.UUID, error) {
	sql, args, _ := r.Builder.
		Select("id").
		From("products").
		Where("reception_id = ?", receptionID).
		OrderBy("created_at DESC").
		Limit(1).
		ToSql()

	var productID uuid.UUID
	err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).QueryRow(ctx, sql, args...).Scan(&productID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrNotFound
		}

		return uuid.Nil, fmt.Errorf("ProductRepository.GetLatestID - QueryRow: %w", err)
	}

	return productID, nil
}

func (r *ProductRepository) DeleteByID(ctx context.Context, productID uuid.UUID) error {
	sql, args, _ := r.Builder.
		Delete("products").
		Where("id = ?", productID).
		ToSql()

	cmdTag, err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("ProductRepository.DeleteByID - Exec: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrNoRowsDeleted
	}

	return nil
}
