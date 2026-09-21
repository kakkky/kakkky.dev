package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	maxBodyReadBytes     = 1 * 1024 * 1024
	maxLoggedStringChars = 128
)

func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &accessLogWriter{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		body := requestBodyParams(r)
		next.ServeHTTP(rw, r)

		slog.LogAttrs(r.Context(), statusToLevel(rw.status), "http request",
			slog.Group("http",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.status),
				slog.Float64("duration", time.Since(start).Seconds()),
				slog.String("host", r.Host),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
				slog.String("referer", r.Referer()),
			),
			slog.Group("params",
				slog.Any("path", pathParams(r)),
				slog.Any("query", queryParams(r)),
				slog.Any("body", body),
			),
		)
	})
}

type accessLogWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *accessLogWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.status = status
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func statusToLevel(status int) slog.Level {
	switch {
	case status >= 500:
		return slog.LevelError
	case status >= 400:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

func pathParams(r *http.Request) map[string]string {
	if r.Pattern == "" {
		return nil
	}
	idx := strings.Index(r.Pattern, "/")
	if idx < 0 {
		return nil
	}
	params := map[string]string{}
	for seg := range strings.SplitSeq(r.Pattern[idx:], "/") {
		if len(seg) < 2 || seg[0] != '{' || seg[len(seg)-1] != '}' {
			continue
		}
		// ServeMux 記法: {name} は通常の wildcard、{name...} は greedy (残り path 全部)、{$} は末尾 anchor。
		// greedy は "..." を剥がした name で PathValue が引ける。{$} は名前を持たないので除外。
		name := strings.TrimSuffix(seg[1:len(seg)-1], "...")
		if name == "" || name == "$" {
			continue
		}
		params[name] = r.PathValue(name)
	}
	if len(params) == 0 {
		return nil
	}
	return params
}

func queryParams(r *http.Request) url.Values {
	q := r.URL.Query()
	if len(q) == 0 {
		return nil
	}
	return q
}

func requestBodyParams(r *http.Request) any {
	if r.Body == nil || r.Body == http.NoBody {
		return nil
	}
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		return nil
	}
	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/") {
		return "[multipart omitted]"
	}

	buf, err := io.ReadAll(io.LimitReader(r.Body, maxBodyReadBytes+1))
	if err != nil {
		return nil
	}

	r.Body = struct {
		io.Reader
		io.Closer
	}{
		// 読んだ buf を先頭に、上限超過で未読の残りがあれば連結し、下流 handler に元と同じ body を再現する。
		Reader: io.MultiReader(bytes.NewReader(buf), r.Body),
		Closer: r.Body,
	}
	if len(buf) > maxBodyReadBytes {
		return "[body too large]"
	}

	switch {
	case strings.HasPrefix(ct, "application/json"):
		var v any
		if err := json.Unmarshal(buf, &v); err != nil {
			return format(string(buf))
		}
		return format(v)
	case strings.HasPrefix(ct, "application/x-www-form-urlencoded"):
		vals, err := url.ParseQuery(string(buf))
		if err != nil {
			return format(string(buf))
		}
		return format(vals)
	default:
		return format(string(buf))
	}
}

func format(v any) any {
	switch x := v.(type) {
	case string:
		return trimString(x)
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, v := range x {
			out[k] = format(v)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, v := range x {
			out[i] = format(v)
		}
		return out
	case url.Values:
		out := make(url.Values, len(x))
		for k, vs := range x {
			trimmed := make([]string, len(vs))
			for i, s := range vs {
				trimmed[i] = trimString(s)
			}
			out[k] = trimmed
		}
		return out
	default:
		return v
	}
}

func trimString(s string) string {
	n := 0
	for i := range s {
		if n == maxLoggedStringChars {
			return s[:i] + "…"
		}
		n++
	}
	return s
}
