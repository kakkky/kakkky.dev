package httpserver

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/kakkky/kakkky.dev/adapter/handler"
	"github.com/kakkky/kakkky.dev/adapter/middleware"
	"github.com/kakkky/kakkky.dev/config"
)

func NewMux(cfg *config.Config, h *handler.Handler, mw *middleware.Middleware) http.Handler {
	adminURL, err := url.Parse(cfg.AdminBaseURL)
	if err != nil {
		panic(fmt.Sprintf("invalid admin base URL: %v", err))
	}

	root := http.NewServeMux()
	for _, route := range h.StaticRoutes() {
		root.Handle(route.Pattern, route.Handler)
	}
	root.Handle("/", newPublicHandler(h, mw))
	root.Handle(fmt.Sprintf("%s/", adminURL.Hostname()), newAdminHandler(h, mw))

	return root
}

func newPublicHandler(h *handler.Handler, mw *middleware.Middleware) http.Handler {
	mux := http.NewServeMux()
	for _, route := range h.PublicRoutes() {
		mux.Handle(route.Pattern, route.Handler)
	}

	var handler http.Handler = mux
	wraps := mw.PublicWraps()
	for i := len(wraps) - 1; i >= 0; i-- {
		handler = wraps[i](handler)
	}
	return handler
}

func newAdminHandler(h *handler.Handler, mw *middleware.Middleware) http.Handler {
	mux := http.NewServeMux()
	for _, route := range h.AdminRoutes() {
		mux.Handle(route.Pattern, route.Handler)
	}

	var handler http.Handler = mux
	wraps := mw.AdminWraps()
	for i := len(wraps) - 1; i >= 0; i-- {
		handler = wraps[i](handler)
	}
	return handler
}
