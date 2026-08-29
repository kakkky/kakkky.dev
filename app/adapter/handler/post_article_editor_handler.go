package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/adapter/view/partials"
	"github.com/kakkky/kakkky.dev/domain"
)

// PostArticleEditorHandler は preview から editor に戻すための turbo-frame
// fragment を返す。body 保持のため form の body を echo する。DB access なし。
type PostArticleEditorHandler struct{}

func NewPostArticleEditorHandler() *PostArticleEditorHandler {
	return &PostArticleEditorHandler{}
}

func (h *PostArticleEditorHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		RenderError(rw, r, domain.ErrInvalidArgument.With("フォームの解析に失敗しました"))
		return
	}

	body := r.PostFormValue("body")
	_ = partials.ArticleEditorAreaEditor(body).Render(ctx, rw)
}
