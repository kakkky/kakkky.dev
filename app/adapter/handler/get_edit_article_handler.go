package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/adapter/view/pages"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/usecase"
)

type GetEditArticleHandler struct {
	getArticleUsecase *usecase.GetArticleUsecase
	listTagsUsecase   *usecase.ListTagsUsecase
}

func NewGetEditArticleHandler(
	getArticleUsecase *usecase.GetArticleUsecase,
	listTagsUsecase *usecase.ListTagsUsecase,
) *GetEditArticleHandler {
	return &GetEditArticleHandler{
		getArticleUsecase: getArticleUsecase,
		listTagsUsecase:   listTagsUsecase,
	}
}

func (h *GetEditArticleHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := domain.Slug(r.PathValue("slug"))

	form, err := editArticleFormVM(ctx, h.listTagsUsecase, h.getArticleUsecase, slug)
	if err != nil {
		RenderError(rw, r, err)
		return
	}

	_ = pages.EditArticle(pages.EditArticleViewModel{Form: form}).Render(ctx, rw)
}
