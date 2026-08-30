package handler

import (
	"net/http"
	"net/url"
	"slices"
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
	feedDefaultItemLimit = 10
	feedMaxFilterTags    = 10
)

type GetFeedHandler struct {
	getFeedUsecase *usecase.GetFeedUsecase
}

func NewGetFeedHandler(getFeedUsecase *usecase.GetFeedUsecase) *GetFeedHandler {
	return &GetFeedHandler{getFeedUsecase: getFeedUsecase}
}

type feedParams struct {
	limit    int
	tagSlugs []string
	cursor   feedCursor
}

type feedCursor struct {
	ID domain.FeedItemID
	At time.Time
}

func (h *GetFeedHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	params, err := parseFeedParams(r)
	if err != nil {
		RenderError(rw, r, err)
		return
	}

	out, err := h.getFeedUsecase.Exec(ctx, usecase.GetFeedUsecaseInput{
		TagSlugs: toSlugs(params.tagSlugs),
		Cursor:   usecase.GetFeedUsecaseCursor{AfterID: params.cursor.ID, AfterPublishedAt: params.cursor.At},
		Limit:    params.limit,
	})
	if err != nil {
		RenderError(rw, r, err)
		return
	}

	// build FeedItemListViewModel
	var nextCursorURL string
	if out.NextCursor.AfterID != "" {
		v := url.Values{}
		v.Set("cursor_id", string(out.NextCursor.AfterID))
		v.Set("cursor_at", out.NextCursor.AfterPublishedAt.Format(time.RFC3339Nano))
		for _, s := range params.tagSlugs {
			v.Add("tag", s)
		}
		nextCursorURL = "/feed?" + v.Encode()
	}

	tagsByID := make(map[domain.TagID]domain.Tag, len(out.Tags))
	for _, t := range out.Tags {
		tagsByID[t.ID] = t
	}
	feedItemVM := make([]components.FeedItemCardViewModel, 0, len(out.Items))
	for _, it := range out.Items {
		tagNames := make([]string, 0, len(it.TagIDs))
		for _, tid := range it.TagIDs {
			if t, ok := tagsByID[tid]; ok {
				tagNames = append(tagNames, t.Name)
			}
		}

		var href string
		switch it.Kind {
		case domain.FeedItemKindArticle:
			href = "/articles/" + string(it.Slug)

		case domain.FeedItemKindSeries:
			href = "/series/" + string(it.Slug)
		}

		feedItemVM = append(feedItemVM, components.FeedItemCardViewModel{
			IsSeries:     it.Kind == domain.FeedItemKindSeries,
			Title:        it.Title,
			Href:         href,
			PublishedAt:  it.PublishedAt,
			Tags:         tagNames,
			ArticleCount: it.ArticleCount,
			SeriesStatus: string(it.SeriesStatus),
		})
	}
	listVM := partials.FeedItemListViewModel{Items: feedItemVM, NextCursorURL: nextCursorURL}

	switch {
	case turbo.IsFrameRequest(r):
		_ = partials.FeedItemList(listVM).Render(ctx, rw)
	default:
		vm := pages.FeedViewModel{
			List: listVM,
			TagFilter: components.TagFilterAreaViewModel{
				Tags:          toTagViewModels(out.Tags),
				TargetFrameID: partials.FeedItemListID,
			},
		}
		_ = pages.Feed(vm).Render(ctx, rw)
	}
}

func parseFeedParams(r *http.Request) (feedParams, error) {
	q := r.URL.Query()

	params := feedParams{limit: feedDefaultItemLimit}

	// limit
	if limitStr := q.Get("limit"); limitStr != "" {
		limitInt, err := strconv.Atoi(limitStr)
		if err != nil || limitInt <= 0 {
			return feedParams{}, domain.ErrInvalidArgument.With("limit は正の整数で指定してください")
		}
		params.limit = limitInt
	}

	// tag
	if tagSlugs := q["tag"]; len(tagSlugs) > 0 {
		slices.Sort(tagSlugs)
		params.tagSlugs = slices.Compact(tagSlugs)
		if len(params.tagSlugs) > feedMaxFilterTags {
			return feedParams{}, domain.ErrInvalidArgument.With("絞り込める タグ は最大 " + strconv.Itoa(feedMaxFilterTags) + " 個までです")
		}
		params.tagSlugs = tagSlugs
	}

	switch cid, cat := q.Get("cursor_id"), q.Get("cursor_at"); {
	case cid != "" && cat != "":
		t, err := time.Parse(time.RFC3339, cat)
		if err != nil {
			return feedParams{}, domain.ErrInvalidArgument.With("cursor_at は RFC3339 形式で指定してください")
		}
		params.cursor = feedCursor{ID: domain.FeedItemID(cid), At: t}
	case cid != "" || cat != "":
		return feedParams{}, domain.ErrInvalidArgument.With("cursor_id と cursor_at は同時に指定してください")
	}

	return params, nil
}

func toSlugs(raw []string) []domain.Slug {
	if len(raw) == 0 {
		return nil
	}
	slugs := make([]domain.Slug, len(raw))
	for i, s := range raw {
		slugs[i] = domain.Slug(s)
	}
	return slugs
}
