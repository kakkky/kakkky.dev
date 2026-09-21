package middleware

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/adapter/handler"
	"github.com/kakkky/kakkky.dev/domain"
)

func NotFound(next http.Handler) http.Handler {
	mux, ok := next.(*http.ServeMux)
	if !ok {
		panic("middleware.NotFound: next handler must be *http.ServeMux (place NotFound as the innermost wrap around mux)")
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, pattern := mux.Handler(r); pattern == "" {
			handler.RenderError(w, r, domain.ErrNotFound)
			return
		}
		mux.ServeHTTP(w, r)
	})
}
