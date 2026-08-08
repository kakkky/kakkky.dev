package domain

import (
	"context"
	"time"
)

type QueryService interface {
	NewFeedQueryService() FeedQueryService
}

type FeedQueryService interface {
	ListFeedItems(ctx context.Context, afterID FeedItemID, afterPublishedAt time.Time, limit int) ([]FeedItem, error)
}
