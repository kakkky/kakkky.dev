package usecase

import (
	"context"
	"errors"

	"github.com/kakkky/kakkky.dev/domain"
)

type GetArticleUsecase struct {
	articleRepo domain.ArticleRepository
	tagRepo     domain.TagRepository
}

func (us *UseCase) NewGetArticleUsecase() *GetArticleUsecase {
	return &GetArticleUsecase{
		articleRepo: us.repo.NewArticleRepository(),
		tagRepo:     us.repo.NewTagRepository(),
	}
}

type GetArticleUsecaseInput struct {
	Slug domain.Slug
}

type GetArticleUsecaseOutput struct {
	Article domain.Article
	Tags    map[domain.TagID]domain.Tag
}

func (us *GetArticleUsecase) Exec(ctx context.Context, input GetArticleUsecaseInput) (GetArticleUsecaseOutput, error) {
	article, err := us.articleRepo.FindBySlug(ctx, input.Slug)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrNotFound):
			return GetArticleUsecaseOutput{}, nil
		default:
			return GetArticleUsecaseOutput{}, err
		}
	}
	tags, err := us.tagRepo.FindByIDs(ctx, article.TagIDs...)
	if err != nil {
		return GetArticleUsecaseOutput{}, err
	}
	tagsByID := make(map[domain.TagID]domain.Tag, len(tags))
	for _, tag := range tags {
		tagsByID[tag.ID] = *tag
	}

	return GetArticleUsecaseOutput{
		Article: *article,
		Tags:    tagsByID,
	}, nil
}
