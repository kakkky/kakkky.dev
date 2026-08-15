package handler

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/kakkky/kakkky.dev/adapter/view/components"
)

const (
	previewFetchTimeout = 3 * time.Second
	previewMaxBodySize  = 512 * 1024
	previewUserAgent    = "kakkky.dev-preview-bot/0.1"
)

type PreviewHandler struct {
	client *http.Client
}

func NewPreviewHandler() *PreviewHandler {
	return &PreviewHandler{
		client: &http.Client{
			Timeout: previewFetchTimeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 3 {
					return http.ErrUseLastResponse
				}
				if err := checkSafeURL(req.URL); err != nil {
					return err
				}
				return nil
			},
		},
	}
}

func (h *PreviewHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	raw := r.URL.Query().Get("url")

	frameID := PreviewFrameID(raw)
	vm := components.LinkPreviewViewModel{URL: raw}

	u, err := url.Parse(raw)
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") && checkSafeURL(u) == nil {
		vm.Host = u.Host
		if meta, err := h.fetch(ctx, raw); err == nil {
			if meta.Title != "" {
				vm.Title = meta.Title
			}
			vm.Description = meta.Description
			vm.Image = meta.Image
		}
	}

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = components.LinkPreviewFrame(frameID, vm).Render(ctx, rw)
}

type ogpMeta struct {
	Title       string
	Description string
	Image       string
}

func (h *PreviewHandler) fetch(ctx context.Context, rawURL string) (ogpMeta, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return ogpMeta{}, err
	}
	req.Header.Set("User-Agent", previewUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	resp, err := h.client.Do(req)
	if err != nil {
		return ogpMeta{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return ogpMeta{}, http.ErrNotSupported
	}

	body := io.LimitReader(resp.Body, previewMaxBodySize)
	doc, err := html.Parse(body)
	if err != nil {
		return ogpMeta{}, err
	}
	return extractOGP(doc), nil
}

func extractOGP(n *html.Node) ogpMeta {
	var meta ogpMeta
	var title string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
					title = strings.TrimSpace(n.FirstChild.Data)
				}
			case "meta":
				var prop, name, content string
				for _, a := range n.Attr {
					switch strings.ToLower(a.Key) {
					case "property":
						prop = strings.ToLower(a.Val)
					case "name":
						name = strings.ToLower(a.Val)
					case "content":
						content = a.Val
					}
				}
				key := prop
				if key == "" {
					key = name
				}
				switch key {
				case "og:title":
					meta.Title = content
				case "og:description", "description":
					if meta.Description == "" {
						meta.Description = content
					}
				case "og:image":
					meta.Image = content
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	if meta.Title == "" {
		meta.Title = title
	}
	return meta
}

func checkSafeURL(u *url.URL) error {
	host := u.Hostname()
	if host == "" {
		return http.ErrNoLocation
	}
	if strings.EqualFold(host, "localhost") {
		return http.ErrNotSupported
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return err
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return http.ErrNotSupported
		}
	}
	return nil
}

// PreviewFrameID returns a stable turbo-frame id for a URL.
func PreviewFrameID(rawURL string) string {
	sum := sha1.Sum([]byte(rawURL))
	return "link-preview-" + hex.EncodeToString(sum[:])[:12]
}
