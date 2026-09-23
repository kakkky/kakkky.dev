package handler

import (
	"net/http"
	"net/url"
	"time"

	"github.com/kakkky/hotwire-go/turbo"

	"github.com/kakkky/kakkky.dev/adapter/view"
	"github.com/kakkky/kakkky.dev/adapter/view/components"
	"github.com/kakkky/kakkky.dev/adapter/view/pages"
	"github.com/kakkky/kakkky.dev/adapter/view/partials"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/errors"
	"github.com/kakkky/kakkky.dev/usecase"
)

const (
	dashboardArticlesLimit          = 10
	dashboardSeriesLimit            = 10
	dashboardArticlesAnalyticsLimit = 5
)

type GetDashboardHandler struct {
	listArticlesUsecase         *usecase.ListArticlesUsecase
	listSeriesUsecase           *usecase.ListSeriesUsecase
	getSiteAnalyticsUsecase     *usecase.GetSiteAnalyticsUsecase
	getArticlesAnalyticsUsecase *usecase.GetArticlesAnalyticsUsecase
	publicBaseURL               string
}

func NewGetDashboardHandler(
	listArticlesUsecase *usecase.ListArticlesUsecase,
	listSeriesUsecase *usecase.ListSeriesUsecase,
	getSiteAnalyticsUsecase *usecase.GetSiteAnalyticsUsecase,
	getArticlesAnalyticsUsecase *usecase.GetArticlesAnalyticsUsecase,
	publicBaseURL string,
) *GetDashboardHandler {
	return &GetDashboardHandler{
		listArticlesUsecase:         listArticlesUsecase,
		listSeriesUsecase:           listSeriesUsecase,
		getSiteAnalyticsUsecase:     getSiteAnalyticsUsecase,
		getArticlesAnalyticsUsecase: getArticlesAnalyticsUsecase,
		publicBaseURL:               publicBaseURL,
	}
}

func (h *GetDashboardHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	if turbo.IsFrameRequest(r) {
		switch q.Get("panel") {
		case "articles":
			h.renderArticlesPartial(rw, r, q)
			return
		case "series":
			h.renderSeriesPartial(rw, r, q)
			return
		case "site_analytics":
			h.renderSiteAnalyticsPartial(rw, r, q)
			return
		case "articles_analytics":
			h.renderArticlesAnalyticsPartial(rw, r, q)
			return
		}
	}

	h.renderFullPage(rw, r)
}

func (h *GetDashboardHandler) renderFullPage(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	aOut, err := h.listArticlesUsecase.Exec(ctx, usecase.ListArticlesUsecaseInput{
		Limit: dashboardArticlesLimit,
	})
	if err != nil {
		errors.Set(r.Context(), err)
		return
	}
	sOut, err := h.listSeriesUsecase.Exec(ctx, usecase.ListSeriesUsecaseInput{
		Limit: dashboardSeriesLimit,
	})
	if err != nil {
		errors.Set(r.Context(), err)
		return
	}
	vm := pages.DashboardViewModel{
		Articles:      buildArticlesListViewModel(aOut.Articles, aOut.SeriesByID, aOut.NextCursor, h.publicBaseURL),
		ArticlesTotal: aOut.Total,
		Series:        buildSeriesListViewModel(sOut.Series, sOut.NextCursor, h.publicBaseURL),
		SeriesTotal:   sOut.Total,
	}
	_ = pages.Dashboard(vm).Render(ctx, rw)
}

func (h *GetDashboardHandler) renderArticlesPartial(rw http.ResponseWriter, r *http.Request, q url.Values) {
	ctx := r.Context()
	cursor, err := parseArticlesCursor(q)
	if err != nil {
		errors.Set(r.Context(), err)
		return
	}
	out, err := h.listArticlesUsecase.Exec(ctx, usecase.ListArticlesUsecaseInput{
		Cursor: cursor,
		Limit:  dashboardArticlesLimit,
	})
	if err != nil {
		errors.Set(r.Context(), err)
		return
	}
	vm := buildArticlesListViewModel(out.Articles, out.SeriesByID, out.NextCursor, h.publicBaseURL)
	_ = partials.ArticlesList(vm).Render(ctx, rw)
}

