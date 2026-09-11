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

func TestCreateArticleUsecase_Exec(t *testing.T) {
	ctx := context.Background()

	var (
		existingTag1 domain.TagID = "11111111-1111-1111-1111-111111111111"
		existingTag2 domain.TagID = "22222222-2222-2222-2222-222222222222"
	)

	wantDBErr := errors.New("boom")

	tests := []struct {
		name    string
		input   CreateArticleUsecaseInput
		mock    func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository)
		wantErr error
	}{
		{
			name: "creates draft article with existing + new tags in single tx",
			input: CreateArticleUsecaseInput{
				Title:          "Hello",
				ExistingTagIDs: []domain.TagID{existingTag1, existingTag2},
				NewTagNames:    []string{"Rust"},
			},
			mock: func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewTagRepository().Return(tr)
				txRepo.EXPECT().NewArticleRepository().Return(ar)

				tr.EXPECT().Store(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, tag *domain.Tag) error {
					assert.Equal(t, "Rust", tag.Name)
					tag.ID = "99999999-9999-9999-9999-999999999999"
					return nil
				})
				ar.EXPECT().Store(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, a *domain.Article) error {
					assert.Equal(t, "Hello", a.Title)
					assert.Equal(t, domain.ArticleStatusDraft, a.Status)
					assert.ElementsMatch(t, []domain.TagID{existingTag1, existingTag2, "99999999-9999-9999-9999-999999999999"}, a.TagIDs)
					a.ID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa01"
					return nil
				})
			},
		},
		{
			name: "translates tag ErrAlreadyExists to ErrInvalidArgument with user-facing message",
			input: CreateArticleUsecaseInput{
				Title:       "Hello",
				NewTagNames: []string{"Go"},
			},
			mock: func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewTagRepository().Return(tr)
				txRepo.EXPECT().NewArticleRepository().Return(ar)

				tr.EXPECT().Store(ctx, gomock.Any()).Return(domain.ErrAlreadyExists)
			},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "translates article ErrAlreadyExists (slug conflict) to ErrInvalidArgument",
			input: CreateArticleUsecaseInput{
				Title: "Hello",
			},
			mock: func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewTagRepository().Return(tr)
				txRepo.EXPECT().NewArticleRepository().Return(ar)

				ar.EXPECT().Store(ctx, gomock.Any()).Return(domain.ErrAlreadyExists)
			},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "propagates unexpected article Store error",
			input: CreateArticleUsecaseInput{
				Title: "Hello",
			},
			mock: func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewTagRepository().Return(tr)
				txRepo.EXPECT().NewArticleRepository().Return(ar)

				ar.EXPECT().Store(ctx, gomock.Any()).Return(wantDBErr)
			},
			wantErr: wantDBErr,
		},
		{
			name: "rejects empty title before tx",
			input: CreateArticleUsecaseInput{
				Title: "",
			},
			mock:    func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "rejects duplicated new tag names before tx",
			input: CreateArticleUsecaseInput{
				Title:       "Hello",
				NewTagNames: []string{"Rust", "Rust"},
			},
			mock:    func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "rejects empty new tag name before tx",
			input: CreateArticleUsecaseInput{
				Title:       "Hello",
				NewTagNames: []string{""},
			},
			mock:    func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {},
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

			uc := NewUseCase(repo, qs).NewCreateArticleUsecase()
			out, err := uc.Exec(ctx, tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, CreateArticleUsecaseOutput{}, out)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, out.ArticleSlug)
		})
	}
}
