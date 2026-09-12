package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/adapter/view/components"
	"github.com/kakkky/kakkky.dev/adapter/view/pages"
	"github.com/kakkky/kakkky.dev/adapter/view/partials"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/usecase"
)

type GetEditSeriesHandler struct {
	getSeriesForAdminUsecase *usecase.GetSeriesForAdminUsecase
	listTagsUsecase          *usecase.ListTagsUsecase
}

func NewGetEditSeriesHandler(
	getSeriesForAdminUsecase *usecase.GetSeriesForAdminUsecase,
	listTagsUsecase *usecase.ListTagsUsecase,
) *GetEditSeriesHandler {
	return &GetEditSeriesHandler{
		getSeriesForAdminUsecase: getSeriesForAdminUsecase,
		listTagsUsecase:          listTagsUsecase,
	}
}

func (h *GetEditSeriesHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := domain.Slug(r.PathValue("slug"))

	seriesOut, err := h.getSeriesForAdminUsecase.Exec(ctx, usecase.GetSeriesForAdminUsecaseInput{Slug: slug})
	if err != nil {
		RenderError(rw, r, err)
		return
	}

	tagsOut, err := h.listTagsUsecase.Exec(ctx)
	if err != nil {
		RenderError(rw, r, err)
		return
	}

	selectedTags := make([]components.TagViewModel, 0, len(seriesOut.Series.TagIDs))
	for _, tid := range seriesOut.Series.TagIDs {
		if t, ok := seriesOut.Tags[tid]; ok {
			selectedTags = append(selectedTags, components.TagViewModel{
				ID:   string(t.ID),
				Slug: string(t.Slug),
				Name: t.Name,
			})
		}
	}

	existingTags := make([]components.TagViewModel, 0, len(tagsOut.Tags))
	for _, t := range tagsOut.Tags {
		existingTags = append(existingTags, components.TagViewModel{
			ID:   string(t.ID),
			Slug: string(t.Slug),
			Name: t.Name,
		})
	}

	current := seriesOut.Series.Status
	statuses := []pages.Status{
		{Name: string(domain.SeriesStatusDraft), IsSelected: current == domain.SeriesStatusDraft},
		{Name: string(domain.SeriesStatusPublishedOngoing), IsSelected: current == domain.SeriesStatusPublishedOngoing},
		{Name: string(domain.SeriesStatusPublishedCompleted), IsSelected: current == domain.SeriesStatusPublishedCompleted},
	}

	articles := make([]partials.SeriesArticleListItemViewModel, len(seriesOut.Articles))
	for i, sa := range seriesOut.Articles {
		articles[i] = partials.SeriesArticleListItemViewModel{
			SeriesSlug: string(seriesOut.Series.Slug),
			ArticleID:  string(sa.Article.ID),
			Slug:       string(sa.Article.Slug),
			Title:      sa.Article.Title,
		}
	}

	_ = pages.EditSeries(pages.EditSeriesViewModel{
		Slug:        string(seriesOut.Series.Slug),
		Title:       seriesOut.Series.Title,
		Description: seriesOut.Series.Description,
		Statuses:    statuses,
		TagInput: components.TagInputViewModel{
			ExistingTags: existingTags,
			SelectedTags: selectedTags,
			FormID:       "series-update-form",
		},
		Articles: articles,
	}).Render(ctx, rw)
}
