package handler

import (
	"net/http"

	"github.com/kakkky/hotwire-go/turbo"

	"github.com/kakkky/kakkky.dev/adapter/view/partials"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/usecase"
)

type PostSeriesBySlugHandler struct {
	updateSeriesUsecase *usecase.UpdateSeriesUsecase
}

func NewPostSeriesBySlugHandler(updateSeriesUsecase *usecase.UpdateSeriesUsecase) *PostSeriesBySlugHandler {
	return &PostSeriesBySlugHandler{
		updateSeriesUsecase: updateSeriesUsecase,
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

	existingTagIDs := make([]domain.TagID, 0, len(tagIDStrs))
	for _, s := range tagIDStrs {
		if s == "" {
			continue
		}
		existingTagIDs = append(existingTagIDs, domain.TagID(s))
	}

	articleIDs := make([]domain.ArticleID, 0, len(articleIDStrs))
	for _, s := range articleIDStrs {
		if s == "" {
			continue
		}
		articleIDs = append(articleIDs, domain.ArticleID(s))
	}

	if _, err := h.updateSeriesUsecase.Exec(ctx, usecase.UpdateSeriesUsecaseInput{
		Slug:           slug,
		Title:          title,
		Description:    description,
		Status:         status,
		ExistingTagIDs: existingTagIDs,
		NewTagNames:    newTagNames,
		ArticleIDs:     articleIDs,
	}); err != nil {
		RenderError(rw, r, err)
		return
	}

	turbo.StreamHeader(rw)
	_ = partials.Flash(partials.FlashViewModel{
		Kind: partials.FlashKindInfo,
		Msg:  "更新しました",
	}).Render(ctx, rw)
}
