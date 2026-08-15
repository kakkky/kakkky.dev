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

func TestGetFeedUsecase_Exec(t *testing.T) {
	ctx := context.Background()
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	var (
		tag1 domain.TagID = "11111111-1111-1111-1111-111111111111"
		tag2 domain.TagID = "22222222-2222-2222-2222-222222222222"

		item1ID domain.FeedItemID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa01"
		item2ID domain.FeedItemID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa02"
		item3ID domain.FeedItemID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa03"
	)

	tag1Val := &domain.Tag{ID: tag1, Slug: "go", Name: "Go"}
	tag2Val := &domain.Tag{ID: tag2, Slug: "db", Name: "DB"}
	allTags := []*domain.Tag{tag1Val, tag2Val}

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
	item3 := domain.FeedItem{
		Kind:        domain.FeedItemKindArticle,
		ID:          item3ID,
		Slug:        "a3",
		Title:       "A3",
		PublishedAt: baseTime.Add(-2 * time.Hour),
		TagIDs:      []domain.TagID{tag2},
	}

	wantErr := errors.New("boom")

	tests := []struct {
		name    string
		input   GetFeedUsecaseInput
		mock    func(qs *mock.MockFeedQueryService, repo *mock.MockTagRepository)
		want    GetFeedUsecaseOutput
		wantErr error
	}{
		{
			name:  "truncates to limit and sets NextCursor when ListFeedItems returns more than limit (limit+1 signal)",
			input: GetFeedUsecaseInput{Limit: 2},
			mock: func(qs *mock.MockFeedQueryService, repo *mock.MockTagRepository) {
				repo.EXPECT().ListAll(ctx).Return(allTags, nil)
				qs.EXPECT().
					ListFeedItems(ctx, []domain.TagID(nil), domain.FeedItemID(""), time.Time{}, 3).
					Return([]domain.FeedItem{item1, item2, item3}, nil)
			},
			want: GetFeedUsecaseOutput{
				Items: []domain.FeedItem{item1, item2},
				Tags:  allTags,
				NextCursor: GetFeedUsecaseCursor{
					AfterID:          item2.ID,
					AfterPublishedAt: item2.PublishedAt,
				},
			},
		},
		{
			name:  "leaves NextCursor zero when items exactly match the limit and there is no more",
			input: GetFeedUsecaseInput{Limit: 2},
			mock: func(qs *mock.MockFeedQueryService, repo *mock.MockTagRepository) {
				repo.EXPECT().ListAll(ctx).Return(allTags, nil)
				qs.EXPECT().
					ListFeedItems(ctx, []domain.TagID(nil), domain.FeedItemID(""), time.Time{}, 3).
					Return([]domain.FeedItem{item1, item2}, nil)
			},
			want: GetFeedUsecaseOutput{
				Items: []domain.FeedItem{item1, item2},
				Tags:  allTags,
			},
		},
		{
			name:  "leaves NextCursor zero when items are fewer than the limit",
			input: GetFeedUsecaseInput{Limit: 10},
			mock: func(qs *mock.MockFeedQueryService, repo *mock.MockTagRepository) {
				repo.EXPECT().ListAll(ctx).Return(allTags, nil)
				qs.EXPECT().
					ListFeedItems(ctx, []domain.TagID(nil), domain.FeedItemID(""), time.Time{}, 11).
					Return([]domain.FeedItem{item1}, nil)
			},
			want: GetFeedUsecaseOutput{
				Items: []domain.FeedItem{item1},
				Tags:  allTags,
			},
		},
		{
			name: "passes cursor through to FeedQueryService.ListFeedItems",
			input: GetFeedUsecaseInput{
				Cursor: GetFeedUsecaseCursor{
					AfterID:          item1ID,
					AfterPublishedAt: baseTime,
				},
				Limit: 10,
			},
			mock: func(qs *mock.MockFeedQueryService, repo *mock.MockTagRepository) {
				repo.EXPECT().ListAll(ctx).Return(allTags, nil)
				qs.EXPECT().
					ListFeedItems(ctx, []domain.TagID(nil), item1ID, baseTime, 11).
					Return([]domain.FeedItem{item2}, nil)
			},
			want: GetFeedUsecaseOutput{
				Items: []domain.FeedItem{item2},
				Tags:  allTags,
			},
		},
		{
			name:  "resolves TagSlugs to TagIDs via the fetched tag list and passes them to ListFeedItems",
			input: GetFeedUsecaseInput{TagSlugs: []domain.Slug{"go", "db"}, Limit: 10},
			mock: func(qs *mock.MockFeedQueryService, repo *mock.MockTagRepository) {
				repo.EXPECT().ListAll(ctx).Return(allTags, nil)
				qs.EXPECT().
					ListFeedItems(ctx, []domain.TagID{tag1, tag2}, domain.FeedItemID(""), time.Time{}, 11).
					Return([]domain.FeedItem{item2}, nil)
			},
			want: GetFeedUsecaseOutput{
				Items: []domain.FeedItem{item2},
				Tags:  allTags,
			},
		},
		{
			name:  "ignores unknown slugs when resolving TagSlugs",
			input: GetFeedUsecaseInput{TagSlugs: []domain.Slug{"go", "unknown"}, Limit: 10},
			mock: func(qs *mock.MockFeedQueryService, repo *mock.MockTagRepository) {
				repo.EXPECT().ListAll(ctx).Return(allTags, nil)
				qs.EXPECT().
					ListFeedItems(ctx, []domain.TagID{tag1}, domain.FeedItemID(""), time.Time{}, 11).
					Return([]domain.FeedItem{item2}, nil)
			},
			want: GetFeedUsecaseOutput{
				Items: []domain.FeedItem{item2},
				Tags:  allTags,
			},
		},
		{
			name:  "propagates error from TagRepository.ListAll and skips ListFeedItems",
			input: GetFeedUsecaseInput{Limit: 10},
			mock: func(qs *mock.MockFeedQueryService, repo *mock.MockTagRepository) {
				repo.EXPECT().ListAll(ctx).Return(nil, wantErr)
			},
			wantErr: wantErr,
		},
		{
			name:  "propagates error from FeedQueryService.ListFeedItems",
			input: GetFeedUsecaseInput{Limit: 10},
			mock: func(qs *mock.MockFeedQueryService, repo *mock.MockTagRepository) {
				repo.EXPECT().ListAll(ctx).Return(allTags, nil)
				qs.EXPECT().
					ListFeedItems(ctx, []domain.TagID(nil), domain.FeedItemID(""), time.Time{}, 11).
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

			gf := NewUseCase(repo, qs).NewGetFeedUsecase()
			got, err := gf.Exec(ctx, tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, GetFeedUsecaseOutput{}, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
