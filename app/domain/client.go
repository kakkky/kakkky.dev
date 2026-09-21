package domain

import (
	"context"
	"io"
)

//go:generate mockgen -source=$GOFILE -destination=../testhelper/mock/mock_client.go -package=mock

type Client interface {
	NewS3Client() S3Client
	NewOGPFetcher() OGPFetcher
	NewGoogleAnalyticsClient() GoogleAnalyticsClient
}

type S3Client interface {
	Upload(ctx context.Context, contentType string, body io.Reader) (string, error)
}

type OGPFetcher interface {
	Fetch(ctx context.Context, url string) (OGPData, error)
}

type OGPData struct {
	Host        string
	Title       string
	Description string
	Image       string
}

type GoogleAnalyticsClient interface {
	FetchSiteMetrics(ctx context.Context, r AnalyticsDateRange) (AnalyticsSiteMetrics, error)
	FetchArticleMetrics(ctx context.Context, r AnalyticsDateRange) ([]AnalyticsArticleMetrics, error)
}
