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
	createdAt := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)

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
					Status: domain.SeriesStatusPublishedOngoing, PublishedAt: publishedAt, CreatedAt: createdAt,
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
				Status: domain.SeriesStatusPublishedOngoing, PublishedAt: publishedAt, CreatedAt: createdAt,
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
					Status: domain.SeriesStatusDraft, CreatedAt: createdAt,
				},
			},
			slug: "empty",
			want: &domain.Series{
				ID: series2, Slug: "empty", Title: "Empty",
				Status:    domain.SeriesStatusDraft,
				CreatedAt: createdAt,
				TagIDs:    []domain.TagID{},
				Articles:  []domain.SeriesArticle{},
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

func TestSeriesRepository_List(t *testing.T) {
	ctx := t.Context()

	var (
		series1 domain.SeriesID = "cccccccc-cccc-cccc-cccc-ccccccccccc1"
		series2 domain.SeriesID = "cccccccc-cccc-cccc-cccc-ccccccccccc2"
		series3 domain.SeriesID = "cccccccc-cccc-cccc-cccc-ccccccccccc3"
	)
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// series1: newest, series3: middle, series2: oldest
	fixtures := []*domain.Series{
		{ID: series1, Slug: "s1", Title: "S1", Status: domain.SeriesStatusPublishedOngoing, CreatedAt: baseTime.Add(3 * time.Hour)},
		{ID: series2, Slug: "s2", Title: "S2", Status: domain.SeriesStatusDraft, CreatedAt: baseTime.Add(1 * time.Hour)},
		{ID: series3, Slug: "s3", Title: "S3", Status: domain.SeriesStatusPublishedCompleted, CreatedAt: baseTime.Add(2 * time.Hour)},
	}

	tests := []struct {
		name           string
		afterID        domain.SeriesID
		afterCreatedAt time.Time
		limit          int
		wantIDs        []domain.SeriesID
	}{
		{
			name:    "returns all series ordered by created_at desc (draft included)",
			limit:   10,
			wantIDs: []domain.SeriesID{series1, series3, series2},
		},
		{
			name:    "applies limit",
			limit:   2,
			wantIDs: []domain.SeriesID{series1, series3},
		},
		{
			name:           "cursor advances to next batch",
			afterID:        series1,
			afterCreatedAt: baseTime.Add(3 * time.Hour),
			limit:          10,
			wantIDs:        []domain.SeriesID{series3, series2},
		},
		{
			name:           "cursor at last item returns empty",
			afterID:        series2,
			afterCreatedAt: baseTime.Add(1 * time.Hour),
			limit:          10,
			wantIDs:        []domain.SeriesID{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				testhelper.TruncateAll(t, ctx, testDB)
			})

			testhelper.Insert(t, ctx, testDB, testhelper.Fixtures{Series: fixtures})

			sr := &SeriesRepository{db: testDB}
			got, err := sr.List(ctx, tt.afterID, tt.afterCreatedAt, tt.limit)
			require.NoError(t, err)

			gotIDs := make([]domain.SeriesID, len(got))
			for i, s := range got {
				gotIDs[i] = s.ID
			}
			assert.Equal(t, tt.wantIDs, gotIDs)
		})
	}
}
