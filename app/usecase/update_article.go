package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/kakkky/kakkky.dev/domain"
)

type UpdateArticleUsecase struct {
	repo domain.Repository
}

func (us *UseCase) NewUpdateArticleUsecase() *UpdateArticleUsecase {
	return &UpdateArticleUsecase{
		repo: us.repo,
	}
}

type UpdateArticleUsecaseInput struct {
	// CurrentSlug は URL path 由来。DB 検索に使い、slug 自体は更新しない
	// (パーマリンク安定のため title から derive し直しはしない)
	CurrentSlug domain.Slug
	Title       string
	Body        string
	Status      string
	TagNames    []string
}

type UpdateArticleUsecaseOutput struct {
	Slug domain.Slug
}

func (us *UpdateArticleUsecase) Exec(ctx context.Context, in UpdateArticleUsecaseInput) (UpdateArticleUsecaseOutput, error) {
	status := domain.ArticleStatus(in.Status)

	out := UpdateArticleUsecaseOutput{}
	err := us.repo.WithTx(ctx, func(tx domain.Repository) error {
		articleRepo := tx.NewArticleRepository()
		current, err := articleRepo.FindBySlug(ctx, in.CurrentSlug)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.ErrNotFound.With("記事 が 見つかりません")
			}
			return err
		}

		publishedAt := current.PublishedAt
		if status == domain.ArticleStatusPublished && publishedAt.IsZero() {
			publishedAt = time.Now().UTC()
		}
		if status == domain.ArticleStatusDraft {
			publishedAt = time.Time{}
		}

		updated, err := domain.NewArticle(current.Slug, in.Title, in.Body, status, publishedAt)
		if err != nil {
			return err
		}
		updated.ID = current.ID
		updated.CreatedAt = current.CreatedAt

		tagIDs, err := resolveTagIDs(ctx, tx.NewTagRepository(), in.TagNames)
		if err != nil {
			return err
		}
		if err := updated.AddTags(tagIDs); err != nil {
			return err
		}
		if err := articleRepo.Update(ctx, updated); err != nil {
			return err
		}
		out.Slug = updated.Slug
		return nil
	})
	if err != nil {
		return UpdateArticleUsecaseOutput{}, err
	}
	return out, nil
}
