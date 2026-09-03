package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/adapter/view/pages"
)

type GetEditArticleHandler struct{}

func NewGetEditArticleHandler() *GetEditArticleHandler {
	return &GetEditArticleHandler{}
}

func (h *GetEditArticleHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_ = pages.EditArticle(pages.EditArticleViewModel{}).Render(ctx, rw)
}
