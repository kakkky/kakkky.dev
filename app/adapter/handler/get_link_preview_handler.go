package handler

import (
	"net/http"
	"sync"
	"time"

	"github.com/kakkky/hotwire-go/turbo"

	"github.com/kakkky/kakkky.dev/adapter/client"
	"github.com/kakkky/kakkky.dev/adapter/view/components"
	"github.com/kakkky/kakkky.dev/adapter/view/partials"
)

const linkPreviewCacheTTL = 6 * time.Hour

type GetLinkPreviewHandler struct {
	ogpFetcher *client.OGPFetcher
	cache      *linkPreviewCache
}

func NewGetLinkPreviewHandler(ogpFetcher *client.OGPFetcher) *GetLinkPreviewHandler {
	return &GetLinkPreviewHandler{
		ogpFetcher: ogpFetcher,
		cache:      newLinkPreviewCache(),
	}
}

func (h *GetLinkPreviewHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rawURL := r.URL.Query().Get("url")

	frameID := turbo.FrameID(r)
	if frameID == "" {
		// 直接アクセス (curl / browser で /link-preview?url=... 直打ち) 時の fallback
		frameID = "link-preview"
	}

	vm := components.LinkPreviewCardViewModel{URL: rawURL}

	// cache 優先。miss なら fetch して cache に保存。
	d, ok := h.cache.get(rawURL)
	if !ok {
		if fetched, err := h.ogpFetcher.Fetch(ctx, rawURL); err == nil {
			h.cache.set(rawURL, fetched)
			d, ok = fetched, true
		}
	}
	if ok {
		vm.Host = d.Host
		vm.Title = d.Title
		vm.Description = d.Description
		vm.Image = d.Image
	}

	_ = partials.LinkPreviewFrame(frameID, vm).Render(ctx, rw)
}

// link preview の結果は in-memory でキャッシュしておく
type linkPreviewCache struct {
	mu      sync.RWMutex
	entries map[string]linkPreviewCacheEntry
}

type linkPreviewCacheEntry struct {
	data client.OGPData
	exp  time.Time
}

func newLinkPreviewCache() *linkPreviewCache {
	return &linkPreviewCache{entries: make(map[string]linkPreviewCacheEntry)}
}

func (c *linkPreviewCache) get(k string) (client.OGPData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[k]
	if !ok || time.Now().After(e.exp) {
		return client.OGPData{}, false
	}
	return e.data, true
}

func (c *linkPreviewCache) set(k string, d client.OGPData) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[k] = linkPreviewCacheEntry{data: d, exp: time.Now().Add(linkPreviewCacheTTL)}
}
