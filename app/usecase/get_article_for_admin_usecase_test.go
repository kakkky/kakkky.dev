package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/testhelper/mock"
)

func TestGetArticleForAdminUsecase_Exec(t *testing.T) {
	ctx := context.Background()

	var (
		articleSlug domain.Slug      = "a1"
		articleID   domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa01"
	)

	tests := []struct {
		name    string
		input   GetArticleForAdminUsecaseInput
		mock    func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository, sr *mock.MockSeriesRepository)
		want    GetArticleUsecaseOutput
		wantErr error
	}{
		{
			name:  "returns draft article that would be hidden from public",
			input: GetArticleForAdminUsecaseInput{Slug: articleSlug},
			mock: func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository, sr *mock.MockSeriesRepository) {
				draft := &domain.Article{
					ID: articleID, Slug: articleSlug, Title: "Draft", Body: "",
					Status: domain.ArticleStatusDraft,
				}
				ar.EXPECT().FindBySlug(ctx, articleSlug).Return(draft, nil)
				tr.EXPECT().FindByIDs(ctx).Return([]*domain.Tag{}, nil)
				sr.EXPECT().FindByArticleID(ctx, articleID).Return(nil, domain.ErrNotFound)
			},
			want: GetArticleUsecaseOutput{
				Article: domain.Article{
					ID: articleID, Slug: articleSlug, Title: "Draft", Body: "",
					Status: domain.ArticleStatusDraft,
				},
				Tags: map[domain.TagID]domain.Tag{},
			},
		},
		{
			name:    "rejects empty slug before repo call",
			input:   GetArticleForAdminUsecaseInput{Slug: ""},
			mock:    func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository, sr *mock.MockSeriesRepository) {},
			wantErr: domain.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			ar := mock.NewMockArticleRepository(ctrl)
			tr := mock.NewMockTagRepository(ctrl)
			sr := mock.NewMockSeriesRepository(ctrl)
			repo := mock.NewMockRepository(ctrl)
			repo.EXPECT().NewArticleRepository().Return(ar)
			repo.EXPECT().NewTagRepository().Return(tr)
			repo.EXPECT().NewSeriesRepository().Return(sr)
			qs := mock.NewMockQueryService(ctrl)

			tt.mock(ar, tr, sr)

			uc := NewUseCase(repo, qs).NewGetArticleForAdminUsecase()
			got, err := uc.Exec(ctx, tt.input)

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
