package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/adapter/view/partials"
)

type GetCancelNewSeriesArticleHandler struct{}

func NewGetCancelNewSeriesArticleHandler() *GetCancelNewSeriesArticleHandler {
	return &GetCancelNewSeriesArticleHandler{}
}

func (h *GetCancelNewSeriesArticleHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	_ = partials.SeriesNewArticleButton(partials.SeriesNewArticleViewModel{
		SeriesSlug: slug,
	}).Render(r.Context(), rw)
}
