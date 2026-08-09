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
