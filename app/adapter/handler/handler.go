package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/usecase"
)

type Handler struct {
	usecase *usecase.UseCase
}

func NewHandler(usecase *usecase.UseCase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

type Route struct {
	Pattern string
	Handler http.Handler
}

func (h *Handler) PublicRoutes() []Route {
	return []Route{
		{
			Pattern: "GET /feed",
			Handler: NewFeedHandler(h.usecase.NewGetFeedUsecase()),
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

func (h *Handler) StaticRoutes() []Route {
	return []Route{
		{
			Pattern: "GET /assets/",
			Handler: http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets/dist"))),
		},
	}
}
