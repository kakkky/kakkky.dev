package usecase

import (
	"context"
	"sync"
	"time"

	"github.com/kakkky/kakkky.dev/domain"
)

const linkPreviewCacheTTL = 6 * time.Hour

type GetLinkPreviewUsecase struct {
	ogpFetcher domain.OGPFetcher
	cache      *linkPreviewCache
}

func (us *UseCase) NewGetLinkPreviewUsecase() *GetLinkPreviewUsecase {
	return &GetLinkPreviewUsecase{
		ogpFetcher: us.client.NewOGPFetcher(),
		cache:      newLinkPreviewCache(),
	}
}

type GetLinkPreviewUsecaseInput struct {
	URL string
}

type GetLinkPreviewUsecaseOutput struct {
	Data domain.OGPData
}

func (us *GetLinkPreviewUsecase) Exec(ctx context.Context, in GetLinkPreviewUsecaseInput) (GetLinkPreviewUsecaseOutput, error) {
	if d, ok := us.cache.get(in.URL); ok {
		return GetLinkPreviewUsecaseOutput{Data: d}, nil
	}
	d, err := us.ogpFetcher.Fetch(ctx, in.URL)
	if err != nil {
		return GetLinkPreviewUsecaseOutput{}, err
	}
	us.cache.set(in.URL, d)
	return GetLinkPreviewUsecaseOutput{Data: d}, nil
}

type linkPreviewCache struct {
	mu      sync.RWMutex
	entries map[string]linkPreviewCacheEntry
}

type linkPreviewCacheEntry struct {
	data domain.OGPData
	exp  time.Time
}

func newLinkPreviewCache() *linkPreviewCache {
	return &linkPreviewCache{entries: make(map[string]linkPreviewCacheEntry)}
}

func (c *linkPreviewCache) get(k string) (domain.OGPData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[k]
	if !ok || time.Now().After(e.exp) {
		return domain.OGPData{}, false
	}
	return e.data, true
}

func (c *linkPreviewCache) set(k string, d domain.OGPData) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[k] = linkPreviewCacheEntry{data: d, exp: time.Now().Add(linkPreviewCacheTTL)}
}
