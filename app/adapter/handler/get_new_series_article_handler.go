package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/adapter/view/partials"
)

type GetNewSeriesArticleHandler struct{}

func NewGetNewSeriesArticleHandler() *GetNewSeriesArticleHandler {
	return &GetNewSeriesArticleHandler{}
}

func (h *GetNewSeriesArticleHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	_ = partials.SeriesNewArticleForm(partials.SeriesNewArticleViewModel{
		SeriesSlug: slug,
	}).Render(r.Context(), rw)
}
