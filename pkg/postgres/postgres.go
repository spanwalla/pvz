package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

const (
	defaultMaxPoolSize  = 1
	defaultConnAttempts = 10
	defaultConnTimeout  = time.Second
)

type Postgres struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration

	Builder   squirrel.StatementBuilderType
	Pool      *pgxpool.Pool
	CtxGetter *trmpgx.CtxGetter
}

func New(url string, opts ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxPoolSize:  defaultMaxPoolSize,
		connAttempts: defaultConnAttempts,
		connTimeout:  defaultConnTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(pg)
	}

	pg.Builder = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	pg.CtxGetter = trmpgx.DefaultCtxGetter

	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("postgres - NewPostgres - pgxpool.ParseConfig: %w", err)
	}

	poolConfig.MaxConns = int32(pg.maxPoolSize)

	for pg.connAttempts > 0 {
		pg.Pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
		if err != nil {
			return nil, fmt.Errorf("postgres - NewPostgres - pgxpool.NewWithConfig: %w", err)
		}

		if err = pg.Pool.Ping(context.Background()); err == nil {
			break
		}

		log.Infof("Postgres is trying to connect, attempts left: %d", pg.connAttempts)
		pg.connAttempts--
		time.Sleep(pg.connTimeout)
	}

	if err != nil {
		return nil, fmt.Errorf("postgres - NewPostgres - connAtempts == 0: %w", err)
	}

	return pg, nil
}

func (pg *Postgres) Close() {
	if pg.Pool != nil {
		pg.Pool.Close()
	}
}
