package handler

import (
	"context"

	"github.com/kakkky/kakkky.dev/adapter/view/partials"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/usecase"
)

// editArticleFormVM は edit 状態の ArticleFormViewModel を組み立てる helper。
// GetEditArticleHandler の初期表示と、Create/Update 成功時の in-place 差し替えで
// 同じ形の form を返すため共通化する。
func editArticleFormVM(
	ctx context.Context,
	listTagsUsecase *usecase.ListTagsUsecase,
	getArticleUsecase *usecase.GetArticleUsecase,
	slug domain.Slug,
) (partials.ArticleFormViewModel, error) {
	articleOut, err := getArticleUsecase.Exec(ctx, usecase.GetArticleUsecaseInput{Slug: slug})
	if err != nil {
		return partials.ArticleFormViewModel{}, err
	}
	tagsOut, err := listTagsUsecase.Exec(ctx)
	if err != nil {
		return partials.ArticleFormViewModel{}, err
	}

	selectedNames := make([]string, 0, len(articleOut.Article.TagIDs))
	for _, tid := range articleOut.Article.TagIDs {
		if t, ok := articleOut.Tags[tid]; ok {
			selectedNames = append(selectedNames, t.Name)
		}
	}

	return partials.ArticleFormViewModel{
		Action:       "/admin/articles/" + string(articleOut.Article.Slug) + "/update",
		SubmitLabel:  "更新",
		Slug:         string(articleOut.Article.Slug),
		Title:        articleOut.Article.Title,
		Body:         articleOut.Article.Body,
		Status:       string(articleOut.Article.Status),
		SelectedTags: selectedNames,
		AllTagNames:  tagNames(tagsOut.Tags),
		CreatedAt:    articleOut.Article.CreatedAt,
		UpdatedAt:    articleOut.Article.UpdatedAt,
	}, nil
}
