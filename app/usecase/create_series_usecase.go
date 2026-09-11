package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"
	"unicode/utf8"

	"github.com/kakkky/kakkky.dev/domain"
)

type CreateSeriesUsecase struct {
	repo domain.Repository
}

func (us *UseCase) NewCreateSeriesUsecase() *CreateSeriesUsecase {
	return &CreateSeriesUsecase{repo: us.repo}
}

type CreateSeriesUsecaseInput struct {
	Title          string
	Description    string
	ExistingTagIDs []domain.TagID
	NewTagNames    []string
}

type CreateSeriesUsecaseOutput struct {
	SeriesSlug domain.Slug
}

func (us *CreateSeriesUsecase) Exec(ctx context.Context, in CreateSeriesUsecaseInput) (CreateSeriesUsecaseOutput, error) {
	if err := in.validate(); err != nil {
		return CreateSeriesUsecaseOutput{}, err
	}
	baseSlug, err := domain.GenerateSlug(in.Title)
	if err != nil {
		return CreateSeriesUsecaseOutput{}, err
	}

	var out CreateSeriesUsecaseOutput
	err = us.repo.WithTx(ctx, func(tx domain.Repository) error {
		tagRepo := tx.NewTagRepository()
		seriesRepo := tx.NewSeriesRepository()

		newTagIDs := make([]domain.TagID, 0, len(in.NewTagNames))
		for _, name := range in.NewTagNames {
			slug, err := domain.GenerateSlug(name)
			if err != nil {
				return err
			}
			tag, err := domain.NewTag(slug, name)
			if err != nil {
				return err
			}
			if err := tagRepo.Store(ctx, tag); err != nil {
				if errors.Is(err, domain.ErrAlreadyExists) {
					return domain.ErrInvalidArgument.With(
						fmt.Sprintf("タグ「%s」は 既に 存在 します", name),
					)
				}
				return err
			}
			newTagIDs = append(newTagIDs, tag.ID)
		}

		tagIDs := append(slices.Clone(in.ExistingTagIDs), newTagIDs...)
		series, err := domain.NewSeries(baseSlug, in.Title, in.Description, domain.SeriesStatusDraft, time.Time{})
		if err != nil {
			return err
		}
		if err := series.AddTags(tagIDs); err != nil {
			return err
		}

		if err := seriesRepo.Store(ctx, series); err != nil {
			if errors.Is(err, domain.ErrAlreadyExists) {
				return domain.ErrInvalidArgument.With(
					fmt.Sprintf("タイトル「%s」から生成した slug は 既に 存在 します", in.Title),
				)
			}
			return err
		}
		out.SeriesSlug = series.Slug
		return nil
	})
	if err != nil {
		return CreateSeriesUsecaseOutput{}, err
	}
	return out, nil
}

func (in CreateSeriesUsecaseInput) validate() error {
	if in.Title == "" {
		return domain.ErrInvalidArgument.With("タイトル は 必須 です")
	}
	if utf8.RuneCountInString(in.Title) > domain.SeriesTitleMaxLength {
		return domain.ErrInvalidArgument.With(
			fmt.Sprintf("タイトル は %d 文字以内 です", domain.SeriesTitleMaxLength),
		)
	}
	if utf8.RuneCountInString(in.Description) > domain.SeriesDescriptionMaxLength {
		return domain.ErrInvalidArgument.With(
			fmt.Sprintf("説明 は %d 文字以内 です", domain.SeriesDescriptionMaxLength),
		)
	}
	seen := make(map[string]struct{}, len(in.NewTagNames))
	for _, name := range in.NewTagNames {
		if name == "" {
			return domain.ErrInvalidArgument.With("新規タグ名 に 空 が 含まれています")
		}
		if _, ok := seen[name]; ok {
			return domain.ErrInvalidArgument.With(
				fmt.Sprintf("新規タグ「%s」が 重複 しています", name),
			)
		}
		seen[name] = struct{}{}
	}
	return nil
}
