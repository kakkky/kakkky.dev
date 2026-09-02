package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/adapter/view/pages"
	"github.com/kakkky/kakkky.dev/usecase"
)

type GetNewArticleHandler struct {
	listTagsUsecase *usecase.ListTagsUsecase
}

func NewGetNewArticleHandler(listTagsUsecase *usecase.ListTagsUsecase) *GetNewArticleHandler {
	return &GetNewArticleHandler{
		listTagsUsecase: listTagsUsecase,
	}
}

func (h *GetNewArticleHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	out, err := h.listTagsUsecase.Exec(ctx)
	if err != nil {
		RenderError(rw, r, err)
		return
	}

	_ = pages.NewArticle(pages.NewArticleViewModel{
		ExistingTags: toTagViewModels(out.Tags),
	}).Render(ctx, rw)
}
