//go:build migrate

package app

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	log "github.com/sirupsen/logrus"
)

const (
	defaultAttempts = 20
	defaultTimeout  = time.Second
)

func init() {
	databaseURL, ok := os.LookupEnv("PG_URL")
	if !ok || len(databaseURL) == 0 {
		log.Fatal("migrate: environment variable PG_URL not set")
	}

	databaseURL += "?sslmode=disable"

	var (
		attempts = defaultAttempts
		err      error
		m        *migrate.Migrate
	)

	for attempts > 0 {
		m, err = migrate.New("file://migrations", databaseURL)
		if err == nil {
			break
		}

		log.Infof("migrate: postgres is trying to connect, attempts left: %d", attempts)
		time.Sleep(defaultTimeout)
		attempts--
	}

	if err != nil {
		log.Fatalf("migrate: postgres connect error: %s", err)
	}

	err = m.Up()
	defer func(m *migrate.Migrate) {
		err, _ = m.Close()
		if err != nil {
			log.Fatalf("migrate: postgres close error: %s", err)
		}
	}(m)
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		panic(fmt.Errorf("migrate: up error: %w", err))
	}

	if errors.Is(err, migrate.ErrNoChange) {
		log.Infof("migrate: no database change")
		return
	}

	log.Info("migrate: up success")
}
