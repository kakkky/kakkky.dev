package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/testhelper/mock"
)

func TestUpdateArticleUsecase_Exec(t *testing.T) {
	ctx := context.Background()

	var (
		articleID    domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1"
		existingTag1 domain.TagID     = "11111111-1111-1111-1111-111111111111"
		existingTag2 domain.TagID     = "22222222-2222-2222-2222-222222222222"
		newTagID     domain.TagID     = "99999999-9999-9999-9999-999999999999"
	)

	wantDBErr := errors.New("boom")

	tests := []struct {
		name    string
		input   UpdateArticleUsecaseInput
		mock    func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository)
		wantErr error
	}{
		{
			name: "updates fields and replaces tags in single tx",
			input: UpdateArticleUsecaseInput{
				Slug:           "first",
				Title:          "New Title",
				Body:           "New body",
				Status:         domain.ArticleStatusPublished,
				ExistingTagIDs: []domain.TagID{existingTag2},
				NewTagNames:    []string{"Rust"},
			},
			mock: func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewArticleRepository().Return(ar)
				txRepo.EXPECT().NewTagRepository().Return(tr)

				ar.EXPECT().FindBySlug(ctx, domain.Slug("first")).Return(&domain.Article{
					ID: articleID, Slug: "first", Title: "Old", Body: "Old body",
					Status: domain.ArticleStatusDraft,
					TagIDs: []domain.TagID{existingTag1},
				}, nil)
				tr.EXPECT().Store(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, tag *domain.Tag) error {
					assert.Equal(t, "Rust", tag.Name)
					tag.ID = newTagID
					return nil
				})
				ar.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, a *domain.Article) error {
					assert.Equal(t, articleID, a.ID)
					assert.Equal(t, "New Title", a.Title)
					assert.Equal(t, "New body", a.Body)
					assert.Equal(t, domain.ArticleStatusPublished, a.Status)
					assert.ElementsMatch(t, []domain.TagID{existingTag2, newTagID}, a.TagIDs)
					assert.False(t, a.PublishedAt.IsZero(), "draft→published transition should set PublishedAt")
					return nil
				})
			},
		},
		{
			name: "updates only fields when no new tag names given",
			input: UpdateArticleUsecaseInput{
				Slug:           "first",
				Title:          "New Title",
				Body:           "",
				Status:         domain.ArticleStatusDraft,
				ExistingTagIDs: []domain.TagID{existingTag1},
			},
			mock: func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewArticleRepository().Return(ar)
				txRepo.EXPECT().NewTagRepository().Return(tr)

				ar.EXPECT().FindBySlug(ctx, domain.Slug("first")).Return(&domain.Article{
					ID: articleID, Slug: "first", Title: "Old", Body: "Old body",
					Status: domain.ArticleStatusDraft,
					TagIDs: []domain.TagID{existingTag1},
				}, nil)
				ar.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, a *domain.Article) error {
					assert.Equal(t, []domain.TagID{existingTag1}, a.TagIDs)
					return nil
				})
			},
		},
		{
			name: "returns not found when slug does not exist",
			input: UpdateArticleUsecaseInput{
				Slug:   "missing",
				Title:  "New",
				Status: domain.ArticleStatusDraft,
			},
			mock: func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewArticleRepository().Return(ar)
				txRepo.EXPECT().NewTagRepository().Return(tr)

				ar.EXPECT().FindBySlug(ctx, domain.Slug("missing")).Return(nil, domain.ErrNotFound)
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "translates tag ErrAlreadyExists to ErrInvalidArgument",
			input: UpdateArticleUsecaseInput{
				Slug:        "first",
				Title:       "New",
				Status:      domain.ArticleStatusDraft,
				NewTagNames: []string{"Go"},
			},
			mock: func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewArticleRepository().Return(ar)
				txRepo.EXPECT().NewTagRepository().Return(tr)

				ar.EXPECT().FindBySlug(ctx, domain.Slug("first")).Return(&domain.Article{
					ID: articleID, Slug: "first", Title: "Old", Body: "",
					Status: domain.ArticleStatusDraft,
				}, nil)
				tr.EXPECT().Store(ctx, gomock.Any()).Return(domain.ErrAlreadyExists)
			},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "propagates unexpected article Update error",
			input: UpdateArticleUsecaseInput{
				Slug:   "first",
				Title:  "New",
				Status: domain.ArticleStatusDraft,
			},
			mock: func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewArticleRepository().Return(ar)
				txRepo.EXPECT().NewTagRepository().Return(tr)

				ar.EXPECT().FindBySlug(ctx, domain.Slug("first")).Return(&domain.Article{
					ID: articleID, Slug: "first", Title: "Old", Body: "",
					Status: domain.ArticleStatusDraft,
				}, nil)
				ar.EXPECT().Update(ctx, gomock.Any()).Return(wantDBErr)
			},
			wantErr: wantDBErr,
		},
		{
			name: "rejects empty slug before tx",
			input: UpdateArticleUsecaseInput{
				Slug:   "",
				Title:  "New",
				Status: domain.ArticleStatusDraft,
			},
			mock:    func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "rejects duplicated new tag names before tx",
			input: UpdateArticleUsecaseInput{
				Slug:        "first",
				Title:       "New",
				Status:      domain.ArticleStatusDraft,
				NewTagNames: []string{"Rust", "Rust"},
			},
			mock:    func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "rejects empty new tag name before tx",
			input: UpdateArticleUsecaseInput{
				Slug:        "first",
				Title:       "New",
				Status:      domain.ArticleStatusDraft,
				NewTagNames: []string{""},
			},
			mock:    func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "propagates invalid title validation from domain",
			input: UpdateArticleUsecaseInput{
				Slug:   "first",
				Title:  "",
				Status: domain.ArticleStatusDraft,
			},
			mock: func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewArticleRepository().Return(ar)
				txRepo.EXPECT().NewTagRepository().Return(tr)

				ar.EXPECT().FindBySlug(ctx, domain.Slug("first")).Return(&domain.Article{
					ID: articleID, Slug: "first", Title: "Old", Body: "",
					Status: domain.ArticleStatusDraft,
				}, nil)
			},
			wantErr: domain.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			repo := mock.NewMockRepository(ctrl)
			txRepo := mock.NewMockRepository(ctrl)
			ar := mock.NewMockArticleRepository(ctrl)
			tr := mock.NewMockTagRepository(ctrl)
			qs := mock.NewMockQueryService(ctrl)

			tt.mock(repo, txRepo, ar, tr)

			uc := NewUseCase(repo, qs).NewUpdateArticleUsecase()
			out, err := uc.Exec(ctx, tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, UpdateArticleUsecaseOutput{}, out)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.input.Slug, out.ArticleSlug)
		})
	}
}
