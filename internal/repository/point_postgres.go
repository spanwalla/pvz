package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	log "github.com/sirupsen/logrus"

	"github.com/spanwalla/pvz/internal/dto"
	"github.com/spanwalla/pvz/internal/entity"
	"github.com/spanwalla/pvz/pkg/postgres"
)

type PointRepository struct {
	*postgres.Postgres
}

func NewPointRepository(pg *postgres.Postgres) *PointRepository {
	return &PointRepository{pg}
}

func (r *PointRepository) Create(ctx context.Context, city string) (entity.Point, error) {
	subQuery := r.Builder.
		Select("id").
		From("cities").
		Where("name = ?", city)

	sql, args, _ := r.Builder.
		Insert("points").
		Columns("city_id").
		Select(subQuery).
		Suffix("RETURNING id, created_at").
		ToSql()

	point := entity.Point{City: city}
	err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).QueryRow(ctx, sql, args...).Scan(
		&point.ID,
		&point.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Point{}, ErrNotFound
		}

		return entity.Point{}, fmt.Errorf("PointRepository.Create - QueryRow: %w", err)
	}

	return point, nil
}

func (r *PointRepository) GetAll(ctx context.Context) ([]entity.Point, error) {
	sql, args, _ := r.Builder.
		Select("points.id, created_at, cities.name AS city").
		From("points").
		InnerJoin("cities ON cities.id = points.city_id").
		ToSql()

	rows, err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("PointRepository.GetAll - Query: %w", err)
	}
	defer rows.Close()

	var points []entity.Point
	for rows.Next() {
		var point entity.Point
		if err = rows.Scan(&point.ID, &point.CreatedAt, &point.City); err != nil {
			return nil, fmt.Errorf("PointRepository.GetAll - rows.Scan: %w", err)
		}

		points = append(points, point)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("PointRepository.GetAll - rows.Err: %w", err)
	}

	return points, nil
}

func (r *PointRepository) GetExtended(ctx context.Context, start, end *time.Time, offset, limit int) ([]dto.PointOutput, error) {
	cte := r.Builder.
		Select(
			"r.id AS reception_id",
			"r.point_id",
			"r.created_at",
			"r.status",
			`COALESCE(
				json_agg(
					json_build_object(
						'id', p.id,
						'receptionId', p.reception_id,
						'createdAt', p.created_at,
						'type', p.type
					)
				) FILTER (WHERE p.id IS NOT NULL), '[]'
			) AS products_json`,
		).
		From("receptions r").
		LeftJoin("products p ON r.id = p.reception_id").
		GroupBy("r.id", "r.point_id", "r.created_at", "r.status")

	if start != nil {
		cte = cte.Where("r.created_at >= ?", start)
	}

	if end != nil {
		cte = cte.Where("r.created_at <= ?", end)
	}

	cteSql, cteArgs, _ := cte.ToSql()

	cteFinal := fmt.Sprintf("WITH reception_products AS (%s)", cteSql)

	sql, args, _ := r.Builder.
		Select(
			"pts.id",
			"pts.created_at",
			"c.name AS city",
			`COALESCE(
				json_agg(
					json_build_object(
						'reception', json_build_object(
							'id', rp.reception_id,
							'pointId', pts.id, 
							'createdAt', rp.created_at,
							'status', rp.status
						),
						'products', rp.products_json
					)
				) FILTER (WHERE rp.reception_id IS NOT NULL), '[]'
			) AS receptions`,
		).
		From("points pts").
		InnerJoin("cities c ON c.id = pts.city_id").
		LeftJoin("reception_products rp ON rp.point_id = pts.id").
		GroupBy("pts.id", "pts.created_at", "c.name").
		Offset(uint64(offset)).
		Limit(uint64(limit)).
		Prefix(cteFinal).
		ToSql()

	cteArgs = append(cteArgs, args...)

	log.Debugf("PointRepository.GetExtended - sql, args: %v %v", sql, cteArgs)

	rows, err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).Query(ctx, sql, cteArgs...)
	if err != nil {
		return nil, fmt.Errorf("PointRepository.GetExtended - Query: %w", err)
	}
	defer rows.Close()

	var results []dto.PointOutput

	for rows.Next() {
		var (
			point   dto.Point
			rawJSON []byte
		)

		if err = rows.Scan(&point.ID, &point.CreatedAt, &point.City, &rawJSON); err != nil {
			return nil, fmt.Errorf("PointRepository.GetExtended - rows.Scan: %w", err)
		}

		var receptions []dto.ReceptionResult
		if err = json.Unmarshal(rawJSON, &receptions); err != nil {
			return nil, fmt.Errorf("PointRepository.GetExtended - json.Unmarshal: %w", err)
		}

		results = append(results, dto.PointOutput{
			Point:      point,
			Receptions: receptions,
		})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("PointRepository.GetExtended - rows.Err: %w", err)
	}

	return results, nil
}
