package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/adapter/view"
	"github.com/kakkky/kakkky.dev/adapter/view/partials"
	"github.com/kakkky/kakkky.dev/domain"
)

// PostArticlePreviewHandler は form 全体を受け取り、body を goldmark で render し、
// editor-area turbo-frame を preview 状態で返す。DB access なし。
type PostArticlePreviewHandler struct{}

func NewPostArticlePreviewHandler() *PostArticlePreviewHandler {
	return &PostArticlePreviewHandler{}
}

func (h *PostArticlePreviewHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		RenderError(rw, r, domain.ErrInvalidArgument.With("フォームの解析に失敗しました"))
		return
	}

	body := r.PostFormValue("body")
	html, _ := view.ParseMarkdownArticle(body)

	_ = partials.ArticleEditorAreaPreview(body, html).Render(ctx, rw)
}
