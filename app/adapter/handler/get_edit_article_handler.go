package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/adapter/view/components"
	"github.com/kakkky/kakkky.dev/adapter/view/pages"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/usecase"
)

type GetEditArticleHandler struct {
	getArticleUsecase *usecase.GetArticleUsecase
	listTagsUsecase   *usecase.ListTagsUsecase
}

func NewGetEditArticleHandler(
	getArticleUsecase *usecase.GetArticleUsecase,
	listTagsUsecase *usecase.ListTagsUsecase,
) *GetEditArticleHandler {
	return &GetEditArticleHandler{
		getArticleUsecase: getArticleUsecase,
		listTagsUsecase:   listTagsUsecase,
	}
}

func (h *GetEditArticleHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := domain.Slug(r.PathValue("slug"))

	articleOut, err := h.getArticleUsecase.Exec(ctx, usecase.GetArticleUsecaseInput{Slug: slug})
	if err != nil {
		RenderError(rw, r, err)
		return
	}

	tagsOut, err := h.listTagsUsecase.Exec(ctx)
	if err != nil {
		RenderError(rw, r, err)
		return
	}

	selectedTags := make([]components.TagViewModel, 0, len(articleOut.Article.TagIDs))
	for _, tid := range articleOut.Article.TagIDs {
		if t, ok := articleOut.Tags[tid]; ok {
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

	current := articleOut.Article.Status
	statuses := []pages.Status{
		{Name: string(domain.ArticleStatusDraft), IsSelected: current == domain.ArticleStatusDraft},
		{Name: string(domain.ArticleStatusPublished), IsSelected: current == domain.ArticleStatusPublished},
	}

	_ = pages.EditArticle(pages.EditArticleViewModel{
		Slug:     string(articleOut.Article.Slug),
		Title:    articleOut.Article.Title,
		Body:     articleOut.Article.Body,
		Statuses: statuses,
		TagInput: components.TagInputViewModel{
			ExistingTags: existingTags,
			SelectedTags: selectedTags,
		},
		CreatedAt: articleOut.Article.CreatedAt,
		UpdatedAt: articleOut.Article.UpdatedAt,
	}).Render(ctx, rw)
}
