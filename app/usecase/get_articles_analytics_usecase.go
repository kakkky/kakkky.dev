package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/errors"
)

// Google Analytics のデータは反映に2~6時間かかるため、キャッシュを30分程度持たせる。
// ref: https://support.google.com/analytics/answer/11198161
const articlesAnalyticsCacheTTL = 30 * time.Minute

type GetArticlesAnalyticsUsecase struct {
	gaClient    domain.GoogleAnalyticsClient
	articleRepo domain.ArticleRepository
	ttl         time.Duration
	cache       domain.CacheClient
}

func (us *UseCase) NewGetArticlesAnalyticsUsecase() *GetArticlesAnalyticsUsecase {
	return &GetArticlesAnalyticsUsecase{
		gaClient:    us.client.NewGoogleAnalyticsClient(),
		articleRepo: us.repo.NewArticleRepository(),
		ttl:         articlesAnalyticsCacheTTL,
		cache:       us.cache.NewInMemoryCacheClient(),
	}
}

type GetArticlesAnalyticsUsecaseInput struct {
	Range domain.AnalyticsDateRange
}

type GetArticlesAnalyticsUsecaseOutput struct {
	Items []ArticleAnalyticsItem
}

type ArticleAnalyticsItem struct {
	Slug      domain.Slug
	Title     string
	Users     int64
	PageViews int64
	ByDate    []domain.AnalyticsDailyPoint
}

func (us *GetArticlesAnalyticsUsecase) Exec(ctx context.Context, in GetArticlesAnalyticsUsecaseInput) (GetArticlesAnalyticsUsecaseOutput, error) {
	key := analyticsCacheKey(in.Range)

	if v, ok := us.cache.Get(key); ok {
		return GetArticlesAnalyticsUsecaseOutput{Items: v.([]ArticleAnalyticsItem)}, nil
	}

	m, err := us.gaClient.FetchArticleMetrics(ctx, in.Range)
	if err != nil {
		if v, ok := us.cache.GetStale(key); ok {
			slog.WarnContext(ctx, "articles analytics: fetch failed, serving stale cache", "err", err)
			return GetArticlesAnalyticsUsecaseOutput{Items: v.([]ArticleAnalyticsItem)}, nil
		}
		return GetArticlesAnalyticsUsecaseOutput{}, err
	}

	// slug で Article を引き Title を解決する
	items := make([]ArticleAnalyticsItem, 0, len(m))
	for _, metric := range m {
		if metric.Slug == "" {
			continue
		}
		a, err := us.articleRepo.FindBySlug(ctx, metric.Slug)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				continue
			}
			return GetArticlesAnalyticsUsecaseOutput{}, err
		}
		items = append(items, ArticleAnalyticsItem{
			Slug:      metric.Slug,
			Title:     a.Title,
			Users:     metric.Users,
			PageViews: metric.PageViews,
			ByDate:    metric.ByDate,
		})
	}

	us.cache.Set(key, items, time.Now().Add(us.ttl))
	return GetArticlesAnalyticsUsecaseOutput{Items: items}, nil
}
