package usecase

import (
	"context"

	"github.com/kakkky/kakkky.dev/domain"
)

type GetSeriesForAdminUsecase struct {
	seriesRepo  domain.SeriesRepository
	articleRepo domain.ArticleRepository
	tagRepo     domain.TagRepository
}

func (us *UseCase) NewGetSeriesForAdminUsecase() *GetSeriesForAdminUsecase {
	return &GetSeriesForAdminUsecase{
		seriesRepo:  us.repo.NewSeriesRepository(),
		articleRepo: us.repo.NewArticleRepository(),
		tagRepo:     us.repo.NewTagRepository(),
	}
}

type GetSeriesForAdminUsecaseInput struct {
	Slug domain.Slug
}

func (us *GetSeriesForAdminUsecase) Exec(ctx context.Context, in GetSeriesForAdminUsecaseInput) (GetSeriesUsecaseOutput, error) {
	if err := in.validate(); err != nil {
		return GetSeriesUsecaseOutput{}, err
	}
	return getSeries(ctx, us.seriesRepo, us.articleRepo, us.tagRepo, in.Slug, true)
}

func (in GetSeriesForAdminUsecaseInput) validate() error {
	if in.Slug == "" {
		return domain.ErrInvalidArgument.With("slug は 必須 です")
	}
	return nil
}
