package domain

import (
	"context"
	"io"
)

//go:generate mockgen -source=$GOFILE -destination=../testhelper/mock/mock_client.go -package=mock

type Client interface {
	NewImageUploader() ImageUploader
}

type ImageUploader interface {
	Upload(ctx context.Context, key, contentType string, body io.Reader) (string, error)
}
