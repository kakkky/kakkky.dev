package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/kakkky/hotwire-go/turbo"
	"github.com/kakkky/kakkky.dev/adapter/view/components"
	"github.com/kakkky/kakkky.dev/adapter/view/pages"
	"github.com/kakkky/kakkky.dev/adapter/view/partials"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/usecase"
)

const (
	feedDefaultLimit = 10
)

type FeedHandler struct {
	getFeedUsecase *usecase.GetFeedUsecase
}

func NewFeedHandler(getFeedUsecase *usecase.GetFeedUsecase) *FeedHandler {
	return &FeedHandler{
		getFeedUsecase: getFeedUsecase,
	}
}

func (h *FeedHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()

	limit := feedDefaultLimit
	if limitStr := q.Get("limit"); limitStr != "" {
		limitInt, err := strconv.Atoi(limitStr)
		if err != nil || limitInt <= 0 {
			http.Error(rw, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = limitInt
	}

	input := usecase.GetFeedUsecaseInput{Limit: limit}
	cursorID := q.Get("cursor_id")
	cursorAt := q.Get("cursor_at")
	switch {
	case cursorID != "" && cursorAt != "":
		t, err := time.Parse(time.RFC3339, cursorAt)
		if err != nil {
			http.Error(rw, "invalid cursor_at", http.StatusBadRequest)
			return
		}
		input.Cursor = usecase.GetFeedUsecaseCursor{
			AfterID:          domain.FeedItemID(cursorID),
			AfterPublishedAt: t,
		}
	case cursorID != "" || cursorAt != "":
		http.Error(rw, "cursor_id と cursor_at は同時に指定してください", http.StatusBadRequest)
		return
	}

	out, err := h.getFeedUsecase.Exec(ctx, input)
	if err != nil {
		writeFeedError(rw, err)
		return
	}

	items := make([]components.FeedItemCardViewModel, 0, len(out.Items))
	for _, it := range out.Items {
		tagNames := make([]string, 0, len(it.TagIDs))
		for _, tid := range it.TagIDs {
			if t, ok := out.Tags[tid]; ok {
				tagNames = append(tagNames, t.Name)
			}
		}
		items = append(items, components.FeedItemCardViewModel{
			IsSeries:     it.Kind == domain.FeedItemKindSeries,
			Title:        it.Title,
			Href:         feedItemHref(it.Kind, it.Slug),
			PublishedAt:  it.PublishedAt,
			Tags:         tagNames,
			ArticleCount: it.ArticleCount,
			SeriesStatus: string(it.SeriesStatus),
		})
	}
	nextID := string(out.NextCursor.AfterID)
	nextAt := out.NextCursor.AfterPublishedAt
	if turbo.IsFrameRequest(r) {
		_ = partials.FeedItems(partials.FeedItemsViewModel{
			Items:        items,
			NextCursorID: nextID,
			NextCursorAt: nextAt,
		}).Render(ctx, rw)
		return
	}
	_ = pages.Feed(pages.FeedViewModel{
		Items:        items,
		NextCursorID: nextID,
		NextCursorAt: nextAt,
	}).Render(ctx, rw)
}

func feedItemHref(kind domain.FeedItemKind, slug domain.Slug) string {
	if kind == domain.FeedItemKindSeries {
		return "/series/" + string(slug)
	}
	return "/articles/" + string(slug)
}

// TODO: エラー画面とかその辺は後で共通化ちゃんとやる。一旦適当。
func writeFeedError(rw http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidArgument):
		http.Error(rw, "invalid argument", http.StatusBadRequest)
	case errors.Is(err, domain.ErrNotFound):
		http.Error(rw, "not found", http.StatusNotFound)
	default:
		http.Error(rw, "internal error", http.StatusInternalServerError)
	}
}
