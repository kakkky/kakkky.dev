package usecase

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/kakkky/kakkky.dev/domain"
)

// resolveTagIDs は 既存 tag ID + 新規タグ名 から 最終的 な tag ID 一覧 を 返す。
// 新規タグ は tagRepo.Store で 永続化。slug 衝突 は ErrInvalidArgument に 翻訳。
func resolveTagIDs(
	ctx context.Context,
	tagRepo domain.TagRepository,
	existingIDs []domain.TagID,
	newNames []string,
) ([]domain.TagID, error) {
	newIDs := make([]domain.TagID, 0, len(newNames))
	for _, name := range newNames {
		slug, err := domain.GenerateSlug(name)
		if err != nil {
			return nil, err
		}
		tag, err := domain.NewTag(slug, name)
		if err != nil {
			return nil, err
		}
		if err := tagRepo.Store(ctx, tag); err != nil {
			if errors.Is(err, domain.ErrAlreadyExists) {
				return nil, domain.ErrInvalidArgument.With(
					fmt.Sprintf("タグ「%s」は 既に 存在 します", name),
				)
			}
			return nil, err
		}
		newIDs = append(newIDs, tag.ID)
	}
	return append(slices.Clone(existingIDs), newIDs...), nil
}
