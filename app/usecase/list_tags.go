package usecase

import (
	"context"

	"github.com/kakkky/kakkky.dev/domain"
)

type ListTagsUsecase struct {
	tagRepo domain.TagRepository
}

func (us *UseCase) NewListTagsUsecase() *ListTagsUsecase {
	return &ListTagsUsecase{
		tagRepo: us.repo.NewTagRepository(),
	}
}

type ListTagsUsecaseOutput struct {
	Tags []domain.Tag
}

func (us *ListTagsUsecase) Exec(ctx context.Context) (ListTagsUsecaseOutput, error) {
	tags, err := us.tagRepo.List(ctx)
	if err != nil {
		return ListTagsUsecaseOutput{}, err
	}
	out := make([]domain.Tag, len(tags))
	for i, t := range tags {
		out[i] = *t
	}
	return ListTagsUsecaseOutput{Tags: out}, nil
}
