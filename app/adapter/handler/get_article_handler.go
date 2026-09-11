package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/adapter/view"
	"github.com/kakkky/kakkky.dev/adapter/view/pages"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/usecase"
)

type GetArticleHandler struct {
	getArticleUsecase *usecase.GetArticleUsecase
}

func NewGetArticleHandler(getArticleUsecase *usecase.GetArticleUsecase) *GetArticleHandler {
	return &GetArticleHandler{
		getArticleUsecase: getArticleUsecase,
	}
}

func (h *GetArticleHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := domain.Slug(r.PathValue("slug"))

	out, err := h.getArticleUsecase.Exec(ctx, usecase.GetArticleUsecaseInput{Slug: slug})
	if err != nil {
		RenderError(rw, r, err)
		return
	}

	tagNames := make([]string, 0, len(out.Article.TagIDs))
	for _, tid := range out.Article.TagIDs {
		if t, ok := out.Tags[tid]; ok {
			tagNames = append(tagNames, t.Name)
		}
	}

	htmlBody, outline := view.ParseMarkdownArticle(out.Article.Body)

	vm := pages.ArticleViewModel{
		Title:       out.Article.Title,
		PublishedAt: out.Article.PublishedAt,
		Tags:        tagNames,
		Body:        htmlBody,
		Outline:     outline,
		InSeriesRef: toArticleSeriesVM(out.InSeriesRef),
	}
	_ = pages.Article(vm).Render(ctx, rw)
}

func toArticleSeriesVM(s *usecase.GetArticleUsecaseSeriesRef) *pages.SeriesRefViewModel {
	if s == nil {
		return nil
	}
	return &pages.SeriesRefViewModel{
		Slug:                   string(s.Slug),
		Title:                  s.Title,
		Status:                 string(s.Status),
		CurrentArticlePosition: s.CurrentArticlePosition,
		PrevArticleRef:         toSeriesNeighborArticleRefVM(s.PrevArticleRef),
		NextArticleRef:         toSeriesNeighborArticleRefVM(s.NextArticleRef),
	}
}

func toSeriesNeighborArticleRefVM(n *usecase.GetArticleUsecaseSeriesNeighborArticleRef) *pages.SeriesNeighborArticleRefViewModel {
	if n == nil {
		return nil
	}
	return &pages.SeriesNeighborArticleRefViewModel{
		Slug:     string(n.Slug),
		Title:    n.Title,
		Position: n.Position,
	}
}
