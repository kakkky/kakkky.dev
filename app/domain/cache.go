package domain

import "time"

//go:generate mockgen -source=$GOFILE -destination=../testhelper/mock/mock_cache.go -package=mock

type Cache interface {
	NewInMemoryCacheClient() CacheClient
}

type CacheClient interface {
	Get(key string) (any, bool)
	GetStale(key string) (any, bool)
	Set(key string, value any, expiresAt time.Time)
}
