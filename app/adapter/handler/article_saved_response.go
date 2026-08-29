package handler

import (
	"net/http"

	"github.com/kakkky/hotwire-go/turbo"

	"github.com/kakkky/kakkky.dev/adapter/view/partials"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/usecase"
)

// renderArticleSavedResponse は Save/Update 成功時の応答を書く。
//   - Turbo Stream request なら turbo-stream fragment
//     (flash 追加 + form 全体を edit 版に replace + URL bar 更新)
//   - それ以外 (curl 等) は edit page へ 303 redirect fallback
func renderArticleSavedResponse(
	rw http.ResponseWriter,
	r *http.Request,
	getArticleUsecase *usecase.GetArticleUsecase,
	listTagsUsecase *usecase.ListTagsUsecase,
	slug domain.Slug,
) {
	editURL := "/admin/articles/" + string(slug) + "/edit"

	if !turbo.IsStreamRequest(r) {
		http.Redirect(rw, r, editURL, http.StatusSeeOther)
		return
	}

	form, err := editArticleFormVM(r.Context(), listTagsUsecase, getArticleUsecase, slug)
	if err != nil {
		RenderError(rw, r, err)
		return
	}

	turbo.StreamHeader(rw)
	rw.WriteHeader(http.StatusOK)
	_ = partials.ArticleFormSavedStream(form, editURL).Render(r.Context(), rw)
}
