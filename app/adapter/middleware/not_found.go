package middleware

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/adapter/handler"
	"github.com/kakkky/kakkky.dev/domain"
)

// NotFound wraps the given mux so that requests not matching any registered
// pattern are handled by handler.RenderError with domain.ErrNotFound.
func NotFound(mux *http.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, pattern := mux.Handler(r); pattern == "" {
			handler.RenderError(w, r, domain.ErrNotFound)
			return
		}
		mux.ServeHTTP(w, r)
	})
}
