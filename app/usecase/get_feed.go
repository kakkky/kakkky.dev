package usecase

import (
	"context"
	"slices"
	"time"

	"github.com/kakkky/kakkky.dev/domain"
)

type GetFeed struct {
	qs   domain.FeedQueryService
	repo domain.TagRepository
}

func (us *UseCase) NewGetFeed() *GetFeed {
	return &GetFeed{
		qs:   us.qs.NewFeedQueryService(),
		repo: us.repo.NewTagRepository(),
	}
}

type GetFeedInput struct {
	Cursor GetFeedCursor
	Limit  int
}

type GetFeedOutput struct {
	Items      []domain.FeedItem
	Tags       map[domain.TagID]domain.Tag
	NextCursor GetFeedCursor
}

type GetFeedCursor struct {
	AfterID          domain.FeedItemID
	AfterPublishedAt time.Time
}

func (us *GetFeed) Exec(ctx context.Context, in GetFeedInput) (*GetFeedOutput, error) {
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

	var next GetFeedCursor
	if len(items) == in.Limit && in.Limit > 0 {
		last := items[len(items)-1]
		next = GetFeedCursor{
			AfterID:          last.ID,
			AfterPublishedAt: last.PublishedAt,
		}
	}

	return &GetFeedOutput{
		Items:      items,
		Tags:       tagMap,
		NextCursor: next,
	}, nil
}
