package usecase

import (
	"context"
	"time"

	"github.com/kakkky/kakkky.dev/domain"
)

type ListArticlesUsecase struct {
	articleRepo domain.ArticleRepository
}

func (us *UseCase) NewListArticlesUsecase() *ListArticlesUsecase {
	return &ListArticlesUsecase{
		articleRepo: us.repo.NewArticleRepository(),
	}
}

type ListArticlesUsecaseCursor struct {
	AfterID        domain.ArticleID
	AfterCreatedAt time.Time
}

type ListArticlesUsecaseInput struct {
	Cursor ListArticlesUsecaseCursor
	Limit  int
}

type ListArticlesUsecaseOutput struct {
	Articles   []domain.Article
	NextCursor ListArticlesUsecaseCursor
}

func (us *ListArticlesUsecase) Exec(ctx context.Context, in ListArticlesUsecaseInput) (ListArticlesUsecaseOutput, error) {
	fetchLimit := in.Limit
	if in.Limit > 0 {
		fetchLimit = in.Limit + 1
	}

	rows, err := us.articleRepo.List(ctx, in.Cursor.AfterID, in.Cursor.AfterCreatedAt, fetchLimit)
	if err != nil {
		return ListArticlesUsecaseOutput{}, err
	}

	articles := make([]domain.Article, len(rows))
	for i, a := range rows {
		articles[i] = *a
	}

	var next ListArticlesUsecaseCursor
	if in.Limit > 0 && len(articles) > in.Limit {
		articles = articles[:in.Limit]
		last := articles[len(articles)-1]
		next = ListArticlesUsecaseCursor{
			AfterID:        last.ID,
			AfterCreatedAt: last.CreatedAt,
		}
	}
	return ListArticlesUsecaseOutput{Articles: articles, NextCursor: next}, nil
}
