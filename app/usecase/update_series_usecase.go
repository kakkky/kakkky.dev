package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/kakkky/kakkky.dev/domain"
)

type UpdateSeriesUsecase struct {
	repo domain.Repository
}

func (us *UseCase) NewUpdateSeriesUsecase() *UpdateSeriesUsecase {
	return &UpdateSeriesUsecase{repo: us.repo}
}

type UpdateSeriesUsecaseInput struct {
	Slug              domain.Slug
	Title             string
	Description       string
	Status            domain.SeriesStatus
	ExistingTagIDs    []domain.TagID
	NewTagNames       []string
	OrderedArticleIDs []domain.ArticleID
	NewArticleTitles  []string
	DeleteArticleIDs  []domain.ArticleID
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
		articleRepo := tx.NewArticleRepository()
		tagRepo := tx.NewTagRepository()

		series, err := seriesRepo.FindBySlug(ctx, in.Slug)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.ErrNotFound.With("series が 見つかりません")
			}
			return err
		}

		for _, id := range in.DeleteArticleIDs {
			if err := articleRepo.Delete(ctx, id); err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					continue
				}
				return err
			}
			if err := series.RemoveArticle(id); err != nil {
				if errors.Is(err, domain.ErrInvalidArgument) {
					continue
				}
				return err
			}
		}

		newArticleIDs := make([]domain.ArticleID, 0, len(in.NewArticleTitles))
		for _, title := range in.NewArticleTitles {
			baseSlug, err := domain.GenerateSlug(title)
			if err != nil {
				return err
			}
			article, err := domain.NewArticle(baseSlug, title, "", domain.ArticleStatusDraft, time.Time{}, nil)
			if err != nil {
				return err
			}
			if err := articleRepo.Store(ctx, article); err != nil {
				if errors.Is(err, domain.ErrAlreadyExists) {
					return domain.ErrInvalidArgument.With(
						fmt.Sprintf("タイトル「%s」から生成した slug は 既に 存在 します", title),
					)
				}
				return err
			}
			if err := series.AddArticle(article.ID); err != nil {
				return err
			}
			newArticleIDs = append(newArticleIDs, article.ID)
		}

		tagIDs, err := resolveTagIDs(ctx, tagRepo, in.ExistingTagIDs, in.NewTagNames)
		if err != nil {
			return err
		}
		if err := series.Update(in.Title, in.Description, in.Status, tagIDs); err != nil {
			return err
		}

		finalArticleOrder := make([]domain.ArticleID, 0, len(in.OrderedArticleIDs)+len(newArticleIDs))
		finalArticleOrder = append(finalArticleOrder, in.OrderedArticleIDs...)
		finalArticleOrder = append(finalArticleOrder, newArticleIDs...)
		if err := series.ReorderArticles(finalArticleOrder); err != nil {
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
	for _, title := range in.NewArticleTitles {
		if title == "" {
			return domain.ErrInvalidArgument.With("新規記事タイトル に 空 が 含まれています")
		}
		if utf8.RuneCountInString(title) > domain.ArticleTitleMaxLength {
			return domain.ErrInvalidArgument.With(
				fmt.Sprintf("新規記事タイトル は %d 文字以内 です", domain.ArticleTitleMaxLength),
			)
		}
	}
	return nil
}
