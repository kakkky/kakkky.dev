package usecase

import (
	"context"
	"slices"
	"time"

	"github.com/kakkky/kakkky.dev/domain"
)

type GetFeedUsecase struct {
	qs   domain.FeedQueryService
	repo domain.TagRepository
}

func (us *UseCase) NewGetFeedUsecase() *GetFeedUsecase {
	return &GetFeedUsecase{
		qs:   us.qs.NewFeedQueryService(),
		repo: us.repo.NewTagRepository(),
	}
}

type GetFeedUsecaseInput struct {
	Cursor GetFeedUsecaseCursor
	Limit  int
}

type GetFeedUsecaseOutput struct {
	Items      []domain.FeedItem
	Tags       map[domain.TagID]domain.Tag
	NextCursor GetFeedUsecaseCursor
}

type GetFeedUsecaseCursor struct {
	AfterID          domain.FeedItemID
	AfterPublishedAt time.Time
}

func (us *GetFeedUsecase) Exec(ctx context.Context, in GetFeedUsecaseInput) (*GetFeedUsecaseOutput, error) {
	items, err := us.qs.ListFeedItems(ctx, in.Cursor.AfterID, in.Cursor.AfterPublishedAt, in.Limit)
	if err != nil {
		return nil, err
	}

	var tagIDs []domain.TagID
	for _, it := range items {
		tagIDs = append(tagIDs, it.TagIDs...)
	}
	slices.Sort(tagIDs)
	tagIDs = slices.Compact(tagIDs)

	tags, err := us.repo.FindByIDs(ctx, tagIDs)
	if err != nil {
		return nil, err
	}

	tagMap := make(map[domain.TagID]domain.Tag, len(tags))
	for _, t := range tags {
		tagMap[t.ID] = *t
	}

	var next GetFeedUsecaseCursor
	if len(items) == in.Limit && in.Limit > 0 {
		last := items[len(items)-1]
		next = GetFeedUsecaseCursor{
			AfterID:          last.ID,
			AfterPublishedAt: last.PublishedAt,
		}
	}

	return &GetFeedUsecaseOutput{
		Items:      items,
		Tags:       tagMap,
		NextCursor: next,
	}, nil
}
