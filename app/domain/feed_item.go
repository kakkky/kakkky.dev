package domain

import "time"

// FeedItemID　は ArticleID または SeriesID を表す型です。
// どちらの ID であるかは FeedItemKind で判別します。
type FeedItemID string

type FeedItemKind string

const (
	FeedItemKindArticle FeedItemKind = "article"
	FeedItemKindSeries  FeedItemKind = "series"
)

type FeedItem struct {
	Kind        FeedItemKind
	ID          FeedItemID
	Slug        Slug
	Title       string
	PublishedAt time.Time
	TagIDs      []TagID

	// 以下のフィールドは FeedItemKind が FeedItemKindSeries の場合のみ有効
	ArticleCount int
	SeriesStatus SeriesStatus
}
