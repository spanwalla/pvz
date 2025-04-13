package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/spanwalla/pvz/internal/entity"
	"github.com/spanwalla/pvz/pkg/postgres"
)

type UserRepository struct {
	*postgres.Postgres
}

func NewUserRepository(pg *postgres.Postgres) *UserRepository {
	return &UserRepository{pg}
}

func (r *UserRepository) Create(ctx context.Context, email, password string, role entity.RoleType) (entity.User, error) {
	sql, args, _ := r.Builder.
		Insert("users").
		Columns("email, password, role").
		Values(email, password, role).
		Suffix("RETURNING id").
		ToSql()

	var user entity.User
	err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).QueryRow(ctx, sql, args...).Scan(&user.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if ok := errors.As(err, &pgErr); ok {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return entity.User{}, ErrAlreadyExists
			}
		}

		return entity.User{}, fmt.Errorf("UserRepository.Create - QueryRow: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (entity.User, error) {
	sql, args, _ := r.Builder.
		Select("id, email, password, role").
		From("users").
		Where("email = ?", email).
		ToSql()

	var user entity.User
	err := r.CtxGetter.DefaultTrOrDB(ctx, r.Pool).QueryRow(ctx, sql, args...).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Role,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, ErrNotFound
		}

		return entity.User{}, fmt.Errorf("UserRepository.GetByEmail - QueryRow: %w", err)
	}

	return user, nil
}
