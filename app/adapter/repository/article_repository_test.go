package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/testhelper"
)

func TestArticleRepository_FindBySlug(t *testing.T) {
	ctx := t.Context()

	var (
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
		slug             domain.Slug
		want             *domain.Article
		wantErr          error
	}{
		{
			name: "returns article with tag ids",
			existingTags: []*domain.Tag{
				{ID: tag1, Slug: "go", Name: "Go"},
				{ID: tag2, Slug: "ts", Name: "TypeScript"},
			},
			existingArticles: []*domain.Article{
				{
					ID: article1, Slug: "first", Title: "First", Body: "body1",
					Status: domain.ArticleStatusPublished, PublishedAt: publishedAt,
					TagIDs: []domain.TagID{tag2, tag1},
				},
				{
					ID: article2, Slug: "second", Title: "Second", Body: "body2",
					Status: domain.ArticleStatusDraft,
				},
			},
			slug: "first",
			want: &domain.Article{
				ID: article1, Slug: "first", Title: "First", Body: "body1",
				Status: domain.ArticleStatusPublished, PublishedAt: publishedAt,
				TagIDs: []domain.TagID{tag1, tag2},
			},
		},
		{
			name: "returns not found error when slug does not exist",
			existingArticles: []*domain.Article{
				{
					ID: article1, Slug: "first", Title: "First", Body: "body1",
					Status: domain.ArticleStatusPublished, PublishedAt: publishedAt,
				},
			},
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
			})

			ar := &ArticleRepository{db: testDB}
			got, err := ar.FindBySlug(ctx, tt.slug)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestArticleRepository_FindByIDs(t *testing.T) {
	ctx := t.Context()

	var (
		article1 domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1"
		article2 domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2"
		article3 domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3"
		tag1     domain.TagID     = "11111111-1111-1111-1111-111111111111"
	)
	publishedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name             string
		existingTags     []*domain.Tag
		existingArticles []*domain.Article
		ids              []domain.ArticleID
		wantIDs          []domain.ArticleID
	}{
		{
			name: "returns articles matching given ids",
			existingTags: []*domain.Tag{
				{ID: tag1, Slug: "go", Name: "Go"},
			},
			existingArticles: []*domain.Article{
				{
					ID: article1, Slug: "first", Title: "First", Body: "body1",
					Status: domain.ArticleStatusPublished, PublishedAt: publishedAt,
					TagIDs: []domain.TagID{tag1},
				},
				{
					ID: article2, Slug: "second", Title: "Second", Body: "body2",
					Status: domain.ArticleStatusPublished, PublishedAt: publishedAt,
				},
				{
					ID: article3, Slug: "third", Title: "Third", Body: "body3",
					Status: domain.ArticleStatusDraft,
				},
			},
			ids:     []domain.ArticleID{article1, article3},
			wantIDs: []domain.ArticleID{article1, article3},
		},
		{
			name:    "returns empty slice when ids is empty",
			ids:     nil,
			wantIDs: []domain.ArticleID{},
		},
		{
			name: "returns empty slice when no articles match",
			existingArticles: []*domain.Article{
				{
					ID: article1, Slug: "first", Title: "First", Body: "body1",
					Status: domain.ArticleStatusPublished, PublishedAt: publishedAt,
				},
			},
			ids:     []domain.ArticleID{article2},
			wantIDs: []domain.ArticleID{},
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
			})

			ar := &ArticleRepository{db: testDB}
			got, err := ar.FindByIDs(ctx, tt.ids...)
			require.NoError(t, err)

			gotIDs := make([]domain.ArticleID, len(got))
			for i, a := range got {
				gotIDs[i] = a.ID
			}
			assert.ElementsMatch(t, tt.wantIDs, gotIDs)
		})
	}
}
