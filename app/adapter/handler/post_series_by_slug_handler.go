package handler

import (
	"net/http"

	"github.com/kakkky/hotwire-go/turbo"

	"github.com/kakkky/kakkky.dev/adapter/view/components"
	"github.com/kakkky/kakkky.dev/adapter/view/partials"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/usecase"
)

type PostSeriesBySlugHandler struct {
	updateSeriesUsecase      *usecase.UpdateSeriesUsecase
	getSeriesForAdminUsecase *usecase.GetSeriesForAdminUsecase
}

func NewPostSeriesBySlugHandler(
	updateSeriesUsecase *usecase.UpdateSeriesUsecase,
	getSeriesForAdminUsecase *usecase.GetSeriesForAdminUsecase,
) *PostSeriesBySlugHandler {
	return &PostSeriesBySlugHandler{
		updateSeriesUsecase:      updateSeriesUsecase,
		getSeriesForAdminUsecase: getSeriesForAdminUsecase,
	}
}

func (h *PostSeriesBySlugHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := domain.Slug(r.PathValue("slug"))

	if err := r.ParseForm(); err != nil {
		RenderError(rw, r, domain.ErrInvalidArgument.Wrap(err, "フォーム を 解釈 できません"))
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	status := domain.SeriesStatus(r.FormValue("status"))
	tagIDStrs := r.Form["tag_id"]
	newTagNames := r.Form["new_tag"]
	articleIDStrs := r.Form["article_id"]
	newArticleTitles := r.Form["new_article_title"]
	deleteArticleIDStrs := r.Form["delete_article_id"]

	existingTagIDs := make([]domain.TagID, 0, len(tagIDStrs))
	for _, s := range tagIDStrs {
		if s == "" {
			continue
		}
		existingTagIDs = append(existingTagIDs, domain.TagID(s))
	}

	orderedArticleIDs := make([]domain.ArticleID, 0, len(articleIDStrs))
	for _, s := range articleIDStrs {
		if s == "" {
			continue
		}
		orderedArticleIDs = append(orderedArticleIDs, domain.ArticleID(s))
	}

	deleteArticleIDs := make([]domain.ArticleID, 0, len(deleteArticleIDStrs))
	for _, s := range deleteArticleIDStrs {
		if s == "" {
			continue
		}
		deleteArticleIDs = append(deleteArticleIDs, domain.ArticleID(s))
	}

	if _, err := h.updateSeriesUsecase.Exec(ctx, usecase.UpdateSeriesUsecaseInput{
		Slug:              slug,
		Title:             title,
		Description:       description,
		Status:            status,
		ExistingTagIDs:    existingTagIDs,
		NewTagNames:       newTagNames,
		OrderedArticleIDs: orderedArticleIDs,
		NewArticleTitles:  newArticleTitles,
		DeleteArticleIDs:  deleteArticleIDs,
	}); err != nil {
		RenderError(rw, r, err)
		return
	}

	seriesOut, err := h.getSeriesForAdminUsecase.Exec(ctx, usecase.GetSeriesForAdminUsecaseInput{Slug: slug})
	if err != nil {
		RenderError(rw, r, err)
		return
	}

	items := make([]components.EditSeriesArticleItemViewModel, len(seriesOut.Articles))
	for i, sa := range seriesOut.Articles {
		items[i] = components.EditSeriesArticleItemViewModel{
			ArticleID: string(sa.Article.ID),
			Slug:      string(sa.Article.Slug),
			Title:     sa.Article.Title,
		}
	}

	turbo.StreamHeader(rw)
	_ = partials.EditSeriesArticlesListStreamUpdate(items).Render(ctx, rw)
	_ = partials.Flash(partials.FlashViewModel{
		Kind: partials.FlashKindInfo,
		Msg:  "更新しました",
	}).Render(ctx, rw)
}
