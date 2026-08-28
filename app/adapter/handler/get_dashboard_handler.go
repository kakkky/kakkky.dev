package handler

import (
	"net/http"
	"net/url"
	"time"

	"github.com/kakkky/hotwire-go/turbo"

	"github.com/kakkky/kakkky.dev/adapter/view/components"
	"github.com/kakkky/kakkky.dev/adapter/view/pages"
	"github.com/kakkky/kakkky.dev/adapter/view/partials"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/usecase"
)

const (
	dashboardArticlesLimit = 10
	dashboardSeriesLimit   = 10
)

type GetDashboardHandler struct {
	listArticlesUsecase *usecase.ListArticlesUsecase
	listSeriesUsecase   *usecase.ListSeriesUsecase
}

func NewGetDashboardHandler(
	listArticlesUsecase *usecase.ListArticlesUsecase,
	listSeriesUsecase *usecase.ListSeriesUsecase,
) *GetDashboardHandler {
	return &GetDashboardHandler{
		listArticlesUsecase: listArticlesUsecase,
		listSeriesUsecase:   listSeriesUsecase,
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
		RenderError(rw, r, err)
		return
	}
	sOut, err := h.listSeriesUsecase.Exec(ctx, usecase.ListSeriesUsecaseInput{
		Limit: dashboardSeriesLimit,
	})
	if err != nil {
		RenderError(rw, r, err)
		return
	}
	vm := pages.DashboardViewModel{
		Articles: buildArticlesListViewModel(aOut.Articles, aOut.NextCursor),
		Series:   buildSeriesListViewModel(sOut.Series, sOut.NextCursor),
	}
	_ = pages.Dashboard(vm).Render(ctx, rw)
}

func (h *GetDashboardHandler) renderArticlesPartial(rw http.ResponseWriter, r *http.Request, q url.Values) {
	ctx := r.Context()
	cursor, err := parseArticlesCursor(q)
	if err != nil {
		RenderError(rw, r, err)
		return
	}
	out, err := h.listArticlesUsecase.Exec(ctx, usecase.ListArticlesUsecaseInput{
		Cursor: cursor,
		Limit:  dashboardArticlesLimit,
	})
	if err != nil {
		RenderError(rw, r, err)
		return
	}
	vm := buildArticlesListViewModel(out.Articles, out.NextCursor)
	_ = partials.ArticlesList(vm).Render(ctx, rw)
}

func (h *GetDashboardHandler) renderSeriesPartial(rw http.ResponseWriter, r *http.Request, q url.Values) {
	ctx := r.Context()
	cursor, err := parseSeriesCursor(q)
	if err != nil {
		RenderError(rw, r, err)
		return
	}
	out, err := h.listSeriesUsecase.Exec(ctx, usecase.ListSeriesUsecaseInput{
		Cursor: cursor,
		Limit:  dashboardSeriesLimit,
	})
	if err != nil {
		RenderError(rw, r, err)
		return
	}
	vm := buildSeriesListViewModel(out.Series, out.NextCursor)
	_ = partials.SeriesList(vm).Render(ctx, rw)
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
	nextCursor usecase.ListArticlesUsecaseCursor,
) partials.ArticlesListViewModel {
	items := make([]components.ArticleRowViewModel, len(articles))
	for i, a := range articles {
		items[i] = components.ArticleRowViewModel{
			Title:     a.Title,
			Status:    string(a.Status),
			CreatedAt: a.CreatedAt,
			Href:      "/articles/" + string(a.Slug),
			EditHref:  "/admin/articles/" + string(a.Slug) + "/edit",
		}
	}
	var nextURL string
	if nextCursor.AfterID != "" {
		v := url.Values{}
		v.Set("panel", "articles")
		v.Set("cursor_id", string(nextCursor.AfterID))
		v.Set("cursor_at", nextCursor.AfterCreatedAt.Format(time.RFC3339Nano))
		nextURL = "/admin/dashboard?" + v.Encode()
	}
	return partials.ArticlesListViewModel{Items: items, NextCursorURL: nextURL}
}

func buildSeriesListViewModel(
	series []domain.Series,
	nextCursor usecase.ListSeriesUsecaseCursor,
) partials.SeriesListViewModel {
	items := make([]components.SeriesRowViewModel, len(series))
	for i, s := range series {
		items[i] = components.SeriesRowViewModel{
			Title:        s.Title,
			Status:       string(s.Status),
			ArticleCount: len(s.Articles),
			CreatedAt:    s.CreatedAt,
			Href:         "/series/" + string(s.Slug),
			EditHref:     "/admin/series/" + string(s.Slug) + "/edit",
		}
	}
	var nextURL string
	if nextCursor.AfterID != "" {
		v := url.Values{}
		v.Set("panel", "series")
		v.Set("cursor_id", string(nextCursor.AfterID))
		v.Set("cursor_at", nextCursor.AfterCreatedAt.Format(time.RFC3339Nano))
		nextURL = "/admin/dashboard?" + v.Encode()
	}
	return partials.SeriesListViewModel{Items: items, NextCursorURL: nextURL}
}
