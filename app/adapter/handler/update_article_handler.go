package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/usecase"
)

type UpdateArticleHandler struct {
	updateArticleUsecase *usecase.UpdateArticleUsecase
	getArticleUsecase    *usecase.GetArticleUsecase
	listTagsUsecase      *usecase.ListTagsUsecase
}

func NewUpdateArticleHandler(
	updateArticleUsecase *usecase.UpdateArticleUsecase,
	getArticleUsecase *usecase.GetArticleUsecase,
	listTagsUsecase *usecase.ListTagsUsecase,
) *UpdateArticleHandler {
	return &UpdateArticleHandler{
		updateArticleUsecase: updateArticleUsecase,
		getArticleUsecase:    getArticleUsecase,
		listTagsUsecase:      listTagsUsecase,
	}
}

func (h *UpdateArticleHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		RenderError(rw, r, domain.ErrInvalidArgument.With("フォームの解析に失敗しました"))
		return
	}

	currentSlug := domain.Slug(r.PathValue("slug"))
	out, err := h.updateArticleUsecase.Exec(ctx, usecase.UpdateArticleUsecaseInput{
		CurrentSlug: currentSlug,
		Title:       r.PostFormValue("title"),
		Body:        r.PostFormValue("body"),
		Status:      r.PostFormValue("status"),
		TagNames:    r.PostForm["tag"],
	})
	if err != nil {
		RenderError(rw, r, err)
		return
	}

	renderArticleSavedResponse(rw, r, h.getArticleUsecase, h.listTagsUsecase, out.Slug)
}
