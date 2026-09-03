package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/testhelper"
)

func TestTagRepository_List(t *testing.T) {
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
			got, err := tr.List(ctx)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTagRepository_Store(t *testing.T) {
	ctx := t.Context()

	var existingID domain.TagID = "11111111-1111-1111-1111-111111111111"

	tests := []struct {
		name         string
		existingTags []*domain.Tag
		tag          *domain.Tag
		wantErr      error
	}{
		{
			name: "inserts tag and writes back generated ID",
			tag:  &domain.Tag{Slug: "go", Name: "Go"},
		},
		{
			name: "returns ErrAlreadyExists when slug conflicts",
			existingTags: []*domain.Tag{
				{ID: existingID, Slug: "go", Name: "Go"},
			},
			tag:     &domain.Tag{Slug: "go", Name: "Golang"},
			wantErr: domain.ErrAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				testhelper.TruncateAll(t, ctx, testDB)
			})

			testhelper.Insert(t, ctx, testDB, testhelper.Fixtures{Tags: tt.existingTags})

			tr := &TagRepository{db: testDB}
			err := tr.Store(ctx, tt.tag)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, tt.tag.ID)

			got, err := tr.FindByIDs(ctx, tt.tag.ID)
			require.NoError(t, err)
			require.Len(t, got, 1)
			assert.Equal(t, tt.tag.Slug, got[0].Slug)
			assert.Equal(t, tt.tag.Name, got[0].Name)
		})
	}
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
