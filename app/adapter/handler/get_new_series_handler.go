package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/adapter/view/pages"
	"github.com/kakkky/kakkky.dev/usecase"
)

type GetNewSeriesHandler struct {
	listTagsUsecase *usecase.ListTagsUsecase
}

func NewGetNewSeriesHandler(listTagsUsecase *usecase.ListTagsUsecase) *GetNewSeriesHandler {
	return &GetNewSeriesHandler{
		listTagsUsecase: listTagsUsecase,
	}
}

func (h *GetNewSeriesHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	out, err := h.listTagsUsecase.Exec(ctx)
	if err != nil {
		RenderError(rw, r, err)
		return
	}

	_ = pages.NewSeries(pages.NewSeriesViewModel{
		ExistingTags: toTagViewModels(out.Tags),
	}).Render(ctx, rw)
}
