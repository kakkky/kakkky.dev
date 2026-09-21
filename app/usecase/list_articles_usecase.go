package usecase

import (
	"context"
	"time"

	"github.com/kakkky/kakkky.dev/domain"
)

type ListArticlesUsecase struct {
	articleRepo domain.ArticleRepository
	seriesRepo  domain.SeriesRepository
}

func (us *UseCase) NewListArticlesUsecase() *ListArticlesUsecase {
	return &ListArticlesUsecase{
		articleRepo: us.repo.NewArticleRepository(),
		seriesRepo:  us.repo.NewSeriesRepository(),
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
	SeriesByID map[domain.ArticleID]*domain.Series
	NextCursor ListArticlesUsecaseCursor
	Total      int
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

	seriesByID := map[domain.ArticleID]*domain.Series{}
	if len(articles) > 0 {
		ids := make([]domain.ArticleID, len(articles))
		for i, a := range articles {
			ids[i] = a.ID
		}
		seriesList, err := us.seriesRepo.FindByArticleIDs(ctx, ids...)
		if err != nil {
			return ListArticlesUsecaseOutput{}, err
		}
		for _, s := range seriesList {
			for _, sa := range s.Articles {
				seriesByID[sa.ArticleID] = s
			}
		}
	}

	total, err := us.articleRepo.Count(ctx)
	if err != nil {
		return ListArticlesUsecaseOutput{}, err
	}
	return ListArticlesUsecaseOutput{Articles: articles, SeriesByID: seriesByID, NextCursor: next, Total: total}, nil
}
