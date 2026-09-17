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

func (m *Middleware) GlobalWraps() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		ContentTypeHTML,
	}
}

func (m *Middleware) AdminWraps() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		CloudflareAccess(m.cfg),
	}
}

func (m *Middleware) MuxWraps(mux *http.ServeMux) http.Handler {
	return NotFound(mux)
}
