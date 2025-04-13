package integration_test

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	. "github.com/Eun/go-hit"
	log "github.com/sirupsen/logrus"

	"github.com/spanwalla/pvz/internal/entity"
)

const (
	host            = "app:8080"
	healthPath      = "http://" + host + "/health"
	defaultAttempts = 20
	basePath        = "http://" + host
)

func TestMain(m *testing.M) {
	err := healthCheck(defaultAttempts)
	if err != nil {
		panic(fmt.Errorf("integration tests: host %s is not available: %w", host, err))
	}

	log.Infof("integration tests: host %s is available", host)
	os.Exit(m.Run())
}

func healthCheck(attempts int) error {
	var err error

	for attempts > 0 {
		err = Do(Get(healthPath), Expect().Status().Equal(http.StatusOK))
		if err == nil {
			return nil
		}

		log.Infof("integration tests: host %s is not available, attempts left: %d", host, attempts)
		time.Sleep(time.Second)
		attempts--
	}

	return err
}

func dummyLogin(role entity.RoleType) (string, error) {
	var token string
	body := map[string]entity.RoleType{
		"role": role,
	}

	if err := Do(
		Post(basePath+"/dummyLogin"),
		Send().Headers("Content-Type").Add("application/json"),
		Send().Body().JSON(body),
		Expect().Status().Equal(http.StatusOK),
		Store().Response().Body().JSON().JQ(".").In(&token),
	); err != nil {
		return "", err
	}

	return token, nil
}
