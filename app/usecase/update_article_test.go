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

func TestUpdateArticleUsecase_Exec(t *testing.T) {
	ctx := context.Background()

	var (
		articleID     domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1"
		existingTagID domain.TagID     = "11111111-1111-1111-1111-111111111111"
	)
	existingTag := &domain.Tag{ID: existingTagID, Slug: "go", Name: "Go"}
	baseCreatedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	basePublishedAt := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)

	wantErr := errors.New("boom")

	newDraft := func() *domain.Article {
		return &domain.Article{
			ID: articleID, Slug: "old", Title: "Old", Body: "old-body",
			Status: domain.ArticleStatusDraft, CreatedAt: baseCreatedAt,
		}
	}
	newPublished := func() *domain.Article {
		return &domain.Article{
			ID: articleID, Slug: "old", Title: "Old", Body: "old-body",
			Status:      domain.ArticleStatusPublished,
			PublishedAt: basePublishedAt, CreatedAt: baseCreatedAt,
		}
	}

	tests := []struct {
		name    string
		input   UpdateArticleUsecaseInput
		mock    func(repo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository)
		wantErr error
		assert  func(t *testing.T, got UpdateArticleUsecaseOutput)
	}{
		{
			name: "updates article keeping existing slug and reuses existing tag",
			input: UpdateArticleUsecaseInput{
				CurrentSlug: "old", Title: "New", Body: "new-body",
				Status:   string(domain.ArticleStatusPublished),
				TagNames: []string{"Go"},
			},
			mock: func(repo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(withTxRunner(repo))
				repo.EXPECT().NewArticleRepository().Return(ar)
				ar.EXPECT().FindBySlug(ctx, domain.Slug("old")).Return(newDraft(), nil)
				repo.EXPECT().NewTagRepository().Return(tr)
				tr.EXPECT().FindByNames(ctx, "Go").Return([]*domain.Tag{existingTag}, nil)
				ar.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, a *domain.Article) error {
					assert.Equal(t, articleID, a.ID)
					assert.Equal(t, domain.Slug("old"), a.Slug)
					assert.Equal(t, "New", a.Title)
					assert.Equal(t, domain.ArticleStatusPublished, a.Status)
					assert.False(t, a.PublishedAt.IsZero())
					assert.Equal(t, []domain.TagID{existingTagID}, a.TagIDs)
					return nil
				})
			},
			assert: func(t *testing.T, got UpdateArticleUsecaseOutput) {
				assert.Equal(t, domain.Slug("old"), got.Slug)
			},
		},
		{
			name: "keeps existing published_at when already published",
			input: UpdateArticleUsecaseInput{
				CurrentSlug: "old", Title: "T", Body: "B",
				Status: string(domain.ArticleStatusPublished),
			},
			mock: func(repo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(withTxRunner(repo))
				repo.EXPECT().NewArticleRepository().Return(ar)
				ar.EXPECT().FindBySlug(ctx, domain.Slug("old")).Return(newPublished(), nil)
				repo.EXPECT().NewTagRepository().Return(tr)
				ar.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, a *domain.Article) error {
					assert.Equal(t, basePublishedAt, a.PublishedAt)
					return nil
				})
			},
		},
		{
			name: "clears published_at when reverting to draft",
			input: UpdateArticleUsecaseInput{
				CurrentSlug: "old", Title: "T", Body: "B",
				Status: string(domain.ArticleStatusDraft),
			},
			mock: func(repo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(withTxRunner(repo))
				repo.EXPECT().NewArticleRepository().Return(ar)
				ar.EXPECT().FindBySlug(ctx, domain.Slug("old")).Return(newPublished(), nil)
				repo.EXPECT().NewTagRepository().Return(tr)
				ar.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, a *domain.Article) error {
					assert.True(t, a.PublishedAt.IsZero())
					return nil
				})
			},
		},
		{
			name: "returns not found when article does not exist",
			input: UpdateArticleUsecaseInput{
				CurrentSlug: "missing", Title: "T", Body: "B",
				Status: string(domain.ArticleStatusDraft),
			},
			mock: func(repo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(withTxRunner(repo))
				repo.EXPECT().NewArticleRepository().Return(ar)
				ar.EXPECT().FindBySlug(ctx, domain.Slug("missing")).Return(nil, domain.ErrNotFound)
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "returns error on invalid article (empty title)",
			input: UpdateArticleUsecaseInput{
				CurrentSlug: "old", Title: "", Body: "B",
				Status: string(domain.ArticleStatusDraft),
			},
			mock: func(repo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(withTxRunner(repo))
				repo.EXPECT().NewArticleRepository().Return(ar)
				ar.EXPECT().FindBySlug(ctx, domain.Slug("old")).Return(newDraft(), nil)
			},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "propagates update error",
			input: UpdateArticleUsecaseInput{
				CurrentSlug: "old", Title: "T", Body: "B",
				Status: string(domain.ArticleStatusDraft),
			},
			mock: func(repo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(withTxRunner(repo))
				repo.EXPECT().NewArticleRepository().Return(ar)
				ar.EXPECT().FindBySlug(ctx, domain.Slug("old")).Return(newDraft(), nil)
				repo.EXPECT().NewTagRepository().Return(tr)
				ar.EXPECT().Update(ctx, gomock.Any()).Return(wantErr)
			},
			wantErr: wantErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			repo := mock.NewMockRepository(ctrl)
			ar := mock.NewMockArticleRepository(ctrl)
			tr := mock.NewMockTagRepository(ctrl)
			qs := mock.NewMockQueryService(ctrl)

			tt.mock(repo, ar, tr)

			us := NewUseCase(repo, qs).NewUpdateArticleUsecase()
			got, err := us.Exec(ctx, tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, UpdateArticleUsecaseOutput{}, got)
				return
			}
			require.NoError(t, err)
			if tt.assert != nil {
				tt.assert(t, got)
			}
		})
	}
}
