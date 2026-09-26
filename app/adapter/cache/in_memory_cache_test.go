package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kakkky/kakkky.dev/domain"
)

func TestInMemoryCache(t *testing.T) {
	t.Parallel()
	const key = "k"
	now := time.Now()

	tests := []struct {
		name           string
		existing       func(c domain.CacheClient)
		wantGetValue   any
		wantGetOk      bool
		wantStaleValue any
		wantStaleOk    bool
	}{
		{
			name:     "success: missing key -- both Get and GetStale return zero and false",
			existing: func(c domain.CacheClient) {},
		},
		{
			name: "success: fresh entry -- both Get and GetStale return the value",
			existing: func(c domain.CacheClient) {
				c.Set(key, "hello", now.Add(time.Hour))
			},
			wantGetValue:   "hello",
			wantGetOk:      true,
			wantStaleValue: "hello",
			wantStaleOk:    true,
		},
		{
			name: "success: expired entry -- Get misses but GetStale returns the value",
			existing: func(c domain.CacheClient) {
				c.Set(key, "stale", now.Add(-time.Second))
			},
			wantStaleValue: "stale",
			wantStaleOk:    true,
		},
		{
			name: "success: second Set overwrites both value and expiration",
			existing: func(c domain.CacheClient) {
				c.Set(key, "old", now.Add(time.Hour))
				c.Set(key, "new", now.Add(2*time.Hour))
			},
			wantGetValue:   "new",
			wantGetOk:      true,
			wantStaleValue: "new",
			wantStaleOk:    true,
		},
		{
			name: "success: Set on an expired entry refreshes it",
			existing: func(c domain.CacheClient) {
				c.Set(key, "v", now.Add(-time.Second))
				c.Set(key, "v", now.Add(time.Hour))
			},
			wantGetValue:   "v",
			wantGetOk:      true,
			wantStaleValue: "v",
			wantStaleOk:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := NewCache().NewInMemoryCacheClient()
			tt.existing(c)

			gotG, okG := c.Get(key)
			assert.Equal(t, tt.wantGetOk, okG, "Get ok")
			assert.Equal(t, tt.wantGetValue, gotG, "Get value")

			gotS, okS := c.GetStale(key)
			assert.Equal(t, tt.wantStaleOk, okS, "GetStale ok")
			assert.Equal(t, tt.wantStaleValue, gotS, "GetStale value")
		})
	}
}
