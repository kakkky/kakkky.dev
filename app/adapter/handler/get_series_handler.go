package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/adapter/view/pages"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/usecase"
)

type GetSeriesHandler struct {
	getSeriesUsecase *usecase.GetSeriesUsecase
}

func NewGetSeriesHandler(getSeriesUsecase *usecase.GetSeriesUsecase) *GetSeriesHandler {
	return &GetSeriesHandler{getSeriesUsecase: getSeriesUsecase}
}

func (h *GetSeriesHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := domain.Slug(r.PathValue("slug"))

	out, err := h.getSeriesUsecase.Exec(ctx, usecase.GetSeriesUsecaseInput{Slug: slug})
	if err != nil {
		RenderError(rw, r, err)
		return
	}

	seriesTagNames := make([]string, 0, len(out.Series.TagIDs))
	for _, tid := range out.Series.TagIDs {
		if t, ok := out.Tags[tid]; ok {
			seriesTagNames = append(seriesTagNames, t.Name)
		}
	}

	articleVMs := make([]pages.SeriesArticleViewModel, 0, len(out.Articles))
	for _, sa := range out.Articles {
		tagNames := make([]string, 0, len(sa.Article.TagIDs))
		for _, tid := range sa.Article.TagIDs {
			if t, ok := out.Tags[tid]; ok {
				tagNames = append(tagNames, t.Name)
			}
		}
		articleVMs = append(articleVMs, pages.SeriesArticleViewModel{
			Position:    sa.Position,
			Title:       sa.Article.Title,
			Href:        "/articles/" + string(sa.Article.Slug),
			PublishedAt: sa.Article.PublishedAt,
			Tags:        tagNames,
		})
	}

	vm := pages.SeriesViewModel{
		Title:       out.Series.Title,
		Description: out.Series.Description,
		Status:      string(out.Series.Status),
		PublishedAt: out.Series.PublishedAt,
		Tags:        seriesTagNames,
		Articles:    articleVMs,
	}
	_ = pages.Series(vm).Render(ctx, rw)
}
