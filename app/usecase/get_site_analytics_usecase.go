package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/kakkky/kakkky.dev/domain"
)

// Google Analytics のデータは反映に2~6時間かかるため、キャッシュを30分程度持たせる。
// ref: https://support.google.com/analytics/answer/11198161
const siteAnalyticsCacheTTL = 30 * time.Minute

type GetSiteAnalyticsUsecase struct {
	gaClient domain.GoogleAnalyticsClient
	ttl      time.Duration
	cache    domain.CacheClient
}

func (us *UseCase) NewGetSiteAnalyticsUsecase() *GetSiteAnalyticsUsecase {
	return &GetSiteAnalyticsUsecase{
		gaClient: us.client.NewGoogleAnalyticsClient(),
		ttl:      siteAnalyticsCacheTTL,
		cache:    us.cache.NewInMemoryCacheClient(),
	}
}

type GetSiteAnalyticsUsecaseInput struct {
	Range domain.AnalyticsDateRange
}

type GetSiteAnalyticsUsecaseOutput struct {
	Metrics domain.AnalyticsSiteMetrics
}

func (us *GetSiteAnalyticsUsecase) Exec(ctx context.Context, in GetSiteAnalyticsUsecaseInput) (GetSiteAnalyticsUsecaseOutput, error) {
	key := analyticsCacheKey(in.Range)

	if v, ok := us.cache.Get(key); ok {
		return GetSiteAnalyticsUsecaseOutput{Metrics: v.(domain.AnalyticsSiteMetrics)}, nil
	}

	m, err := us.gaClient.FetchSiteMetrics(ctx, in.Range)
	if err != nil {
		if v, ok := us.cache.GetStale(key); ok {
			slog.WarnContext(ctx, "site analytics: fetch failed, serving stale cache", "err", err)
			return GetSiteAnalyticsUsecaseOutput{Metrics: v.(domain.AnalyticsSiteMetrics)}, nil
		}
		return GetSiteAnalyticsUsecaseOutput{}, err
	}

	us.cache.Set(key, m, time.Now().Add(us.ttl))
	return GetSiteAnalyticsUsecaseOutput{Metrics: m}, nil
}

// TZ ずれ回避のため UTC の YYYY-MM-DD で組む。
func analyticsCacheKey(r domain.AnalyticsDateRange) string {
	return r.From.UTC().Format("2006-01-02") + "_" + r.To.UTC().Format("2006-01-02")
}
