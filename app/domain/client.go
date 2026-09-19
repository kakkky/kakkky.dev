package domain

import (
	"context"
)

//go:generate mockgen -source=$GOFILE -destination=../testhelper/mock/mock_client.go -package=mock

type Client interface {
	NewOGPFetcher() OGPFetcher
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
