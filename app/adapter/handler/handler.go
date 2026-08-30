package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/adapter/client"
	"github.com/kakkky/kakkky.dev/usecase"
)

type Handler struct {
	usecase    *usecase.UseCase
	ogpFetcher *client.OGPFetcher // GetLinkPreviewHandler 専用
}

func NewHandler(usecase *usecase.UseCase, ogpFetcher *client.OGPFetcher) *Handler {
	return &Handler{
		usecase:    usecase,
		ogpFetcher: ogpFetcher,
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
			Handler: NewGetFeedHandler(h.usecase.NewGetFeedUsecase()),
		},
		{
			Pattern: "GET /articles/{slug}",
			Handler: NewGetArticleHandler(h.usecase.NewGetArticleUsecase()),
		},
		{
			Pattern: "GET /link-preview",
			Handler: NewGetLinkPreviewHandler(h.ogpFetcher),
		},
	}
}

func (h *Handler) AdminRoutes() []Route {
	return []Route{
		{
			Pattern: "GET /dashboard",
			Handler: NewGetDashboardHandler(
				h.usecase.NewListArticlesUsecase(),
				h.usecase.NewListSeriesUsecase(),
			),
		},
		{
			Pattern: "GET /articles/new",
			Handler: NewGetNewArticleHandler(
				h.usecase.NewListTagsUsecase(),
			),
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
