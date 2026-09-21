package domain

import "time"

type AnalyticsDateRange struct {
	From time.Time
	To   time.Time
}

type AnalyticsSiteMetrics struct {
	TotalUsers     int64
	TotalPageViews int64
	ByDate         []AnalyticsDailyPoint
	ByHour         []AnalyticsHourlyPoint
	TopReferrers   []AnalyticsReferrerCount
}

type AnalyticsArticleMetrics struct {
	Slug         Slug
	Users        int64
	PageViews    int64
	ByDate       []AnalyticsDailyPoint
	ByHour       []AnalyticsHourlyPoint
	TopReferrers []AnalyticsReferrerCount
}

type AnalyticsDailyPoint struct {
	Date      time.Time
	Users     int64
	PageViews int64
}

type AnalyticsHourlyPoint struct {
	Weekday   time.Weekday
	Hour      int
	Users     int64
	PageViews int64
}

type AnalyticsReferrerCount struct {
	Source   string // 流入元 (どこから来たか)。例: "google", "twitter.com", "(direct)"
	Medium   string // 流入媒体 (どういう経路で来たか)。例: "organic", "referral", "(none)"
	Sessions int64  // (Source, Medium) の組み合わせでの期間内セッション数
}
