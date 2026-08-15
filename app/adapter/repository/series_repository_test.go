package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/testhelper"
)

func TestSeriesRepository_FindBySlug(t *testing.T) {
	ctx := t.Context()

	var (
		series1  domain.SeriesID  = "cccccccc-cccc-cccc-cccc-ccccccccccc1"
		series2  domain.SeriesID  = "cccccccc-cccc-cccc-cccc-ccccccccccc2"
		article1 domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1"
		article2 domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2"
		tag1     domain.TagID     = "11111111-1111-1111-1111-111111111111"
		tag2     domain.TagID     = "22222222-2222-2222-2222-222222222222"
	)
	publishedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name             string
		existingTags     []*domain.Tag
		existingArticles []*domain.Article
		existingSeries   []*domain.Series
		slug             domain.Slug
		want             *domain.Series
		wantErr          error
	}{
		{
			name: "returns series with articles ordered by position and tag ids",
			existingTags: []*domain.Tag{
				{ID: tag1, Slug: "go", Name: "Go"},
				{ID: tag2, Slug: "ts", Name: "TypeScript"},
			},
			existingArticles: []*domain.Article{
				{
					ID: article1, Slug: "first", Title: "First", Body: "body1",
					Status: domain.ArticleStatusPublished, PublishedAt: publishedAt,
				},
				{
					ID: article2, Slug: "second", Title: "Second", Body: "body2",
					Status: domain.ArticleStatusPublished, PublishedAt: publishedAt,
				},
			},
			existingSeries: []*domain.Series{
				{
					ID: series1, Slug: "clean-arch", Title: "Clean Arch", Description: "desc",
					Status: domain.SeriesStatusPublishedOngoing, PublishedAt: publishedAt,
					TagIDs: []domain.TagID{tag2, tag1},
					Articles: []domain.SeriesArticle{
						{ArticleID: article2, Position: 2},
						{ArticleID: article1, Position: 1},
					},
				},
			},
			slug: "clean-arch",
			want: &domain.Series{
				ID: series1, Slug: "clean-arch", Title: "Clean Arch", Description: "desc",
				Status: domain.SeriesStatusPublishedOngoing, PublishedAt: publishedAt,
				TagIDs: []domain.TagID{tag1, tag2},
				Articles: []domain.SeriesArticle{
					{ArticleID: article1, Position: 1},
					{ArticleID: article2, Position: 2},
				},
			},
		},
		{
			name: "returns series with empty articles and tags when none linked",
			existingSeries: []*domain.Series{
				{
					ID: series2, Slug: "empty", Title: "Empty",
					Status: domain.SeriesStatusDraft,
				},
			},
			slug: "empty",
			want: &domain.Series{
				ID: series2, Slug: "empty", Title: "Empty",
				Status:   domain.SeriesStatusDraft,
				TagIDs:   []domain.TagID{},
				Articles: []domain.SeriesArticle{},
			},
		},
		{
			name:    "returns not found error when slug does not exist",
			slug:    "missing",
			wantErr: domain.ErrNotFound,
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

			sr := &SeriesRepository{db: testDB}
			got, err := sr.FindBySlug(ctx, tt.slug)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
