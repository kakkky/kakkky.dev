package client

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	ogpFetchTimeout = 3 * time.Second
	maxBodySize     = 512 * 1024

	// ref: https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Headers/User-Agent#crawler_and_bot_ua_strings
	userAgent = "Mozilla/5.0 (compatible; kakkky.dev-link-preview/0.1; +https://kakkky.dev)"
)

type OGPFetcher struct {
	client *http.Client
}

func NewOGPFetcher() *OGPFetcher {
	// Dialer.Control は TCP接続の 直前に呼ばれる hook
	// SSRF対策として、IP アドレスを見て内部リソース (loopback / private / link-local) への接続を弾く。
	dialer := &net.Dialer{
		Control: func(_, address string, _ syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}
			ip := net.ParseIP(host)
			if ip == nil ||
				ip.IsLoopback() ||
				ip.IsPrivate() ||
				ip.IsLinkLocalUnicast() ||
				ip.IsLinkLocalMulticast() ||
				ip.IsUnspecified() {
				return errors.New("blocked ip") // 特にエラーを判別しない上、ドメインロジックに関係ない点も考慮してdomain errは返さない
			}
			return nil
		},
	}

	// DefaultTransport を継承しつつ、DialContext / pool / timeout を用途向けに調整。
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = dialer.DialContext

	// OGP fetch は多様な host にバラつくため connection pool の再利用機会が少ない。
	// pool を小さく保って idle fd を節約。
	transport.MaxIdleConns = 10
	transport.IdleConnTimeout = 30 * time.Second

	// 全体 Client.Timeout 3s に対して、各 phase を fail-fast させる。
	transport.TLSHandshakeTimeout = 2 * time.Second
	transport.ResponseHeaderTimeout = 2 * time.Second

	return &OGPFetcher{
		client: &http.Client{
			Timeout:   ogpFetchTimeout,
			Transport: transport,
			// URLがリダイレクトを何段階も挟むようになっていた場合はこちらのリソースを食われてしまうので
			// リダイレクト段数を制限する
			CheckRedirect: func(_ *http.Request, via []*http.Request) error {
				if len(via) >= 3 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
	}
}

type OGPData struct {
	Host        string
	Title       string
	Description string
	Image       string
}

func (f *OGPFetcher) Fetch(ctx context.Context, rawURL string) (OGPData, error) {
	u, err := url.Parse(rawURL)

	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return OGPData{}, errors.New("invalid url")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return OGPData{}, err
	}

	req.Header.Set("User-Agent", userAgent)
	// Open Graph Protocol 仕様はHTMLを前提としている
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := f.client.Do(req)
	if err != nil {
		return OGPData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return OGPData{}, errors.New("bad status")
	}

	// 巨大 HTML による OOM 攻撃防止。<head>は先頭にあるため切り詰めても問題ない。
	doc, err := goquery.NewDocumentFromReader(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return OGPData{}, err
	}

	data := extractOGP(doc)
	data.Host = u.Host

	return data, nil
}

func extractOGP(doc *goquery.Document) OGPData {
	var d OGPData

	d.Title = doc.Find(`meta[property="og:title"]`).AttrOr("content", "")
	if d.Title == "" {
		d.Title = strings.TrimSpace(doc.Find("title").First().Text())
	}

	d.Description = doc.Find(`meta[property="og:description"]`).AttrOr("content", "")
	if d.Description == "" {
		d.Description = doc.Find(`meta[name="description"]`).AttrOr("content", "")
	}

	d.Image = doc.Find(`meta[property="og:image"]`).AttrOr("content", "")

	return d
}
