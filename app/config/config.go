package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	DatabaseURL        string `envconfig:"DATABASE_URL" required:"true"`
	HTTPPort           string `envconfig:"HTTP_PORT" default:"8080"`
	PublicBaseURL      string `envconfig:"PUBLIC_BASE_URL" required:"true"`
	AdminBaseURL       string `envconfig:"ADMIN_BASE_URL" required:"true"`
	Env                string `envconfig:"ENV" default:"DEVELOPMENT"`
	CFAccessTeamDomain string `envconfig:"CF_ACCESS_TEAM_DOMAIN"`
	CFAccessAudience   string `envconfig:"CF_ACCESS_AUD"`
	AdminEmail         string `envconfig:"ADMIN_EMAIL"`

	S3Endpoint        string `envconfig:"S3_ENDPOINT" required:"true"`
	S3Region          string `envconfig:"S3_REGION" default:"auto"`
	S3Bucket          string `envconfig:"S3_BUCKET" required:"true"`
	S3AccessKeyID     string `envconfig:"S3_ACCESS_KEY_ID" required:"true"`
	S3SecretAccessKey string `envconfig:"S3_SECRET_ACCESS_KEY" required:"true"`
	S3UsePathStyle    bool   `envconfig:"S3_USE_PATH_STYLE" default:"false"`
	ImageBaseURL      string `envconfig:"IMAGE_BASE_URL" required:"true"`

	SentryDSN string `envconfig:"SENTRY_DSN"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
