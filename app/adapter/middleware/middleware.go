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
		RequestScope,
		AccessLog,
		SecurityHeaders(m.cfg),
		ContentTypeHTML,
		ErrorHandler,
		NotFound,
	}
}

func (m *Middleware) AdminWraps() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		RequestScope,
		AccessLog,
		SecurityHeaders(m.cfg),
		ContentTypeHTML,
		ErrorHandler,
		CloudflareAccess(m.cfg),
		NotFound,
	}
}
