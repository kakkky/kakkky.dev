package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/kakkky/kakkky.dev/domain"
)

type CreateArticleInSeriesUsecase struct {
	repo domain.Repository
}

func (us *UseCase) NewCreateArticleInSeriesUsecase() *CreateArticleInSeriesUsecase {
	return &CreateArticleInSeriesUsecase{repo: us.repo}
}

type CreateArticleInSeriesUsecaseInput struct {
	SeriesSlug domain.Slug
	Title      string
}

type CreateArticleInSeriesUsecaseOutput struct {
	ArticleID    domain.ArticleID
	ArticleSlug  domain.Slug
	ArticleTitle string
	Position     int
}

func (us *CreateArticleInSeriesUsecase) Exec(ctx context.Context, in CreateArticleInSeriesUsecaseInput) (CreateArticleInSeriesUsecaseOutput, error) {
	if err := in.validate(); err != nil {
		return CreateArticleInSeriesUsecaseOutput{}, err
	}
	baseSlug, err := domain.GenerateSlug(in.Title)
	if err != nil {
		return CreateArticleInSeriesUsecaseOutput{}, err
	}

	var out CreateArticleInSeriesUsecaseOutput
	err = us.repo.WithTx(ctx, func(tx domain.Repository) error {
		seriesRepo := tx.NewSeriesRepository()
		articleRepo := tx.NewArticleRepository()

		series, err := seriesRepo.FindBySlug(ctx, in.SeriesSlug)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.ErrNotFound.With("series が 見つかりません")
			}
			return err
		}

		article, err := domain.NewArticle(baseSlug, in.Title, "", domain.ArticleStatusDraft, time.Time{}, nil)
		if err != nil {
			return err
		}
		if err := articleRepo.Store(ctx, article); err != nil {
			if errors.Is(err, domain.ErrAlreadyExists) {
				return domain.ErrInvalidArgument.With(
					fmt.Sprintf("タイトル「%s」から生成した slug は 既に 存在 します", in.Title),
				)
			}
			return err
		}

		if err := series.AddArticle(article.ID); err != nil {
			return err
		}
		if err := seriesRepo.Update(ctx, series); err != nil {
			return err
		}

		out.ArticleID = article.ID
		out.ArticleSlug = article.Slug
		out.ArticleTitle = article.Title
		out.Position = series.Articles[len(series.Articles)-1].Position
		return nil
	})
	if err != nil {
		return CreateArticleInSeriesUsecaseOutput{}, err
	}
	return out, nil
}

func (in CreateArticleInSeriesUsecaseInput) validate() error {
	if in.SeriesSlug == "" {
		return domain.ErrInvalidArgument.With("series slug は 必須 です")
	}
	if in.Title == "" {
		return domain.ErrInvalidArgument.With("タイトル は 必須 です")
	}
	if utf8.RuneCountInString(in.Title) > domain.ArticleTitleMaxLength {
		return domain.ErrInvalidArgument.With(
			fmt.Sprintf("タイトル は %d 文字以内 です", domain.ArticleTitleMaxLength),
		)
	}
	return nil
}
