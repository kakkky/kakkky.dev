package httpserver

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/adapter/handler"
	"github.com/kakkky/kakkky.dev/adapter/middleware"
)

func NewMux(h *handler.Handler, mw *middleware.Middleware) http.Handler {
	mux := http.NewServeMux()
	registerAdminRoutes(mux, h, mw)
	registerPublicRoutes(mux, h, mw)
	registerStaticRoutes(mux, h)
	return mw.MuxWraps(mux)
}

func registerStaticRoutes(mux *http.ServeMux, h *handler.Handler) {
	for _, route := range h.StaticRoutes() {
		mux.Handle(route.Pattern, route.Handler)
	}
}

func registerPublicRoutes(mux *http.ServeMux, h *handler.Handler, mw *middleware.Middleware) {
	routes := h.PublicRoutes()
	for _, route := range routes {
		handler := route.Handler
		globalWraps := mw.GlobalWraps()
		for i := len(globalWraps) - 1; i >= 0; i-- {
			handler = globalWraps[i](handler)
		}
		mux.Handle(route.Pattern, handler)
	}
}

func registerAdminRoutes(mux *http.ServeMux, h *handler.Handler, mw *middleware.Middleware) {
	adminMux := http.NewServeMux()
	routes := h.AdminRoutes()

	adminWraps := mw.AdminWraps()
	globalWraps := mw.GlobalWraps()

	for _, route := range routes {
		handler := route.Handler
		for i := len(adminWraps) - 1; i >= 0; i-- {
			handler = adminWraps[i](handler)
		}
		for i := len(globalWraps) - 1; i >= 0; i-- {
			handler = globalWraps[i](handler)
		}
		adminMux.Handle(route.Pattern, handler)
	}

	mux.Handle("/admin/", http.StripPrefix("/admin", adminMux))
}
