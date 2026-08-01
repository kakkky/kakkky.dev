package handler

import "net/http"

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

type Route struct {
	Pattern string
	Handler http.Handler
}

func (h *Handler) PublicRoutes() []Route {
	return []Route{
		{
			Pattern: "GET /hello",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("Hello, World!"))
			}),
		},
	}
}

func (h *Handler) AdminRoutes() []Route {
	return []Route{
		{
			Pattern: "GET /",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("Hello, World!"))
			}),
		},
	}
}
