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

type CreateArticleUsecase struct {
	repo domain.Repository
}

func (us *UseCase) NewCreateArticleUsecase() *CreateArticleUsecase {
	return &CreateArticleUsecase{repo: us.repo}
}

type CreateArticleUsecaseInput struct {
	Title          string
	ExistingTagIDs []domain.TagID
	NewTagNames    []string
}

type CreateArticleUsecaseOutput struct {
	ArticleSlug domain.Slug
}

func (us *CreateArticleUsecase) Exec(ctx context.Context, in CreateArticleUsecaseInput) (CreateArticleUsecaseOutput, error) {
	if err := in.validate(); err != nil {
		return CreateArticleUsecaseOutput{}, err
	}
	baseSlug, err := domain.GenerateSlug(in.Title)
	if err != nil {
		return CreateArticleUsecaseOutput{}, err
	}

	var out CreateArticleUsecaseOutput
	err = us.repo.WithTx(ctx, func(tx domain.Repository) error {
		tagRepo := tx.NewTagRepository()
		articleRepo := tx.NewArticleRepository()

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

		article, err := domain.NewArticle(baseSlug, in.Title, "", domain.ArticleStatusDraft, time.Time{})
		if err != nil {
			return err
		}
		tagIDs := append(slices.Clone(in.ExistingTagIDs), newTagIDs...)
		if err := article.AddTags(tagIDs); err != nil {
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
		out.ArticleSlug = article.Slug
		return nil
	})
	if err != nil {
		return CreateArticleUsecaseOutput{}, err
	}
	return out, nil
}

func (in CreateArticleUsecaseInput) validate() error {
	if in.Title == "" {
		return domain.ErrInvalidArgument.With("タイトル は 必須 です")
	}
	if utf8.RuneCountInString(in.Title) > domain.ArticleTitleMaxLength {
		return domain.ErrInvalidArgument.With(
			fmt.Sprintf("タイトル は %d 文字以内 です", domain.ArticleTitleMaxLength),
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
