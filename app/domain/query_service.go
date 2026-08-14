package domain

import (
	"context"
	"time"
)

//go:generate mockgen -source=$GOFILE -destination=../testhelper/mock/mock_query_service.go -package=mock

type QueryService interface {
	NewFeedQueryService() FeedQueryService
}

type FeedQueryService interface {
	ListFeedItems(ctx context.Context, tagIDs []TagID, afterID FeedItemID, afterPublishedAt time.Time, limit int) ([]FeedItem, error)
}
