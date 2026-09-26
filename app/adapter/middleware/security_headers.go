package middleware

import (
	"net/http"
	"strings"

	"github.com/kakkky/kakkky.dev/config"
)

// contentSecurityPolicy は defense-in-depth の CSP。実運用中の外部ホストを許可する
// ホワイトリスト方式で、まずは 'unsafe-inline' を許容し、nonce/hash 化は段階導入で進める。
var contentSecurityPolicy = strings.Join([]string{
	"default-src 'self'",
	"script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://esm.sh https://www.googletagmanager.com",
	"style-src 'self' 'unsafe-inline'",
	"img-src 'self' data: https:",
	"font-src 'self' data:",
	"connect-src 'self' https://www.google-analytics.com https://analytics.google.com https://www.googletagmanager.com",
	"frame-ancestors 'none'",
	"base-uri 'self'",
	"form-action 'self'",
	"object-src 'none'",
}, "; ")

func SecurityHeaders(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("Content-Security-Policy", contentSecurityPolicy)
			h.Set("X-Frame-Options", "DENY")
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
			if cfg.Env != "DEVELOPMENT" {
				// HSTS はローカル HTTP で有効化すると次回以降強制 HTTPS になり dev が壊れるため本番のみ。
				h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	}
}
