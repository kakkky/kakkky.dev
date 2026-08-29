package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/testhelper"
)

func TestTagRepository_ListAll(t *testing.T) {
	ctx := t.Context()

	var (
		tag1 domain.TagID = "11111111-1111-1111-1111-111111111111"
		tag2 domain.TagID = "22222222-2222-2222-2222-222222222222"
		tag3 domain.TagID = "33333333-3333-3333-3333-333333333333"
	)

	tests := []struct {
		name         string
		existingTags []*domain.Tag
		want         []*domain.Tag
	}{
		{
			name: "returns all tags ordered by name",
			existingTags: []*domain.Tag{
				{ID: tag2, Slug: "ts", Name: "TypeScript"},
				{ID: tag1, Slug: "go", Name: "Go"},
				{ID: tag3, Slug: "db", Name: "DB"},
			},
			want: []*domain.Tag{
				{ID: tag3, Slug: "db", Name: "DB"},
				{ID: tag1, Slug: "go", Name: "Go"},
				{ID: tag2, Slug: "ts", Name: "TypeScript"},
			},
		},
		{
			name:         "returns empty slice when no tags exist",
			existingTags: nil,
			want:         []*domain.Tag{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				testhelper.TruncateAll(t, ctx, testDB)
			})

			testhelper.Insert(t, ctx, testDB, testhelper.Fixtures{
				Tags: tt.existingTags,
			})

			tr := &TagRepository{db: testDB}
			got, err := tr.ListAll(ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTagRepository_FindByNames(t *testing.T) {
	ctx := t.Context()

	var (
		tag1 domain.TagID = "11111111-1111-1111-1111-111111111111"
		tag2 domain.TagID = "22222222-2222-2222-2222-222222222222"
		tag3 domain.TagID = "33333333-3333-3333-3333-333333333333"
	)

	tests := []struct {
		name         string
		existingTags []*domain.Tag
		names        []string
		want         []*domain.Tag
	}{
		{
			name: "returns tags matching given names ordered by name",
			existingTags: []*domain.Tag{
				{ID: tag1, Slug: "go", Name: "Go"},
				{ID: tag2, Slug: "ts", Name: "TypeScript"},
				{ID: tag3, Slug: "db", Name: "DB"},
			},
			names: []string{"Go", "DB"},
			want: []*domain.Tag{
				{ID: tag3, Slug: "db", Name: "DB"},
				{ID: tag1, Slug: "go", Name: "Go"},
			},
		},
		{
			name: "returns empty when names is empty",
			existingTags: []*domain.Tag{
				{ID: tag1, Slug: "go", Name: "Go"},
			},
			names: nil,
			want:  []*domain.Tag{},
		},
		{
			name: "returns empty when no name matches",
			existingTags: []*domain.Tag{
				{ID: tag1, Slug: "go", Name: "Go"},
			},
			names: []string{"missing"},
			want:  []*domain.Tag{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				testhelper.TruncateAll(t, ctx, testDB)
			})

			testhelper.Insert(t, ctx, testDB, testhelper.Fixtures{Tags: tt.existingTags})

			tr := &TagRepository{db: testDB}
			got, err := tr.FindByNames(ctx, tt.names...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTagRepository_Create(t *testing.T) {
	ctx := t.Context()

	t.Run("creates tag and populates id", func(t *testing.T) {
		t.Cleanup(func() {
			testhelper.TruncateAll(t, ctx, testDB)
		})

		tr := &TagRepository{db: testDB}
		tag := &domain.Tag{Slug: "go", Name: "Go"}
		require.NoError(t, tr.Create(ctx, tag))
		require.NotEmpty(t, tag.ID)

		got, err := tr.FindByNames(ctx, "Go")
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, tag.ID, got[0].ID)
		assert.Equal(t, domain.Slug("go"), got[0].Slug)
	})

	t.Run("returns error on duplicate slug", func(t *testing.T) {
		t.Cleanup(func() {
			testhelper.TruncateAll(t, ctx, testDB)
		})

		tr := &TagRepository{db: testDB}
		require.NoError(t, tr.Create(ctx, &domain.Tag{Slug: "go", Name: "Go"}))
		err := tr.Create(ctx, &domain.Tag{Slug: "go", Name: "Golang"})
		require.Error(t, err)
	})
}

func TestTagRepository_FindByIDs(t *testing.T) {
	ctx := t.Context()

	var (
		tag1 domain.TagID = "11111111-1111-1111-1111-111111111111"
		tag2 domain.TagID = "22222222-2222-2222-2222-222222222222"
		tag3 domain.TagID = "33333333-3333-3333-3333-333333333333"
	)

	tests := []struct {
		name         string
		existingTags []*domain.Tag
		ids          []domain.TagID
		want         []*domain.Tag
	}{
		{
			name: "returns tags matching given ids ordered by name",
			existingTags: []*domain.Tag{
				{ID: tag1, Slug: "go", Name: "Go"},
				{ID: tag2, Slug: "ts", Name: "TypeScript"},
				{ID: tag3, Slug: "db", Name: "DB"},
			},
			ids: []domain.TagID{tag2, tag3},
			want: []*domain.Tag{
				{ID: tag3, Slug: "db", Name: "DB"},
				{ID: tag2, Slug: "ts", Name: "TypeScript"},
			},
		},
		{
			name: "returns empty slice when ids is empty",
			existingTags: []*domain.Tag{
				{ID: tag1, Slug: "go", Name: "Go"},
			},
			ids:  nil,
			want: []*domain.Tag{},
		},
		{
			name: "returns empty slice when no tags match",
			existingTags: []*domain.Tag{
				{ID: tag1, Slug: "go", Name: "Go"},
			},
			ids:  []domain.TagID{tag2},
			want: []*domain.Tag{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				testhelper.TruncateAll(t, ctx, testDB)
			})

			testhelper.Insert(t, ctx, testDB, testhelper.Fixtures{
				Tags: tt.existingTags,
			})

			tr := &TagRepository{db: testDB}
			got, err := tr.FindByIDs(ctx, tt.ids...)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
