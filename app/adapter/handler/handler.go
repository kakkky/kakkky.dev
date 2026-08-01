package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/domain"
)

type Handler struct {
	repo domain.Repository
}

func NewHandler(repo domain.Repository) *Handler {
	return &Handler{
		repo: repo,
	}
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
