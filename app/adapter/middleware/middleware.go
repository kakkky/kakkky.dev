package middleware

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/config"
)

type Middleware struct {
	cfg *config.Config
}

func NewMiddleware(cfg *config.Config) *Middleware {
	return &Middleware{
		cfg: cfg,
	}
}

func (m *Middleware) PublicWraps() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		AccessLog,
		ContentTypeHTML,
		NotFound,
	}
}

func (m *Middleware) AdminWraps() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		AccessLog,
		ContentTypeHTML,
		CloudflareAccess(m.cfg),
		NotFound,
	}
}
