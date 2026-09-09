package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/adapter/view"
	"github.com/kakkky/kakkky.dev/adapter/view/partials"
	"github.com/kakkky/kakkky.dev/domain"
)

type PostArticleEditorPreviewHandler struct{}

func NewPostArticleEditorPreviewHandler() *PostArticleEditorPreviewHandler {
	return &PostArticleEditorPreviewHandler{}
}

func (h *PostArticleEditorPreviewHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		RenderError(rw, r, domain.ErrInvalidArgument.Wrap(err, "フォーム を 解釈 できません"))
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
