package usecase

import (
	"context"

	"github.com/kakkky/kakkky.dev/domain"
)

type GetArticleForAdminUsecase struct {
	articleRepo domain.ArticleRepository
	seriesRepo  domain.SeriesRepository
	tagRepo     domain.TagRepository
}

func (us *UseCase) NewGetArticleForAdminUsecase() *GetArticleForAdminUsecase {
	return &GetArticleForAdminUsecase{
		articleRepo: us.repo.NewArticleRepository(),
		seriesRepo:  us.repo.NewSeriesRepository(),
		tagRepo:     us.repo.NewTagRepository(),
	}
}

type GetArticleForAdminUsecaseInput struct {
	Slug domain.Slug
}

func (us *GetArticleForAdminUsecase) Exec(ctx context.Context, input GetArticleForAdminUsecaseInput) (GetArticleUsecaseOutput, error) {
	if err := input.validate(); err != nil {
		return GetArticleUsecaseOutput{}, err
	}
	return getArticle(ctx, us.articleRepo, us.tagRepo, us.seriesRepo, input.Slug, true)
}

func (in GetArticleForAdminUsecaseInput) validate() error {
	if in.Slug == "" {
		return domain.ErrInvalidArgument.With("slug は 必須 です")
	}
	return nil
}
