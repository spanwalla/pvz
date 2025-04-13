package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		App        App        `yaml:"app"`
		GRPC       GRPC       `yaml:"grpc"`
		HTTP       HTTP       `yaml:"http"`
		Prometheus Prometheus `yaml:"prometheus"`
		Log        Log        `yaml:"logger"`
		PG         PG         `yaml:"postgres"`
		Auth       Auth       `yaml:"auth"`
	}

	App struct {
		Name    string `env-required:"true" yaml:"name" env:"APP_NAME"`
		Version string `env-required:"true" yaml:"version" env:"APP_VERSION"`
	}

	GRPC struct {
		Port string `env-required:"true" yaml:"port" env:"GRPC_PORT"`
	}

	HTTP struct {
		Port string `env-required:"true" yaml:"port" env:"HTTP_PORT"`
	}

	Prometheus struct {
		Port string `env-required:"true" yaml:"port" env:"PROMETHEUS_PORT"`
	}

	Log struct {
		Level string `env-required:"true" yaml:"level" env:"LOG_LEVEL"`
	}

	PG struct {
		PoolMax int    `env-required:"true" yaml:"pool_max" env:"PG_POOL_MAX"`
		URL     string `env-required:"true" env:"PG_URL"`
	}

	Auth struct {
		JWTSecretKey string        `env-required:"true" env:"AUTH_JWT_SECRET_KEY"`
		TokenTTL     time.Duration `env_required:"true" yaml:"token_ttl" env:"AUTH_TOKEN_TTL"`
	}
)

func New(configPath string) (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig(configPath, cfg)
	if err != nil {
		return nil, fmt.Errorf("config - NewConfig - cleanenv.ReadConfig: %w", err)
	}

	err = cleanenv.UpdateEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("config - NewConfig - cleanenv.UpdateEnv: %w", err)
	}

	return cfg, nil
}
