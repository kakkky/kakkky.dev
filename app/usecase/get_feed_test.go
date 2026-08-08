package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/testhelper/mock"
)

func TestGetFeed_Exec(t *testing.T) {
	ctx := context.Background()
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	var (
		tag1 domain.TagID = "11111111-1111-1111-1111-111111111111"
		tag2 domain.TagID = "22222222-2222-2222-2222-222222222222"

		item1ID domain.FeedItemID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa01"
		item2ID domain.FeedItemID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa02"
	)

	tag1Val := domain.Tag{ID: tag1, Slug: "go", Name: "Go"}
	tag2Val := domain.Tag{ID: tag2, Slug: "db", Name: "DB"}

	item1 := domain.FeedItem{
		Kind:        domain.FeedItemKindArticle,
		ID:          item1ID,
		Slug:        "a1",
		Title:       "A1",
		PublishedAt: baseTime,
		TagIDs:      []domain.TagID{tag1, tag2},
	}
	item2 := domain.FeedItem{
		Kind:        domain.FeedItemKindArticle,
		ID:          item2ID,
		Slug:        "a2",
		Title:       "A2",
		PublishedAt: baseTime.Add(-1 * time.Hour),
		TagIDs:      []domain.TagID{tag1},
	}

	wantErr := errors.New("boom")

	tests := []struct {
		name    string
		input   GetFeedInput
		mock    func(qs *mock.MockFeedQueryService, repo *mock.MockTagRepository)
		want    *GetFeedOutput
		wantErr error
	}{
		{
			name:  "enriches tags and sets NextCursor when items reach the limit",
			input: GetFeedInput{Limit: 2},
			mock: func(qs *mock.MockFeedQueryService, repo *mock.MockTagRepository) {
				qs.EXPECT().
					ListFeedItems(ctx, domain.FeedItemID(""), time.Time{}, 2).
					Return([]domain.FeedItem{item1, item2}, nil)
				repo.EXPECT().
					FindByIDs(ctx, []domain.TagID{tag1, tag2}).
					Return([]*domain.Tag{&tag1Val, &tag2Val}, nil)
			},
			want: &GetFeedOutput{
				Items: []domain.FeedItem{item1, item2},
				Tags: map[domain.TagID]domain.Tag{
					tag1: tag1Val,
					tag2: tag2Val,
				},
				NextCursor: GetFeedCursor{
					AfterID:          item2.ID,
					AfterPublishedAt: item2.PublishedAt,
				},
			},
		},
		{
			name:  "leaves NextCursor zero when items are fewer than the limit",
			input: GetFeedInput{Limit: 10},
			mock: func(qs *mock.MockFeedQueryService, repo *mock.MockTagRepository) {
				qs.EXPECT().
					ListFeedItems(ctx, domain.FeedItemID(""), time.Time{}, 10).
					Return([]domain.FeedItem{item1}, nil)
				repo.EXPECT().
					FindByIDs(ctx, []domain.TagID{tag1, tag2}).
					Return([]*domain.Tag{&tag1Val, &tag2Val}, nil)
			},
			want: &GetFeedOutput{
				Items: []domain.FeedItem{item1},
				Tags: map[domain.TagID]domain.Tag{
					tag1: tag1Val,
					tag2: tag2Val,
				},
			},
		},
		{
			name: "passes cursor through to FeedQueryService.ListFeedItems",
			input: GetFeedInput{
				Cursor: GetFeedCursor{
					AfterID:          item1ID,
					AfterPublishedAt: baseTime,
				},
				Limit: 10,
			},
			mock: func(qs *mock.MockFeedQueryService, repo *mock.MockTagRepository) {
				qs.EXPECT().
					ListFeedItems(ctx, item1ID, baseTime, 10).
					Return([]domain.FeedItem{item2}, nil)
				repo.EXPECT().
					FindByIDs(ctx, []domain.TagID{tag1}).
					Return([]*domain.Tag{&tag1Val}, nil)
			},
			want: &GetFeedOutput{
				Items: []domain.FeedItem{item2},
				Tags: map[domain.TagID]domain.Tag{
					tag1: tag1Val,
				},
			},
		},
		{
			name:  "propagates error from FeedQueryService.ListFeedItems and skips tag fetch",
			input: GetFeedInput{Limit: 10},
			mock: func(qs *mock.MockFeedQueryService, repo *mock.MockTagRepository) {
				qs.EXPECT().
					ListFeedItems(ctx, domain.FeedItemID(""), time.Time{}, 10).
					Return(nil, wantErr)
			},
			wantErr: wantErr,
		},
		{
			name:  "propagates error from TagRepository.FindByIDs",
			input: GetFeedInput{Limit: 10},
			mock: func(qs *mock.MockFeedQueryService, repo *mock.MockTagRepository) {
				qs.EXPECT().
					ListFeedItems(ctx, domain.FeedItemID(""), time.Time{}, 10).
					Return([]domain.FeedItem{item1}, nil)
				repo.EXPECT().
					FindByIDs(ctx, []domain.TagID{tag1, tag2}).
					Return(nil, wantErr)
			},
			wantErr: wantErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			fqs := mock.NewMockFeedQueryService(ctrl)
			qs := mock.NewMockQueryService(ctrl)
			qs.EXPECT().NewFeedQueryService().Return(fqs)

			tr := mock.NewMockTagRepository(ctrl)
			repo := mock.NewMockRepository(ctrl)
			repo.EXPECT().NewTagRepository().Return(tr)

			tt.mock(fqs, tr)

			gf := NewUseCase(repo, qs).NewGetFeed()
			got, err := gf.Exec(ctx, tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
