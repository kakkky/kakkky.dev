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
	root.Handle("/", newPublicHandler(h, mw))
	root.Handle(fmt.Sprintf("%s/", adminURL.Hostname()), newAdminHandler(h, mw))

	return root
}

func newPublicHandler(h *handler.Handler, mw *middleware.Middleware) http.Handler {
	mux := http.NewServeMux()
	registerStaticRoutes(mux, h)
	registerPublicRoutes(mux, h, mw)
	return mw.MuxWraps(mux)
}

func newAdminHandler(h *handler.Handler, mw *middleware.Middleware) http.Handler {
	mux := http.NewServeMux()
	registerStaticRoutes(mux, h)
	registerAdminRoutes(mux, h, mw)
	return mw.MuxWraps(mux)
}

func registerStaticRoutes(mux *http.ServeMux, h *handler.Handler) {
	for _, route := range h.StaticRoutes() {
		mux.Handle(route.Pattern, route.Handler)
	}
}

func registerPublicRoutes(mux *http.ServeMux, h *handler.Handler, mw *middleware.Middleware) {
	globalWraps := mw.GlobalWraps()
	for _, route := range h.PublicRoutes() {
		handler := route.Handler
		for i := len(globalWraps) - 1; i >= 0; i-- {
			handler = globalWraps[i](handler)
		}
		mux.Handle(route.Pattern, handler)
	}
}

func registerAdminRoutes(mux *http.ServeMux, h *handler.Handler, mw *middleware.Middleware) {
	adminWraps := mw.AdminWraps()
	globalWraps := mw.GlobalWraps()

	for _, route := range h.AdminRoutes() {
		handler := route.Handler
		for i := len(adminWraps) - 1; i >= 0; i-- {
			handler = adminWraps[i](handler)
		}
		for i := len(globalWraps) - 1; i >= 0; i-- {
			handler = globalWraps[i](handler)
		}
		mux.Handle(route.Pattern, handler)
	}
}
