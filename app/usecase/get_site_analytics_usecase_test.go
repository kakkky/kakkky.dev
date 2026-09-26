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

func TestGetSiteAnalyticsUsecase_Exec(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	r := domain.AnalyticsDateRange{
		From: time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 9, 22, 0, 0, 0, 0, time.UTC),
	}
	key := analyticsCacheKey(r)
	fresh := domain.AnalyticsSiteMetrics{TotalUsers: 100, TotalPageViews: 200}
	stale := domain.AnalyticsSiteMetrics{TotalUsers: 50, TotalPageViews: 90}
	fetchErr := errors.New("ga boom")

	tests := []struct {
		name          string
		mock          func(ga *mock.MockGoogleAnalyticsClient)
		existingCache func(c domain.CacheClient)
		execTimes     int
		wantData      domain.AnalyticsSiteMetrics
		wantErr       error
	}{
		{
			name: "success: cache miss triggers fetch",
			mock: func(ga *mock.MockGoogleAnalyticsClient) {
				ga.EXPECT().FetchSiteMetrics(ctx, r).Return(fresh, nil).Times(1)
			},
			wantData: fresh,
		},
		{
			name: "success: fresh cache hit skips fetch",
			existingCache: func(c domain.CacheClient) {
				c.Set(key, fresh, time.Now().Add(time.Hour))
			},
			wantData: fresh,
		},
		{
			name: "success: expired cache is refreshed via fetch",
			existingCache: func(c domain.CacheClient) {
				c.Set(key, stale, time.Now().Add(-time.Hour))
			},
			mock: func(ga *mock.MockGoogleAnalyticsClient) {
				ga.EXPECT().FetchSiteMetrics(ctx, r).Return(fresh, nil).Times(1)
			},
			wantData: fresh,
		},
		{
			name: "success: stale-while-error returns expired cache when fetch fails",
			existingCache: func(c domain.CacheClient) {
				c.Set(key, stale, time.Now().Add(-time.Hour))
			},
			mock: func(ga *mock.MockGoogleAnalyticsClient) {
				ga.EXPECT().FetchSiteMetrics(ctx, r).Return(domain.AnalyticsSiteMetrics{}, fetchErr).Times(1)
			},
			wantData: stale,
		},
		{
			name: "error: no cache and fetch error propagates",
			mock: func(ga *mock.MockGoogleAnalyticsClient) {
				ga.EXPECT().FetchSiteMetrics(ctx, r).Return(domain.AnalyticsSiteMetrics{}, fetchErr).Times(1)
			},
			wantErr: fetchErr,
		},
		{
			name: "success: second Exec uses cache populated by first fetch",
			mock: func(ga *mock.MockGoogleAnalyticsClient) {
				ga.EXPECT().FetchSiteMetrics(ctx, r).Return(fresh, nil).Times(1)
			},
			execTimes: 2,
			wantData:  fresh,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			client := mock.NewMockClient(ctrl)
			ga := mock.NewMockGoogleAnalyticsClient(ctrl)
			client.EXPECT().NewGoogleAnalyticsClient().Return(ga)
			if tt.mock != nil {
				tt.mock(ga)
			}

			us := NewUseCase(nil, nil, client, cache.NewCache()).NewGetSiteAnalyticsUsecase()
			if tt.existingCache != nil {
				tt.existingCache(us.cache)
			}

			n := tt.execTimes
			if n <= 0 {
				n = 1
			}
			var (
				out GetSiteAnalyticsUsecaseOutput
				err error
			)
			for range n {
				out, err = us.Exec(ctx, GetSiteAnalyticsUsecaseInput{Range: r})
			}

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, GetSiteAnalyticsUsecaseOutput{}, out)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantData, out.Metrics)
		})
	}
}
