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
	createdAt := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 3, 15, 30, 0, 0, time.UTC)

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
					CreatedAt: createdAt, UpdatedAt: updatedAt,
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
				CreatedAt: createdAt, UpdatedAt: updatedAt,
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

func TestArticleRepository_Create(t *testing.T) {
	ctx := t.Context()

	var (
		tag1 domain.TagID = "11111111-1111-1111-1111-111111111111"
		tag2 domain.TagID = "22222222-2222-2222-2222-222222222222"
	)
	publishedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		existingTags []*domain.Tag
		article      *domain.Article
	}{
		{
			name: "creates article with tags",
			existingTags: []*domain.Tag{
				{ID: tag1, Slug: "go", Name: "Go"},
				{ID: tag2, Slug: "ts", Name: "TypeScript"},
			},
			article: &domain.Article{
				Slug: "hello", Title: "Hello", Body: "body",
				Status: domain.ArticleStatusPublished, PublishedAt: publishedAt,
				TagIDs: []domain.TagID{tag1, tag2},
			},
		},
		{
			name: "creates draft article with zero published_at and no tags",
			article: &domain.Article{
				Slug: "draft-slug", Title: "Draft", Body: "body",
				Status: domain.ArticleStatusDraft,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				testhelper.TruncateAll(t, ctx, testDB)
			})

			testhelper.Insert(t, ctx, testDB, testhelper.Fixtures{Tags: tt.existingTags})

			ar := &ArticleRepository{db: testDB}
			require.NoError(t, ar.Create(ctx, tt.article))
			require.NotEmpty(t, tt.article.ID)
			require.False(t, tt.article.CreatedAt.IsZero())

			got, err := ar.FindBySlug(ctx, tt.article.Slug)
			require.NoError(t, err)
			assert.Equal(t, tt.article.Slug, got.Slug)
			assert.Equal(t, tt.article.Title, got.Title)
			assert.Equal(t, tt.article.Body, got.Body)
			assert.Equal(t, tt.article.Status, got.Status)
			assert.Equal(t, tt.article.PublishedAt.UTC(), got.PublishedAt)
			assert.ElementsMatch(t, tt.article.TagIDs, got.TagIDs)
		})
	}
}

func TestArticleRepository_Update(t *testing.T) {
	ctx := t.Context()

	var (
		articleID domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1"
		tag1      domain.TagID     = "11111111-1111-1111-1111-111111111111"
		tag2      domain.TagID     = "22222222-2222-2222-2222-222222222222"
		tag3      domain.TagID     = "33333333-3333-3333-3333-333333333333"
	)
	publishedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	t.Run("updates fields and replaces tags", func(t *testing.T) {
		t.Cleanup(func() {
			testhelper.TruncateAll(t, ctx, testDB)
		})

		testhelper.Insert(t, ctx, testDB, testhelper.Fixtures{
			Tags: []*domain.Tag{
				{ID: tag1, Slug: "go", Name: "Go"},
				{ID: tag2, Slug: "ts", Name: "TypeScript"},
				{ID: tag3, Slug: "db", Name: "DB"},
			},
			Articles: []*domain.Article{
				{
					ID: articleID, Slug: "old", Title: "Old", Body: "old-body",
					Status: domain.ArticleStatusDraft,
					TagIDs: []domain.TagID{tag1, tag2},
				},
			},
		})

		ar := &ArticleRepository{db: testDB}
		err := ar.Update(ctx, &domain.Article{
			ID: articleID, Slug: "new", Title: "New", Body: "new-body",
			Status: domain.ArticleStatusPublished, PublishedAt: publishedAt,
			TagIDs: []domain.TagID{tag3},
		})
		require.NoError(t, err)

		got, err := ar.FindBySlug(ctx, "new")
		require.NoError(t, err)
		assert.Equal(t, articleID, got.ID)
		assert.Equal(t, "New", got.Title)
		assert.Equal(t, "new-body", got.Body)
		assert.Equal(t, domain.ArticleStatusPublished, got.Status)
		assert.Equal(t, publishedAt, got.PublishedAt)
		assert.Equal(t, []domain.TagID{tag3}, got.TagIDs)
	})

	t.Run("returns not found when article does not exist", func(t *testing.T) {
		t.Cleanup(func() {
			testhelper.TruncateAll(t, ctx, testDB)
		})

		ar := &ArticleRepository{db: testDB}
		err := ar.Update(ctx, &domain.Article{
			ID: articleID, Slug: "s", Title: "T", Body: "B",
			Status: domain.ArticleStatusDraft,
		})
		require.ErrorIs(t, err, domain.ErrNotFound)
	})
}

func TestArticleRepository_List(t *testing.T) {
	ctx := t.Context()

	var (
		article1 domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1"
		article2 domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2"
		article3 domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3"
	)
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// article1: newest, article3: middle, article2: oldest
	fixtures := []*domain.Article{
		{ID: article1, Slug: "a1", Title: "A1", Body: "b1", Status: domain.ArticleStatusPublished, CreatedAt: baseTime.Add(3 * time.Hour)},
		{ID: article2, Slug: "a2", Title: "A2", Body: "b2", Status: domain.ArticleStatusDraft, CreatedAt: baseTime.Add(1 * time.Hour)},
		{ID: article3, Slug: "a3", Title: "A3", Body: "b3", Status: domain.ArticleStatusPublished, CreatedAt: baseTime.Add(2 * time.Hour)},
	}

	tests := []struct {
		name           string
		afterID        domain.ArticleID
		afterCreatedAt time.Time
		limit          int
		wantIDs        []domain.ArticleID
	}{
		{
			name:    "returns all articles ordered by created_at desc (draft included)",
			limit:   10,
			wantIDs: []domain.ArticleID{article1, article3, article2},
		},
		{
			name:    "applies limit",
			limit:   2,
			wantIDs: []domain.ArticleID{article1, article3},
		},
		{
			name:           "cursor advances to next batch",
			afterID:        article1,
			afterCreatedAt: baseTime.Add(3 * time.Hour),
			limit:          10,
			wantIDs:        []domain.ArticleID{article3, article2},
		},
		{
			name:           "cursor at last item returns empty",
			afterID:        article2,
			afterCreatedAt: baseTime.Add(1 * time.Hour),
			limit:          10,
			wantIDs:        []domain.ArticleID{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				testhelper.TruncateAll(t, ctx, testDB)
			})

			testhelper.Insert(t, ctx, testDB, testhelper.Fixtures{Articles: fixtures})

			ar := &ArticleRepository{db: testDB}
			got, err := ar.List(ctx, tt.afterID, tt.afterCreatedAt, tt.limit)
			require.NoError(t, err)

			gotIDs := make([]domain.ArticleID, len(got))
			for i, a := range got {
				gotIDs[i] = a.ID
			}
			assert.Equal(t, tt.wantIDs, gotIDs)
		})
	}
}
