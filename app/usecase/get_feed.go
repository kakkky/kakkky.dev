package usecase

import (
	"context"
	"time"

	"github.com/kakkky/kakkky.dev/domain"
)

type GetFeedUsecase struct {
	feedQS  domain.FeedQueryService
	tagRepo domain.TagRepository
}

func (us *UseCase) NewGetFeedUsecase() *GetFeedUsecase {
	return &GetFeedUsecase{
		feedQS:  us.qs.NewFeedQueryService(),
		tagRepo: us.repo.NewTagRepository(),
	}
}

type GetFeedUsecaseInput struct {
	TagSlugs []domain.Slug
	Cursor   GetFeedUsecaseCursor
	Limit    int
}

type GetFeedUsecaseOutput struct {
	Items      []domain.FeedItem
	Tags       []domain.Tag
	NextCursor GetFeedUsecaseCursor
}

type GetFeedUsecaseCursor struct {
	AfterID          domain.FeedItemID
	AfterPublishedAt time.Time
}

func (us *GetFeedUsecase) Exec(ctx context.Context, in GetFeedUsecaseInput) (GetFeedUsecaseOutput, error) {
	tagRows, err := us.tagRepo.List(ctx)
	if err != nil {
		return GetFeedUsecaseOutput{}, err
	}
	allTags := make([]domain.Tag, len(tagRows))
	for i, t := range tagRows {
		allTags[i] = *t
	}

	var filterTagIDs []domain.TagID
	if len(in.TagSlugs) > 0 {
		tagBySlug := make(map[domain.Slug]domain.Tag, len(allTags))
		for _, t := range allTags {
			tagBySlug[t.Slug] = t
		}
		filterTagIDs = make([]domain.TagID, 0, len(in.TagSlugs))
		for _, s := range in.TagSlugs {
			if t, ok := tagBySlug[s]; ok {
				filterTagIDs = append(filterTagIDs, t.ID)
			}
		}
	}

	fetchLimit := in.Limit
	if in.Limit > 0 {
		fetchLimit = in.Limit + 1
	}
	items, err := us.feedQS.ListFeedItems(ctx, filterTagIDs, in.Cursor.AfterID, in.Cursor.AfterPublishedAt, fetchLimit)
	if err != nil {
		return GetFeedUsecaseOutput{}, err
	}

	var next GetFeedUsecaseCursor
	if in.Limit > 0 && len(items) > in.Limit {
		items = items[:in.Limit]
		last := items[len(items)-1]
		next = GetFeedUsecaseCursor{
			AfterID:          last.ID,
			AfterPublishedAt: last.PublishedAt,
		}
	}

	return GetFeedUsecaseOutput{
		Items:      items,
		Tags:       allTags,
		NextCursor: next,
	}, nil
}
