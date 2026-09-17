package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	DatabaseURL   string `envconfig:"DATABASE_URL" required:"true"`
	HTTPPort      string `envconfig:"HTTP_PORT" default:"8080"`
	PublicBaseURL string `envconfig:"PUBLIC_BASE_URL" required:"true"`
	AdminBaseURL  string `envconfig:"ADMIN_BASE_URL" required:"true"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
