package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/usecase"
)

type CreateArticleHandler struct {
	createArticleUsecase *usecase.CreateArticleUsecase
	getArticleUsecase    *usecase.GetArticleUsecase
	listTagsUsecase      *usecase.ListTagsUsecase
}

func NewCreateArticleHandler(
	createArticleUsecase *usecase.CreateArticleUsecase,
	getArticleUsecase *usecase.GetArticleUsecase,
	listTagsUsecase *usecase.ListTagsUsecase,
) *CreateArticleHandler {
	return &CreateArticleHandler{
		createArticleUsecase: createArticleUsecase,
		getArticleUsecase:    getArticleUsecase,
		listTagsUsecase:      listTagsUsecase,
	}
}

func (h *CreateArticleHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		RenderError(rw, r, domain.ErrInvalidArgument.With("フォームの解析に失敗しました"))
		return
	}

	out, err := h.createArticleUsecase.Exec(ctx, usecase.CreateArticleUsecaseInput{
		Title:    r.PostFormValue("title"),
		Body:     r.PostFormValue("body"),
		Status:   r.PostFormValue("status"),
		TagNames: r.PostForm["tag"],
	})
	if err != nil {
		RenderError(rw, r, err)
		return
	}

	renderArticleSavedResponse(rw, r, h.getArticleUsecase, h.listTagsUsecase, out.Slug)
}
