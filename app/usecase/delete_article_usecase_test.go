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

func TestDeleteArticleUsecase_Exec(t *testing.T) {
	ctx := context.Background()

	var (
		articleID domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1"
	)
	wantDBErr := errors.New("boom")

	tests := []struct {
		name      string
		input     DeleteArticleUsecaseInput
		mock      func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository)
		wantTitle string
		wantErr   error
	}{
		{
			name:  "deletes article and returns title",
			input: DeleteArticleUsecaseInput{ID: articleID},
			mock: func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewArticleRepository().Return(ar)

				ar.EXPECT().FindByIDs(ctx, articleID).Return([]*domain.Article{
					{ID: articleID, Title: "Target"},
				}, nil)
				ar.EXPECT().Delete(ctx, articleID).Return(nil)
			},
			wantTitle: "Target",
		},
		{
			name:  "returns not found when article is missing before delete",
			input: DeleteArticleUsecaseInput{ID: articleID},
			mock: func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewArticleRepository().Return(ar)

				ar.EXPECT().FindByIDs(ctx, articleID).Return([]*domain.Article{}, nil)
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name:  "propagates unexpected delete error",
			input: DeleteArticleUsecaseInput{ID: articleID},
			mock: func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewArticleRepository().Return(ar)

				ar.EXPECT().FindByIDs(ctx, articleID).Return([]*domain.Article{
					{ID: articleID, Title: "Target"},
				}, nil)
				ar.EXPECT().Delete(ctx, articleID).Return(wantDBErr)
			},
			wantErr: wantDBErr,
		},
		{
			name:    "rejects empty id before tx",
			input:   DeleteArticleUsecaseInput{ID: ""},
			mock:    func(repo, txRepo *mock.MockRepository, ar *mock.MockArticleRepository) {},
			wantErr: domain.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			repo := mock.NewMockRepository(ctrl)
			txRepo := mock.NewMockRepository(ctrl)
			ar := mock.NewMockArticleRepository(ctrl)
			qs := mock.NewMockQueryService(ctrl)

			tt.mock(repo, txRepo, ar)

			uc := NewUseCase(repo, qs).NewDeleteArticleUsecase()
			out, err := uc.Exec(ctx, tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, DeleteArticleUsecaseOutput{}, out)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantTitle, out.Title)
		})
	}
}
