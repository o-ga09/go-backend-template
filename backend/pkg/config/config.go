package config

import (
	"context"

	"github.com/caarlos0/env"
)

type CfgKey string

const ConfigKey CfgKey = "config"

type Config struct {
	Env          string `env:"ENV" envDefault:"dev"`
	Port         string `env:"PORT" envDefault:"80"`
	Database_url string `env:"DATABASE_URL" envDefult:""`
	ProjectID    string `env:"PROJECTID" envDefault:""`
}

func New(ctx context.Context) (context.Context, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return context.WithValue(ctx, ConfigKey, cfg), nil
}
