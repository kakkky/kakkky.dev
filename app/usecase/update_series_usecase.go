package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/kakkky/kakkky.dev/domain"
)

type UpdateSeriesUsecase struct {
	repo domain.Repository
}

func (us *UseCase) NewUpdateSeriesUsecase() *UpdateSeriesUsecase {
	return &UpdateSeriesUsecase{repo: us.repo}
}

type UpdateSeriesUsecaseInput struct {
	Slug           domain.Slug
	Title          string
	Description    string
	Status         domain.SeriesStatus
	ExistingTagIDs []domain.TagID
	NewTagNames    []string
	ArticleIDs     []domain.ArticleID
}

type UpdateSeriesUsecaseOutput struct {
	SeriesSlug domain.Slug
}

func (us *UpdateSeriesUsecase) Exec(ctx context.Context, in UpdateSeriesUsecaseInput) (UpdateSeriesUsecaseOutput, error) {
	if err := in.validate(); err != nil {
		return UpdateSeriesUsecaseOutput{}, err
	}

	var out UpdateSeriesUsecaseOutput
	err := us.repo.WithTx(ctx, func(tx domain.Repository) error {
		seriesRepo := tx.NewSeriesRepository()
		tagRepo := tx.NewTagRepository()

		series, err := seriesRepo.FindBySlug(ctx, in.Slug)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.ErrNotFound.With("series が 見つかりません")
			}
			return err
		}

		tagIDs, err := resolveTagIDs(ctx, tagRepo, in.ExistingTagIDs, in.NewTagNames)
		if err != nil {
			return err
		}
		if err := series.Update(in.Title, in.Description, in.Status, tagIDs); err != nil {
			return err
		}
		if err := series.ReorderArticles(in.ArticleIDs); err != nil {
			return err
		}

		if err := seriesRepo.Update(ctx, series); err != nil {
			return err
		}

		out.SeriesSlug = series.Slug
		return nil
	})
	if err != nil {
		return UpdateSeriesUsecaseOutput{}, err
	}
	return out, nil
}

func (in UpdateSeriesUsecaseInput) validate() error {
	if in.Slug == "" {
		return domain.ErrInvalidArgument.With("slug は 必須 です")
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
