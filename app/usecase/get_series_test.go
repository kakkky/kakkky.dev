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

func TestGetSeriesUsecase_Exec(t *testing.T) {
	ctx := context.Background()
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	var (
		seriesSlug domain.Slug = "s1"

		seriesID domain.SeriesID = "cccccccc-cccc-cccc-cccc-ccccccccccc1"

		article1ID domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa01"
		article2ID domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa02"
		article3ID domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa03"

		tag1ID domain.TagID = "11111111-1111-1111-1111-111111111111"
		tag2ID domain.TagID = "22222222-2222-2222-2222-222222222222"
		tag3ID domain.TagID = "33333333-3333-3333-3333-333333333333"
	)

	tag1 := &domain.Tag{ID: tag1ID, Slug: "go", Name: "Go"}
	tag2 := &domain.Tag{ID: tag2ID, Slug: "db", Name: "DB"}
	tag3 := &domain.Tag{ID: tag3ID, Slug: "ts", Name: "TypeScript"}

	publishedSeries := &domain.Series{
		ID:          seriesID,
		Slug:        seriesSlug,
		Title:       "S1",
		Description: "desc",
		Status:      domain.SeriesStatusPublishedOngoing,
		PublishedAt: baseTime,
		TagIDs:      []domain.TagID{tag1ID},
		Articles: []domain.SeriesArticle{
			{ArticleID: article1ID, Position: 3},
			{ArticleID: article2ID, Position: 1},
			{ArticleID: article3ID, Position: 2},
		},
	}
	draftSeries := &domain.Series{
		ID:     seriesID,
		Slug:   seriesSlug,
		Title:  "S1",
		Status: domain.SeriesStatusDraft,
	}

	article1 := &domain.Article{
		ID:          article1ID,
		Slug:        "a1",
		Title:       "A1",
		Status:      domain.ArticleStatusPublished,
		PublishedAt: baseTime,
		TagIDs:      []domain.TagID{tag2ID},
	}
	article2 := &domain.Article{
		ID:          article2ID,
		Slug:        "a2",
		Title:       "A2",
		Status:      domain.ArticleStatusPublished,
		PublishedAt: baseTime,
		TagIDs:      []domain.TagID{tag1ID, tag3ID},
	}
	article3Draft := &domain.Article{
		ID:     article3ID,
		Slug:   "a3",
		Title:  "A3",
		Status: domain.ArticleStatusDraft,
	}

	wantErr := errors.New("boom")

	tests := []struct {
		name    string
		input   GetSeriesUsecaseInput
		mock    func(sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository)
		want    GetSeriesUsecaseOutput
		wantErr error
	}{
		{
			name:  "returns Series with published articles sorted by position and deduplicated tags",
			input: GetSeriesUsecaseInput{Slug: seriesSlug},
			mock: func(sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				sr.EXPECT().FindBySlug(ctx, seriesSlug).Return(publishedSeries, nil)
				ar.EXPECT().FindByIDs(ctx, article1ID, article2ID, article3ID).
					Return([]*domain.Article{article1, article2, article3Draft}, nil)
				tr.EXPECT().FindByIDs(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, ids ...domain.TagID) ([]*domain.Tag, error) {
						assert.ElementsMatch(t, []domain.TagID{tag1ID, tag2ID, tag3ID}, ids)
						return []*domain.Tag{tag1, tag2, tag3}, nil
					})
			},
			want: GetSeriesUsecaseOutput{
				Series: *publishedSeries,
				Articles: []GetSeriesUsecaseSeriesArticle{
					{Article: *article2, Position: 1},
					{Article: *article1, Position: 3},
				},
				Tags: map[domain.TagID]domain.Tag{
					tag1ID: *tag1,
					tag2ID: *tag2,
					tag3ID: *tag3,
				},
			},
		},
		{
			name:  "wraps ErrNotFound with a user-facing message when FindBySlug returns ErrNotFound",
			input: GetSeriesUsecaseInput{Slug: seriesSlug},
			mock: func(sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				sr.EXPECT().FindBySlug(ctx, seriesSlug).Return(nil, domain.ErrNotFound)
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name:  "returns ErrNotFound when series status is draft",
			input: GetSeriesUsecaseInput{Slug: seriesSlug},
			mock: func(sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				sr.EXPECT().FindBySlug(ctx, seriesSlug).Return(draftSeries, nil)
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name:  "propagates error from SeriesRepository.FindBySlug",
			input: GetSeriesUsecaseInput{Slug: seriesSlug},
			mock: func(sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				sr.EXPECT().FindBySlug(ctx, seriesSlug).Return(nil, wantErr)
			},
			wantErr: wantErr,
		},
		{
			name:  "propagates error from ArticleRepository.FindByIDs and skips tag fetch",
			input: GetSeriesUsecaseInput{Slug: seriesSlug},
			mock: func(sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				sr.EXPECT().FindBySlug(ctx, seriesSlug).Return(publishedSeries, nil)
				ar.EXPECT().FindByIDs(ctx, article1ID, article2ID, article3ID).Return(nil, wantErr)
			},
			wantErr: wantErr,
		},
		{
			name:  "propagates error from TagRepository.FindByIDs",
			input: GetSeriesUsecaseInput{Slug: seriesSlug},
			mock: func(sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				sr.EXPECT().FindBySlug(ctx, seriesSlug).Return(publishedSeries, nil)
				ar.EXPECT().FindByIDs(ctx, article1ID, article2ID, article3ID).
					Return([]*domain.Article{article1, article2, article3Draft}, nil)
				tr.EXPECT().FindByIDs(ctx, gomock.Any()).Return(nil, wantErr)
			},
			wantErr: wantErr,
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

			gs := NewUseCase(repo, qs).NewGetSeriesUsecase()
			got, err := gs.Exec(ctx, tt.input)

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