func (h *GetDashboardHandler) renderSeriesPartial(rw http.ResponseWriter, r *http.Request, q url.Values) {
	ctx := r.Context()
	cursor, err := parseSeriesCursor(q)
	if err != nil {
		errors.Set(r.Context(), err)
		return
	}
	out, err := h.listSeriesUsecase.Exec(ctx, usecase.ListSeriesUsecaseInput{
		Cursor: cursor,
		Limit:  dashboardSeriesLimit,
	})
	if err != nil {
		errors.Set(r.Context(), err)
		return
	}
	vm := buildSeriesListViewModel(out.Series, out.NextCursor, h.publicBaseURL)
	_ = partials.SeriesList(vm).Render(ctx, rw)
}

func (h *GetDashboardHandler) renderSiteAnalyticsPartial(rw http.ResponseWriter, r *http.Request, q url.Values) {
	ctx := r.Context()
	rangeKey := q.Get("range")
	out, err := h.getSiteAnalyticsUsecase.Exec(ctx, usecase.GetSiteAnalyticsUsecaseInput{Range: rangeKeyToDateRange(rangeKey)})
	if err != nil {
		errors.Set(r.Context(), err)
		return
	}
	vm := buildSiteAnalyticsViewModel(out.Metrics, rangeKey)
	_ = partials.SiteAnalytics(vm).Render(ctx, rw)
}

func (h *GetDashboardHandler) renderArticlesAnalyticsPartial(rw http.ResponseWriter, r *http.Request, q url.Values) {
	ctx := r.Context()
	rangeKey := q.Get("range")
	dateRange := rangeKeyToDateRange(rangeKey)
	out, err := h.getArticlesAnalyticsUsecase.Exec(ctx, usecase.GetArticlesAnalyticsUsecaseInput{Range: dateRange})
	if err != nil {
		errors.Set(r.Context(), err)
		return
	}
	items := out.Items
	if len(items) > dashboardArticlesAnalyticsLimit {
		items = items[:dashboardArticlesAnalyticsLimit]
	}
	vm := buildArticlesAnalyticsViewModel(items, rangeKey, h.publicBaseURL, dateRange)
	_ = partials.ArticlesAnalytics(vm).Render(ctx, rw)
}

func parseArticlesCursor(q url.Values) (usecase.ListArticlesUsecaseCursor, error) {
	cid, cat, err := parseCursorParams(q)
	if err != nil || cid == "" {
		return usecase.ListArticlesUsecaseCursor{}, err
	}
	return usecase.ListArticlesUsecaseCursor{
		AfterID:        domain.ArticleID(cid),
		AfterCreatedAt: cat,
	}, nil
}

func parseSeriesCursor(q url.Values) (usecase.ListSeriesUsecaseCursor, error) {
	cid, cat, err := parseCursorParams(q)
	if err != nil || cid == "" {
		return usecase.ListSeriesUsecaseCursor{}, err
	}
	return usecase.ListSeriesUsecaseCursor{
		AfterID:        domain.SeriesID(cid),
		AfterCreatedAt: cat,
	}, nil
}

func parseCursorParams(q url.Values) (cursorID string, cursorAt time.Time, err error) {
	cid, cat := q.Get("cursor_id"), q.Get("cursor_at")
	switch {
	case cid == "" && cat == "":
		return
	case cid == "" || cat == "":
		err = domain.ErrInvalidArgument.With("cursor_id と cursor_at は同時に指定してください")
		return
	}
	cursorAt, err = time.Parse(time.RFC3339, cat)
	if err != nil {
		err = domain.ErrInvalidArgument.With("cursor_at は RFC3339 形式で指定してください")
		return
	}
	cursorID = cid
	return
}

func buildArticlesListViewModel(
	articles []domain.Article,
	seriesByID map[domain.ArticleID]*domain.Series,
	nextCursor usecase.ListArticlesUsecaseCursor,
	publicBaseURL string,
) partials.ArticlesListViewModel {
	items := make([]components.ArticleRowViewModel, len(articles))
	for i, a := range articles {
		row := components.ArticleRowViewModel{
			Title:     a.Title,
			Status:    string(a.Status),
			CreatedAt: a.CreatedAt,
			Href:      publicBaseURL + "/articles/" + string(a.Slug),
			EditHref:  "/articles/" + string(a.Slug) + "/edit",
		}
		if s, ok := seriesByID[a.ID]; ok && s != nil {
			row.SeriesTitle = s.Title
			row.SeriesHref = publicBaseURL + "/series/" + string(s.Slug)
		}
		items[i] = row
	}
	var nextURL string
	if nextCursor.AfterID != "" {
		v := url.Values{}
		v.Set("panel", "articles")
		v.Set("cursor_id", string(nextCursor.AfterID))
		v.Set("cursor_at", nextCursor.AfterCreatedAt.Format(time.RFC3339Nano))
		nextURL = "/dashboard?" + v.Encode()
	}
	return partials.ArticlesListViewModel{Items: items, NextCursorURL: nextURL}
}

