package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"

	"github.com/kakkky/kakkky.dev/config"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/errors"
)

func CloudflareAccess(cfg *config.Config) func(http.Handler) http.Handler {
	if cfg.Env == "DEVELOPMENT" {
		return func(next http.Handler) http.Handler { return next }
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	provider, err := oidc.NewProvider(ctx, cfg.CFAccessTeamDomain)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize cloudflare access provider: %v", err))
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.CFAccessAudience})
	email := cfg.AdminEmail

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Cf-Access-Jwt-Assertion")
			if token == "" {
				errors.Set(r.Context(), domain.ErrNotFound)
				return
			}
			idToken, err := verifier.Verify(r.Context(), token)
			if err != nil {
				errors.Set(r.Context(), domain.ErrNotFound)
				return
			}
			var claims struct {
				Email string `json:"email"`
			}
			if err := idToken.Claims(&claims); err != nil || claims.Email != email {
				errors.Set(r.Context(), domain.ErrNotFound)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
