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
			Pattern: "GET /series/{slug}",
			Handler: NewGetSeriesHandler(h.usecase.NewGetSeriesUsecase()),
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
		{
			Pattern: "POST /articles",
			Handler: NewPostArticlesHandler(
				h.usecase.NewCreateArticleUsecase(),
			),
		},
		{
			Pattern: "GET /articles/{slug}/edit",
			Handler: NewGetEditArticleHandler(
				h.usecase.NewGetArticleForAdminUsecase(),
				h.usecase.NewListTagsUsecase(),
			),
		},
		{
			Pattern: "POST /articles/{slug}",
			Handler: NewPostArticleHandler(
				h.usecase.NewUpdateArticleUsecase(),
			),
		},
		{
			Pattern: "POST /article-editor",
			Handler: NewPostArticleEditorHandler(),
		},
		{
			Pattern: "GET /series/new",
			Handler: NewGetNewSeriesHandler(
				h.usecase.NewListTagsUsecase(),
			),
		},
		{
			Pattern: "POST /series",
			Handler: NewPostSeriesHandler(
				h.usecase.NewCreateSeriesUsecase(),
			),
		},
		{
			Pattern: "GET /series/{slug}/edit",
			Handler: NewGetEditSeriesHandler(
				h.usecase.NewGetSeriesForAdminUsecase(),
				h.usecase.NewListTagsUsecase(),
			),
		},
		{
			Pattern: "POST /series/{slug}",
			Handler: NewPostSeriesBySlugHandler(
				h.usecase.NewUpdateSeriesUsecase(),
				h.usecase.NewGetSeriesForAdminUsecase(),
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