func buildSeriesListViewModel(
	series []domain.Series,
	nextCursor usecase.ListSeriesUsecaseCursor,
	publicBaseURL string,
) partials.SeriesListViewModel {
	items := make([]components.SeriesRowViewModel, len(series))
	for i, s := range series {
		items[i] = components.SeriesRowViewModel{
			Title:        s.Title,
			Status:       string(s.Status),
			ArticleCount: len(s.Articles),
			CreatedAt:    s.CreatedAt,
			Href:         publicBaseURL + "/series/" + string(s.Slug),
			EditHref:     "/series/" + string(s.Slug) + "/edit",
		}
	}
	var nextURL string
	if nextCursor.AfterID != "" {
		v := url.Values{}
		v.Set("panel", "series")
		v.Set("cursor_id", string(nextCursor.AfterID))
		v.Set("cursor_at", nextCursor.AfterCreatedAt.Format(time.RFC3339Nano))
		nextURL = "/dashboard?" + v.Encode()
	}
	return partials.SeriesListViewModel{Items: items, NextCursorURL: nextURL}
}

func buildSiteAnalyticsViewModel(m domain.AnalyticsSiteMetrics, rangeKey string) partials.SiteAnalyticsViewModel {
	refs := make([]partials.AnalyticsReferrerRow, len(m.TopReferrers))
	for i, ref := range m.TopReferrers {
		refs[i] = partials.AnalyticsReferrerRow{
			Source:   ref.Source,
			Medium:   ref.Medium,
			Sessions: ref.Sessions,
		}
	}
	return partials.SiteAnalyticsViewModel{
		RangeKey:       rangeKey,
		TotalUsers:     m.TotalUsers,
		TotalPageViews: m.TotalPageViews,
		LineChartHTML:  view.RenderDailyLineChart(m.ByDate),
		Referrers:      refs,
	}
}

func buildArticlesAnalyticsViewModel(items []usecase.ArticleAnalyticsItem, rangeKey string, publicBaseURL string, dateRange domain.AnalyticsDateRange) partials.ArticlesAnalyticsViewModel {
	rows := make([]partials.ArticleAnalyticsRow, len(items))
	for i, item := range items {
		rows[i] = partials.ArticleAnalyticsRow{
			Title:         item.Title,
			Href:          publicBaseURL + "/articles/" + string(item.Slug),
			Users:         item.Users,
			PageViews:     item.PageViews,
			SparklineHTML: view.RenderSparkline(fillDailyPoints(item.ByDate, dateRange)),
		}
	}
	return partials.ArticlesAnalyticsViewModel{
		RangeKey: rangeKey,
		Items:    rows,
	}
}

// fillDailyPoints は sparkline 用に指定期間内の全日付を欠損なく揃える。
// GA は traffic のあった日しか返さないので、記事ごとに ByDate の長さや先頭日付が変わる。
// そのままだと sparkline の同じ x 位置に異なる日付が並び、横並びの記事間で日付が揃わない。
func fillDailyPoints(points []domain.AnalyticsDailyPoint, r domain.AnalyticsDateRange) []domain.AnalyticsDailyPoint {
	byDate := make(map[string]domain.AnalyticsDailyPoint, len(points))
	for _, p := range points {
		byDate[p.Date.Format("2006-01-02")] = p
	}
	from := time.Date(r.From.Year(), r.From.Month(), r.From.Day(), 0, 0, 0, 0, r.From.Location())
	to := time.Date(r.To.Year(), r.To.Month(), r.To.Day(), 0, 0, 0, 0, r.To.Location())
	var out []domain.AnalyticsDailyPoint
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		if p, ok := byDate[d.Format("2006-01-02")]; ok {
			out = append(out, p)
			continue
		}
		out = append(out, domain.AnalyticsDailyPoint{Date: d})
	}
	return out
}

func rangeKeyToDateRange(k string) domain.AnalyticsDateRange {
	now := time.Now()
	switch k {
	case components.AnalyticsRange30d:
		return domain.AnalyticsDateRange{From: now.AddDate(0, 0, -30), To: now}
	case components.AnalyticsRange90d:
		return domain.AnalyticsDateRange{From: now.AddDate(0, 0, -90), To: now}
	default:
		return domain.AnalyticsDateRange{From: now.AddDate(0, 0, -7), To: now}
	}
}
