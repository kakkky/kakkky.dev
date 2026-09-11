package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/kakkky/kakkky.dev/domain"
)

type UpdateArticleUsecase struct {
	repo domain.Repository
}

func (us *UseCase) NewUpdateArticleUsecase() *UpdateArticleUsecase {
	return &UpdateArticleUsecase{repo: us.repo}
}

type UpdateArticleUsecaseInput struct {
	Slug           domain.Slug
	Title          string
	Body           string
	Status         domain.ArticleStatus
	ExistingTagIDs []domain.TagID
	NewTagNames    []string
}

type UpdateArticleUsecaseOutput struct {
	ArticleSlug domain.Slug
}

func (us *UpdateArticleUsecase) Exec(ctx context.Context, in UpdateArticleUsecaseInput) (UpdateArticleUsecaseOutput, error) {
	if err := in.validate(); err != nil {
		return UpdateArticleUsecaseOutput{}, err
	}

	var out UpdateArticleUsecaseOutput
	err := us.repo.WithTx(ctx, func(tx domain.Repository) error {
		articleRepo := tx.NewArticleRepository()
		tagRepo := tx.NewTagRepository()

		article, err := articleRepo.FindBySlug(ctx, in.Slug)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.ErrNotFound.With("article が 見つかりません")
			}
			return err
		}

		tagIDs, err := resolveTagIDs(ctx, tagRepo, in.ExistingTagIDs, in.NewTagNames)
		if err != nil {
			return err
		}
		if err := article.Update(in.Title, in.Body, in.Status, tagIDs); err != nil {
			return err
		}

		if err := articleRepo.Update(ctx, article); err != nil {
			return err
		}

		out.ArticleSlug = article.Slug
		return nil
	})
	if err != nil {
		return UpdateArticleUsecaseOutput{}, err
	}
	return out, nil
}

func (in UpdateArticleUsecaseInput) validate() error {
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
