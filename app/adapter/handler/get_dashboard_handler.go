package handler

import (
	"net/http"

	"github.com/kakkky/kakkky.dev/adapter/view/components"
	"github.com/kakkky/kakkky.dev/adapter/view/pages"
	"github.com/kakkky/kakkky.dev/usecase"
)

type GetDashboardHandler struct {
	getDashboardUsecase *usecase.GetDashboardUsecase
}

func NewGetDashboardHandler(getDashboardUsecase *usecase.GetDashboardUsecase) *GetDashboardHandler {
	return &GetDashboardHandler{getDashboardUsecase: getDashboardUsecase}
}

func (h *GetDashboardHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if _, err := h.getDashboardUsecase.Exec(ctx, usecase.GetDashboardUsecaseInput{}); err != nil {
		RenderError(rw, r, err)
		return
	}

	vm := pages.AdminDashboardViewModel{
		Stats: []components.StatCardViewModel{
			{Label: "Articles", Value: "12", Hint: "総数"},
			{Label: "Series", Value: "3", Hint: "総数"},
			{Label: "PV", Value: "---", Hint: "GA未接続"},
			{Label: "UU", Value: "---", Hint: "GA未接続"},
		},
		TrafficNote: "Chart placeholder — GA未接続",
		TopArticles: []pages.AdminDashboardTopArticleRow{
			{Title: "kakkky.dev で採用した hotwire-go の話", PageViews: "1,234", PageViewsPrev: "1,100", Growth: 12.1, Users: "890", AvgStay: "2:15", TopReferrer: "google / organic"},
			{Title: "SSRF 対策付き OGP fetcher の設計", PageViews: "876", PageViewsPrev: "920", Growth: -4.8, Users: "612", AvgStay: "3:02", TopReferrer: "x.com / referral"},
			{Title: "sqlx で書く thin repository", PageViews: "543", PageViewsPrev: "500", Growth: 8.6, Users: "410", AvgStay: "1:48", TopReferrer: "google / organic"},
			{Title: "Wire を使った DI の勘所", PageViews: "412", PageViewsPrev: "420", Growth: -1.9, Users: "298", AvgStay: "2:33", TopReferrer: "b.hatena.ne.jp / referral"},
			{Title: "Turbo Frame と Stimulus の分担", PageViews: "389", PageViewsPrev: "389", Growth: 0, Users: "310", AvgStay: "2:01", TopReferrer: "google / organic"},
			{Title: "templ の component 設計指針", PageViews: "302", PageViewsPrev: "255", Growth: 18.4, Users: "241", AvgStay: "2:44", TopReferrer: "zenn.dev / referral"},
		},
		Referrers: []pages.AdminDashboardReferrerRow{
			{Source: "google / organic", Sessions: "2,340", Users: "1,890"},
			{Source: "(direct) / (none)", Sessions: "1,102", Users: "923"},
			{Source: "x.com / referral", Sessions: "487", Users: "412"},
			{Source: "b.hatena.ne.jp / referral", Sessions: "234", Users: "198"},
			{Source: "github.com / referral", Sessions: "156", Users: "134"},
			{Source: "zenn.dev / referral", Sessions: "89", Users: "72"},
		},
		Articles: []pages.AdminDashboardArticleRow{
			{Title: "goldmark 拡張の書き方メモ", Status: "draft", UpdatedAt: "2026-08-25", EditHref: "/admin/articles/1/edit"},
			{Title: "kakkky.dev で採用した hotwire-go の話", Status: "published", UpdatedAt: "2026-08-20", EditHref: "/admin/articles/2/edit"},
			{Title: "SSRF 対策付き OGP fetcher の設計", Status: "published", UpdatedAt: "2026-08-15", EditHref: "/admin/articles/3/edit"},
			{Title: "link preview turbo-frame 化", Status: "draft", UpdatedAt: "2026-08-12", EditHref: "/admin/articles/4/edit"},
			{Title: "sqlx で書く thin repository", Status: "published", UpdatedAt: "2026-08-08", EditHref: "/admin/articles/5/edit"},
			{Title: "Turbo Frame と Stimulus の分担", Status: "published", UpdatedAt: "2026-08-03", EditHref: "/admin/articles/6/edit"},
			{Title: "templ の component 設計指針", Status: "draft", UpdatedAt: "2026-07-28", EditHref: "/admin/articles/7/edit"},
			{Title: "Wire を使った DI の勘所", Status: "published", UpdatedAt: "2026-07-20", EditHref: "/admin/articles/8/edit"},
			{Title: "PostgreSQL の JSON 型を業務で使う", Status: "published", UpdatedAt: "2026-07-15", EditHref: "/admin/articles/9/edit"},
			{Title: "個人開発の技術選定ログ", Status: "draft", UpdatedAt: "2026-07-10", EditHref: "/admin/articles/10/edit"},
		},
		Series: []pages.AdminDashboardSeriesRow{
			{Title: "個人ブログを作る", Status: "published_ongoing", ArticleCount: 3, UpdatedAt: "2026-08-24", EditHref: "/admin/series/1/edit"},
			{Title: "Go で書く小さな DDD", Status: "published_completed", ArticleCount: 5, UpdatedAt: "2026-07-30", EditHref: "/admin/series/2/edit"},
			{Title: "Hotwire 入門", Status: "published_ongoing", ArticleCount: 2, UpdatedAt: "2026-07-25", EditHref: "/admin/series/3/edit"},
			{Title: "個人開発の運用記録", Status: "draft", UpdatedAt: "2026-07-01", EditHref: "/admin/series/4/edit"},
			{Title: "小さな OSS を作る", Status: "published_completed", ArticleCount: 4, UpdatedAt: "2026-06-20", EditHref: "/admin/series/5/edit"},
		},
	}
	_ = pages.AdminDashboard(vm).Render(ctx, rw)
}
