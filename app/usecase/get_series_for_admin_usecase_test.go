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

func TestGetSeriesForAdminUsecase_Exec(t *testing.T) {
	ctx := context.Background()

	var (
		seriesSlug domain.Slug     = "s1"
		seriesID   domain.SeriesID = "cccccccc-cccc-cccc-cccc-ccccccccccc1"
		article1ID domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa01"
	)

	draftSeriesWithDraftArticle := &domain.Series{
		ID: seriesID, Slug: seriesSlug, Title: "S1",
		Status: domain.SeriesStatusDraft,
		Articles: []domain.SeriesArticle{
			{ArticleID: article1ID, Position: 1},
		},
	}
	draftArticle := &domain.Article{
		ID: article1ID, Slug: "a1", Title: "A1",
		Status: domain.ArticleStatusDraft,
	}

	tests := []struct {
		name    string
		input   GetSeriesForAdminUsecaseInput
		mock    func(sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository)
		want    GetSeriesUsecaseOutput
		wantErr error
	}{
		{
			name:  "returns draft series and includes draft articles",
			input: GetSeriesForAdminUsecaseInput{Slug: seriesSlug},
			mock: func(sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				sr.EXPECT().FindBySlug(ctx, seriesSlug).Return(draftSeriesWithDraftArticle, nil)
				ar.EXPECT().FindByIDs(ctx, article1ID).Return([]*domain.Article{draftArticle}, nil)
				tr.EXPECT().FindByIDs(ctx).Return([]*domain.Tag{}, nil)
			},
			want: GetSeriesUsecaseOutput{
				Series: *draftSeriesWithDraftArticle,
				Articles: []GetSeriesUsecaseSeriesArticle{
					{Article: *draftArticle, Position: 1},
				},
				Tags: map[domain.TagID]domain.Tag{},
			},
		},
		{
			name:    "rejects empty slug",
			input:   GetSeriesForAdminUsecaseInput{Slug: ""},
			mock:    func(sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {},
			wantErr: domain.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			sr := mock.NewMockSeriesRepository(ctrl)
			ar := mock.NewMockArticleRepository(ctrl)
			tr := mock.NewMockTagRepository(ctrl)
			repo := mock.NewMockRepository(ctrl)
			repo.EXPECT().NewSeriesRepository().Return(sr)
			repo.EXPECT().NewArticleRepository().Return(ar)
			repo.EXPECT().NewTagRepository().Return(tr)
			qs := mock.NewMockQueryService(ctrl)

			tt.mock(sr, ar, tr)

			uc := NewUseCase(repo, qs).NewGetSeriesForAdminUsecase()
			got, err := uc.Exec(ctx, tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, GetSeriesUsecaseOutput{}, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
