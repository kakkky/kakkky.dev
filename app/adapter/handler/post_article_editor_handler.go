package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/adapter/view"
	"github.com/kakkky/kakkky.dev/adapter/view/partials"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/errors"
)

type PostArticleEditorHandler struct{}

func NewPostArticleEditorHandler() *PostArticleEditorHandler {
	return &PostArticleEditorHandler{}
}

func (h *PostArticleEditorHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		errors.Set(r.Context(), domain.ErrInvalidArgument.Wrap(err, "フォーム を 解釈 できません"))
		return
	}

	body := r.PostFormValue("body")

	switch r.URL.Query().Get("state") {
	case "preview":
		html, _ := view.ParseMarkdownArticle(body)
		_ = partials.ArticleEditorPreview(body, html).Render(ctx, rw)
	default:
		_ = partials.ArticleEditor(body).Render(ctx, rw)
	}
}
