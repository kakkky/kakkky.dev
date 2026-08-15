package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/testhelper/mock"
)

func TestGetArticleUsecase_Exec(t *testing.T) {
	ctx := context.Background()
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	var (
		tag1ID domain.TagID = "11111111-1111-1111-1111-111111111111"
		tag2ID domain.TagID = "22222222-2222-2222-2222-222222222222"

		articleSlug domain.Slug = "a1"
	)

	tag1 := &domain.Tag{ID: tag1ID, Slug: "go", Name: "Go"}
	tag2 := &domain.Tag{ID: tag2ID, Slug: "db", Name: "DB"}

	article := &domain.Article{
		ID:          "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa01",
		Slug:        articleSlug,
		Title:       "A1",
		Body:        "# hello",
		Status:      domain.ArticleStatusPublished,
		PublishedAt: baseTime,
		TagIDs:      []domain.TagID{tag1ID, tag2ID},
	}
	articleNoTags := &domain.Article{
		ID:          "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa02",
		Slug:        articleSlug,
		Title:       "A2",
		Body:        "# hi",
		Status:      domain.ArticleStatusPublished,
		PublishedAt: baseTime,
	}

	wantErr := errors.New("boom")

	tests := []struct {
		name    string
		input   GetArticleUsecaseInput
		mock    func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository)
		want    GetArticleUsecaseOutput
		wantErr error
	}{
		{
			name:  "returns Article and Tags keyed by TagID when FindBySlug and FindByIDs both succeed",
			input: GetArticleUsecaseInput{Slug: articleSlug},
			mock: func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				ar.EXPECT().FindBySlug(ctx, articleSlug).Return(article, nil)
				tr.EXPECT().FindByIDs(ctx, tag1ID, tag2ID).Return([]*domain.Tag{tag1, tag2}, nil)
			},
			want: GetArticleUsecaseOutput{
				Article: *article,
				Tags: map[domain.TagID]domain.Tag{
					tag1ID: *tag1,
					tag2ID: *tag2,
				},
			},
		},
		{
			name:  "returns Article with empty Tags map when the article has no TagIDs",
			input: GetArticleUsecaseInput{Slug: articleSlug},
			mock: func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				ar.EXPECT().FindBySlug(ctx, articleSlug).Return(articleNoTags, nil)
				tr.EXPECT().FindByIDs(ctx).Return([]*domain.Tag{}, nil)
			},
			want: GetArticleUsecaseOutput{
				Article: *articleNoTags,
				Tags:    map[domain.TagID]domain.Tag{},
			},
		},
		{
			name:  "returns zero Output and nil error when FindBySlug returns ErrNotFound",
			input: GetArticleUsecaseInput{Slug: articleSlug},
			mock: func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				ar.EXPECT().FindBySlug(ctx, articleSlug).Return(nil, domain.ErrNotFound)
			},
			want: GetArticleUsecaseOutput{},
		},
		{
			name:  "propagates error from ArticleRepository.FindBySlug and skips FindByIDs",
			input: GetArticleUsecaseInput{Slug: articleSlug},
			mock: func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				ar.EXPECT().FindBySlug(ctx, articleSlug).Return(nil, wantErr)
			},
			wantErr: wantErr,
		},
		{
			name:  "propagates error from TagRepository.FindByIDs",
			input: GetArticleUsecaseInput{Slug: articleSlug},
			mock: func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				ar.EXPECT().FindBySlug(ctx, articleSlug).Return(article, nil)
				tr.EXPECT().FindByIDs(ctx, tag1ID, tag2ID).Return(nil, wantErr)
			},
			wantErr: wantErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			ar := mock.NewMockArticleRepository(ctrl)
			tr := mock.NewMockTagRepository(ctrl)
			repo := mock.NewMockRepository(ctrl)
			repo.EXPECT().NewArticleRepository().Return(ar)
			repo.EXPECT().NewTagRepository().Return(tr)

			qs := mock.NewMockQueryService(ctrl)

			tt.mock(ar, tr)

			ga := NewUseCase(repo, qs).NewGetArticleUsecase()
			got, err := ga.Exec(ctx, tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, GetArticleUsecaseOutput{}, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
