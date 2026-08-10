package middleware

import "net/http"

type Middleware struct{}

func NewMiddleware() *Middleware {
	return &Middleware{}
}

func (m *Middleware) GlobalWraps() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		ContentTypeHTML,
	}
}

func (m *Middleware) AdminWraps() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{}
}

func (m *Middleware) MuxWraps(mux *http.ServeMux) http.Handler {
	return NotFound(mux)
}
