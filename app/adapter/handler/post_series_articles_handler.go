package handler

import (
	"fmt"
	"net/http"

	"github.com/kakkky/hotwire-go/turbo"

	"github.com/kakkky/kakkky.dev/adapter/view/partials"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/usecase"
)

type PostSeriesArticlesHandler struct {
	createArticleInSeriesUsecase *usecase.CreateArticleInSeriesUsecase
}

func NewPostSeriesArticlesHandler(uc *usecase.CreateArticleInSeriesUsecase) *PostSeriesArticlesHandler {
	return &PostSeriesArticlesHandler{createArticleInSeriesUsecase: uc}
}

func (h *PostSeriesArticlesHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	seriesSlug := domain.Slug(r.PathValue("slug"))

	if err := r.ParseForm(); err != nil {
		RenderError(rw, r, domain.ErrInvalidArgument.Wrap(err, "フォーム を 解釈 できません"))
		return
	}
	title := r.FormValue("title")

	out, err := h.createArticleInSeriesUsecase.Exec(ctx, usecase.CreateArticleInSeriesUsecaseInput{
		SeriesSlug: seriesSlug,
		Title:      title,
	})
	if err != nil {
		RenderError(rw, r, err)
		return
	}

	turbo.StreamHeader(rw)
	_ = partials.SeriesArticleListItemStream(partials.SeriesArticleListItemViewModel{
		SeriesSlug: string(seriesSlug),
		ArticleID:  string(out.ArticleID),
		Slug:       string(out.ArticleSlug),
		Title:      out.ArticleTitle,
	}).Render(ctx, rw)
	_ = partials.SeriesNewArticleButtonStream(partials.SeriesNewArticleViewModel{
		SeriesSlug: string(seriesSlug),
	}).Render(ctx, rw)
	_ = partials.Flash(partials.FlashViewModel{
		Kind: partials.FlashKindInfo,
		Msg:  fmt.Sprintf("記事「%s」を 追加 しました", out.ArticleTitle),
	}).Render(ctx, rw)
}
