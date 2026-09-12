package handler

import (
	"fmt"
	"net/http"

	"github.com/kakkky/hotwire-go/turbo"

	"github.com/kakkky/kakkky.dev/adapter/view/partials"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/usecase"
)

type PostDeleteSeriesArticleHandler struct {
	deleteArticleUsecase *usecase.DeleteArticleUsecase
}

func NewPostDeleteSeriesArticleHandler(uc *usecase.DeleteArticleUsecase) *PostDeleteSeriesArticleHandler {
	return &PostDeleteSeriesArticleHandler{deleteArticleUsecase: uc}
}

func (h *PostDeleteSeriesArticleHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	articleID := domain.ArticleID(r.PathValue("article_id"))

	out, err := h.deleteArticleUsecase.Exec(ctx, usecase.DeleteArticleUsecaseInput{ID: articleID})
	if err != nil {
		RenderError(rw, r, err)
		return
	}

	turbo.StreamHeader(rw)
	_ = turbo.StreamRemove("series-article-" + string(articleID)).Render(ctx, rw)
	_ = partials.Flash(partials.FlashViewModel{
		Kind: partials.FlashKindInfo,
		Msg:  fmt.Sprintf("記事「%s」を 削除 しました", out.Title),
	}).Render(ctx, rw)
}
