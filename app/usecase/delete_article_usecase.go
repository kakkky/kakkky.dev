package usecase

import (
	"context"
	"errors"

	"github.com/kakkky/kakkky.dev/domain"
)

type DeleteArticleUsecase struct {
	repo domain.Repository
}

func (us *UseCase) NewDeleteArticleUsecase() *DeleteArticleUsecase {
	return &DeleteArticleUsecase{repo: us.repo}
}

type DeleteArticleUsecaseInput struct {
	ID domain.ArticleID
}

type DeleteArticleUsecaseOutput struct {
	Title string
}

func (us *DeleteArticleUsecase) Exec(ctx context.Context, in DeleteArticleUsecaseInput) (DeleteArticleUsecaseOutput, error) {
	if in.ID == "" {
		return DeleteArticleUsecaseOutput{}, domain.ErrInvalidArgument.With("id は 必須 です")
	}

	var out DeleteArticleUsecaseOutput
	err := us.repo.WithTx(ctx, func(tx domain.Repository) error {
		articleRepo := tx.NewArticleRepository()

		articles, err := articleRepo.FindByIDs(ctx, in.ID)
		if err != nil {
			return err
		}
		if len(articles) == 0 {
			return domain.ErrNotFound.With("article が 見つかりません")
		}
		out.Title = articles[0].Title

		if err := articleRepo.Delete(ctx, in.ID); err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return domain.ErrNotFound.With("article が 見つかりません")
			}
			return err
		}
		return nil
	})
	if err != nil {
		return DeleteArticleUsecaseOutput{}, err
	}
	return out, nil
}
