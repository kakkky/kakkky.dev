package query

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/testhelper"
)

func TestFeedQueryService_ListFeedItems(t *testing.T) {
	ctx := t.Context()

	var (
		tag1 domain.TagID = "11111111-1111-1111-1111-111111111111"
		tag2 domain.TagID = "22222222-2222-2222-2222-222222222222"

		art1 domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa01"
		art2 domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa02"
		art3 domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa03"
		art4 domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa04"
		art5 domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa05"

		ser1 domain.SeriesID = "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbb01"
	)

	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name             string
		existingTags     []*domain.Tag
		existingArticles []*domain.Article
		existingSeries   []*domain.Series
		afterID          domain.FeedItemID
		afterPublishedAt time.Time
		limit            int
		want             []domain.FeedItem
	}{
		{
			name: "returns articles and series in descending order of published_at",
			existingArticles: []*domain.Article{
				{ID: art1, Slug: "a1", Title: "A1", Body: "b", Status: domain.ArticleStatusPublished, PublishedAt: baseTime.Add(-3 * time.Hour)},
				{ID: art2, Slug: "a2", Title: "A2", Body: "b", Status: domain.ArticleStatusPublished, PublishedAt: baseTime.Add(-1 * time.Hour)},
			},
			existingSeries: []*domain.Series{
				{ID: ser1, Slug: "s1", Title: "S1", Status: domain.SeriesStatusPublishedOngoing, PublishedAt: baseTime.Add(-2 * time.Hour)},
			},
			limit: 10,
			want: []domain.FeedItem{
				{
					Kind:         domain.FeedItemKindArticle,
					ID:           domain.FeedItemID(art2),
					Slug:         "a2",
					Title:        "A2",
					PublishedAt:  baseTime.Add(-1 * time.Hour),
					TagIDs:       []domain.TagID{},
					ArticleCount: 0,
					SeriesStatus: "",
				},
				{
					Kind:         domain.FeedItemKindSeries,
					ID:           domain.FeedItemID(ser1),
					Slug:         "s1",
					Title:        "S1",
					PublishedAt:  baseTime.Add(-2 * time.Hour),
					TagIDs:       []domain.TagID{},
					ArticleCount: 0,
					SeriesStatus: domain.SeriesStatusPublishedOngoing,
				},
				{
					Kind:         domain.FeedItemKindArticle,
					ID:           domain.FeedItemID(art1),
					Slug:         "a1",
					Title:        "A1",
					PublishedAt:  baseTime.Add(-3 * time.Hour),
					TagIDs:       []domain.TagID{},
					ArticleCount: 0,
					SeriesStatus: "",
				},
			},
		},
		{
			name: "excludes articles that belong to a series and reports correct article_count",
			existingArticles: []*domain.Article{
				{ID: art1, Slug: "a1", Title: "A1", Body: "b", Status: domain.ArticleStatusPublished, PublishedAt: baseTime.Add(-3 * time.Hour)},
				{ID: art2, Slug: "a2", Title: "A2", Body: "b", Status: domain.ArticleStatusPublished, PublishedAt: baseTime.Add(-2 * time.Hour)},
				{ID: art3, Slug: "a3", Title: "A3", Body: "b", Status: domain.ArticleStatusPublished, PublishedAt: baseTime.Add(-1 * time.Hour)},
			},
			existingSeries: []*domain.Series{
				{
					ID: ser1, Slug: "s1", Title: "S1", Status: domain.SeriesStatusPublishedCompleted,
					PublishedAt: baseTime.Add(-4 * time.Hour),
					Articles: []domain.SeriesArticle{
						{ArticleID: art1, Position: 1},
						{ArticleID: art2, Position: 2},
					},
				},
			},
			limit: 10,
			want: []domain.FeedItem{
				{
					Kind:         domain.FeedItemKindArticle,
					ID:           domain.FeedItemID(art3),
					Slug:         "a3",
					Title:        "A3",
					PublishedAt:  baseTime.Add(-1 * time.Hour),
					TagIDs:       []domain.TagID{},
					ArticleCount: 0,
					SeriesStatus: "",
				},
				{
					Kind:         domain.FeedItemKindSeries,
					ID:           domain.FeedItemID(ser1),
					Slug:         "s1",
					Title:        "S1",
					PublishedAt:  baseTime.Add(-4 * time.Hour),
					TagIDs:       []domain.TagID{},
					ArticleCount: 2,
					SeriesStatus: domain.SeriesStatusPublishedCompleted,
				},
			},
		},
		{
			name: "returns tag_ids for tagged articles",
			existingTags: []*domain.Tag{
				{ID: tag1, Slug: "go", Name: "Go"},
				{ID: tag2, Slug: "db", Name: "DB"},
			},
			existingArticles: []*domain.Article{
				{
					ID: art1, Slug: "a1", Title: "A1", Body: "b", Status: domain.ArticleStatusPublished,
					PublishedAt: baseTime.Add(-1 * time.Hour),
					TagIDs:      []domain.TagID{tag1, tag2},
				},
			},
			limit: 10,
			want: []domain.FeedItem{
				{
					Kind:         domain.FeedItemKindArticle,
					ID:           domain.FeedItemID(art1),
					Slug:         "a1",
					Title:        "A1",
					PublishedAt:  baseTime.Add(-1 * time.Hour),
					TagIDs:       []domain.TagID{tag1, tag2},
					ArticleCount: 0,
					SeriesStatus: "",
				},
			},
		},
		{
			name: "returns items strictly after the given cursor limited to the given count",
			existingArticles: []*domain.Article{
				{ID: art1, Slug: "a1", Title: "A1", Body: "b", Status: domain.ArticleStatusPublished, PublishedAt: baseTime.Add(-1 * time.Hour)},
				{ID: art2, Slug: "a2", Title: "A2", Body: "b", Status: domain.ArticleStatusPublished, PublishedAt: baseTime.Add(-2 * time.Hour)},
				{ID: art3, Slug: "a3", Title: "A3", Body: "b", Status: domain.ArticleStatusPublished, PublishedAt: baseTime.Add(-4 * time.Hour)},
				{ID: art4, Slug: "a4", Title: "A4 (in series)", Body: "b", Status: domain.ArticleStatusPublished, PublishedAt: baseTime.Add(-6 * time.Hour)},
				{ID: art5, Slug: "a5", Title: "A5 (in series)", Body: "b", Status: domain.ArticleStatusPublished, PublishedAt: baseTime.Add(-7 * time.Hour)},
			},
			existingSeries: []*domain.Series{
				{
					ID: ser1, Slug: "s1", Title: "S1", Status: domain.SeriesStatusPublishedOngoing,
					PublishedAt: baseTime.Add(-3 * time.Hour),
					Articles: []domain.SeriesArticle{
						{ArticleID: art4, Position: 1},
						{ArticleID: art5, Position: 2},
					},
				},
			},
			afterID:          domain.FeedItemID(art1),
			afterPublishedAt: baseTime.Add(-1 * time.Hour),
			limit:            2,
			want: []domain.FeedItem{
				{
					Kind:         domain.FeedItemKindArticle,
					ID:           domain.FeedItemID(art2),
					Slug:         "a2",
					Title:        "A2",
					PublishedAt:  baseTime.Add(-2 * time.Hour),
					TagIDs:       []domain.TagID{},
					ArticleCount: 0,
					SeriesStatus: "",
				},
				{
					Kind:         domain.FeedItemKindSeries,
					ID:           domain.FeedItemID(ser1),
					Slug:         "s1",
					Title:        "S1",
					PublishedAt:  baseTime.Add(-3 * time.Hour),
					TagIDs:       []domain.TagID{},
					ArticleCount: 2,
					SeriesStatus: domain.SeriesStatusPublishedOngoing,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				testhelper.TruncateAll(t, ctx, testDB)
			})

			testhelper.Insert(t, ctx, testDB, testhelper.Fixtures{
				Tags:     tt.existingTags,
				Articles: tt.existingArticles,
				Series:   tt.existingSeries,
			})

			qs := &FeedQueryService{db: testDB}
			got, err := qs.ListFeedItems(ctx, tt.afterID, tt.afterPublishedAt, tt.limit)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
