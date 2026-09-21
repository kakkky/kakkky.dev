package middleware

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/errors"
	"github.com/kakkky/kakkky.dev/sentry"
)

func RequestScope(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = errors.NewContext(ctx)
		ctx = sentry.NewContext(ctx, r)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}
