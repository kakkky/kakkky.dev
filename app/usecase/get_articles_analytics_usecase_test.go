package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kakkky/kakkky.dev/adapter/cache"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/errors"
	"github.com/kakkky/kakkky.dev/testhelper/mock"
)

func TestGetArticlesAnalyticsUsecase_Exec(t *testing.T) {
	ctx := context.Background()

	r := domain.AnalyticsDateRange{
		From: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC),
	}
	key := analyticsCacheKey(r)

	slugA := domain.Slug("foo")
	slugB := domain.Slug("bar")
	slugMissing := domain.Slug("gone")

	articleA := &domain.Article{Slug: slugA, Title: "Foo"}
	articleB := &domain.Article{Slug: slugB, Title: "Bar"}

	raw := []domain.AnalyticsArticleMetrics{
		{Slug: slugA, Users: 100, PageViews: 300},
		{Slug: slugMissing, Users: 90, PageViews: 200},
		{Slug: slugB, Users: 80, PageViews: 100},
	}
	itemsFull := []ArticleAnalyticsItem{
		{Slug: slugA, Title: "Foo", Users: 100, PageViews: 300},
		{Slug: slugB, Title: "Bar", Users: 80, PageViews: 100},
	}
	itemsStale := []ArticleAnalyticsItem{
		{Slug: "stale-slug", Title: "Stale"},
	}

	fetchErr := errors.New("ga boom")
	repoErr := errors.New("db boom")

	tests := []struct {
		name          string
		mock          func(ga *mock.MockGoogleAnalyticsClient, ar *mock.MockArticleRepository)
		existingCache func(c domain.CacheClient)
		execTimes     int
		wantItems     []ArticleAnalyticsItem
		wantErr       error
	}{
		{
			name: "success: cache miss triggers fetch and drops slugs missing in DB",
			mock: func(ga *mock.MockGoogleAnalyticsClient, ar *mock.MockArticleRepository) {
				ga.EXPECT().FetchArticleMetrics(ctx, r).Return(raw, nil).Times(1)
				ar.EXPECT().FindBySlug(ctx, slugA).Return(articleA, nil)
				ar.EXPECT().FindBySlug(ctx, slugMissing).Return(nil, domain.ErrNotFound.With("not found"))
				ar.EXPECT().FindBySlug(ctx, slugB).Return(articleB, nil)
			},
			wantItems: itemsFull,
		},
		{
			name: "success: fresh cache hit skips fetch",
			existingCache: func(c domain.CacheClient) {
				c.Set(key, itemsFull, time.Now().Add(time.Hour))
			},
			wantItems: itemsFull,
		},
		{
			name: "success: expired cache is refreshed via fetch",
			existingCache: func(c domain.CacheClient) {
				c.Set(key, itemsStale, time.Now().Add(-time.Hour))
			},
			mock: func(ga *mock.MockGoogleAnalyticsClient, ar *mock.MockArticleRepository) {
				ga.EXPECT().FetchArticleMetrics(ctx, r).Return(raw, nil).Times(1)
				ar.EXPECT().FindBySlug(ctx, slugA).Return(articleA, nil)
				ar.EXPECT().FindBySlug(ctx, slugMissing).Return(nil, domain.ErrNotFound.With("not found"))
				ar.EXPECT().FindBySlug(ctx, slugB).Return(articleB, nil)
			},
			wantItems: itemsFull,
		},
		{
			name: "success: stale-while-error returns expired cache when fetch fails",
			existingCache: func(c domain.CacheClient) {
				c.Set(key, itemsFull, time.Now().Add(-time.Hour))
			},
			mock: func(ga *mock.MockGoogleAnalyticsClient, _ *mock.MockArticleRepository) {
				ga.EXPECT().FetchArticleMetrics(ctx, r).Return(nil, fetchErr).Times(1)
			},
			wantItems: itemsFull,
		},
		{
			name: "error: no cache and fetch error propagates",
			mock: func(ga *mock.MockGoogleAnalyticsClient, _ *mock.MockArticleRepository) {
				ga.EXPECT().FetchArticleMetrics(ctx, r).Return(nil, fetchErr).Times(1)
			},
			wantErr: fetchErr,
		},
		{
			name: "error: non-NotFound repository error propagates",
			mock: func(ga *mock.MockGoogleAnalyticsClient, ar *mock.MockArticleRepository) {
				ga.EXPECT().FetchArticleMetrics(ctx, r).Return(raw, nil).Times(1)
				ar.EXPECT().FindBySlug(ctx, slugA).Return(nil, repoErr)
			},
			wantErr: repoErr,
		},
		{
			name: "success: second Exec uses cache populated by first fetch",
			mock: func(ga *mock.MockGoogleAnalyticsClient, ar *mock.MockArticleRepository) {
				ga.EXPECT().FetchArticleMetrics(ctx, r).Return(raw, nil).Times(1)
				ar.EXPECT().FindBySlug(ctx, slugA).Return(articleA, nil).Times(1)
				ar.EXPECT().FindBySlug(ctx, slugMissing).Return(nil, domain.ErrNotFound.With("not found")).Times(1)
				ar.EXPECT().FindBySlug(ctx, slugB).Return(articleB, nil).Times(1)
			},
			execTimes: 2,
			wantItems: itemsFull,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			client := mock.NewMockClient(ctrl)
			ga := mock.NewMockGoogleAnalyticsClient(ctrl)
			client.EXPECT().NewGoogleAnalyticsClient().Return(ga)

			repo := mock.NewMockRepository(ctrl)
			ar := mock.NewMockArticleRepository(ctrl)
			repo.EXPECT().NewArticleRepository().Return(ar)

			if tt.mock != nil {
				tt.mock(ga, ar)
			}

			us := NewUseCase(repo, nil, client, cache.NewCache()).NewGetArticlesAnalyticsUsecase()
			if tt.existingCache != nil {
				tt.existingCache(us.cache)
			}

			n := tt.execTimes
			if n <= 0 {
				n = 1
			}
			var (
				out GetArticlesAnalyticsUsecaseOutput
				err error
			)
			for range n {
				out, err = us.Exec(ctx, GetArticlesAnalyticsUsecaseInput{Range: r})
			}

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, GetArticlesAnalyticsUsecaseOutput{}, out)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantItems, out.Items)
		})
	}
}
