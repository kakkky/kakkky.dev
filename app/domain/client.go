package domain

import (
	"context"
	"io"
)

//go:generate mockgen -source=$GOFILE -destination=../testhelper/mock/mock_client.go -package=mock

type Client interface {
	NewS3Client() S3Client
	NewOGPFetcher() OGPFetcher
}

type S3Client interface {
	Upload(ctx context.Context, prefix, contentType string, body io.Reader) (string, error)
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
