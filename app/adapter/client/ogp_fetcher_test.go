package client

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOGPFetcher_Fetch(t *testing.T) {
	tests := []struct {
		name string
		html string
		want OGPData
	}{
		{
			name: "extracts title, description and image from full og:* metadata",
			html: `<html><head>
				<meta property="og:title" content="OG Title"/>
				<meta property="og:description" content="OG Description"/>
				<meta property="og:image" content="https://example.com/og.png"/>
			</head></html>`,
			want: OGPData{Title: "OG Title", Description: "OG Description", Image: "https://example.com/og.png"},
		},
		{
			name: "prefers og:title over <title> when both are present",
			html: `<html><head>
				<title>Fallback Title</title>
				<meta property="og:title" content="OG Title"/>
			</head></html>`,
			want: OGPData{Title: "OG Title"},
		},
		{
			name: "falls back to <title> when og:title is missing",
			html: `<html><head><title>Fallback Title</title></head></html>`,
			want: OGPData{Title: "Fallback Title"},
		},
		{
			name: "trims surrounding whitespace from <title>",
			html: "<html><head><title>\n  Padded Title  \n</title></head></html>",
			want: OGPData{Title: "Padded Title"},
		},
		{
			name: "prefers og:description over meta[name=description] when both are present",
			html: `<html><head>
				<meta property="og:description" content="OG Description"/>
				<meta name="description" content="Fallback Description"/>
			</head></html>`,
			want: OGPData{Description: "OG Description"},
		},
		{
			name: "falls back to meta[name=description] when og:description is missing",
			html: `<html><head><meta name="description" content="Fallback Description"/></head></html>`,
			want: OGPData{Description: "Fallback Description"},
		},
		{
			name: "returns zero-value fields when no relevant meta or <title> is present",
			html: `<html><head></head><body>hello</body></html>`,
			want: OGPData{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				fmt.Fprint(w, tt.html)
			}))
			t.Cleanup(srv.Close)

			// SSRF ガードなしの fetcher。httptest は 127.0.0.1 で listen するので
			// production の Dialer.Control 付き fetcher だと connect 自体が弾かれる。
			f := &OGPFetcher{client: &http.Client{Timeout: ogpFetchTimeout}}

			got, err := f.Fetch(context.Background(), srv.URL)
			require.NoError(t, err)

			u, _ := url.Parse(srv.URL)
			expect := tt.want
			expect.Host = u.Host
			assert.Equal(t, expect, got)
		})
	}
}

func TestOGPFetcher_Fetch_BlocksLoopback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html></html>`))
	}))
	t.Cleanup(srv.Close)

	f := NewOGPFetcher() // Dialer.Control が 127.0.0.1 を弾くはず
	got, err := f.Fetch(context.Background(), srv.URL)
	require.Error(t, err)
	assert.Equal(t, OGPData{}, got)
}
