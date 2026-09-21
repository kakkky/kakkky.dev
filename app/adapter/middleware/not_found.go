package middleware

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/errors"
)

func NotFound(next http.Handler) http.Handler {
	mux, ok := next.(*http.ServeMux)
	if !ok {
		panic("middleware.NotFound: next handler must be *http.ServeMux (place NotFound as the innermost wrap around mux)")
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, pattern := mux.Handler(r); pattern == "" {
			errors.Set(r.Context(), domain.ErrNotFound)
			return
		}
		mux.ServeHTTP(w, r)
	})
}
