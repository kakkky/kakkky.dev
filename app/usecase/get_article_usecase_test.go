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

		currentID domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa02"
		prevID    domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa01"
		nextID    domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa03"

		seriesID   domain.SeriesID = "cccccccc-cccc-cccc-cccc-ccccccccccc1"
		seriesSlug domain.Slug     = "s1"
	)

	tag1 := &domain.Tag{ID: tag1ID, Slug: "go", Name: "Go"}
	tag2 := &domain.Tag{ID: tag2ID, Slug: "db", Name: "DB"}

	article := &domain.Article{
		ID:          currentID,
		Slug:        articleSlug,
		Title:       "A2",
		Body:        "# hello",
		Status:      domain.ArticleStatusPublished,
		PublishedAt: baseTime,
		TagIDs:      []domain.TagID{tag1ID, tag2ID},
	}
	articleNoTags := &domain.Article{
		ID:          currentID,
		Slug:        articleSlug,
		Title:       "A2",
		Body:        "# hi",
		Status:      domain.ArticleStatusPublished,
		PublishedAt: baseTime,
	}

	prevArticle := &domain.Article{
		ID:     prevID,
		Slug:   "a1",
		Title:  "A1",
		Status: domain.ArticleStatusPublished,
	}
	nextArticle := &domain.Article{
		ID:     nextID,
		Slug:   "a3",
		Title:  "A3",
		Status: domain.ArticleStatusPublished,
	}
	prevDraft := &domain.Article{
		ID:     prevID,
		Slug:   "a1",
		Title:  "A1",
		Status: domain.ArticleStatusDraft,
	}

	seriesMiddle := &domain.Series{
		ID: seriesID, Slug: seriesSlug, Title: "S1", Status: domain.SeriesStatusPublishedOngoing,
		Articles: []domain.SeriesArticle{
			{ArticleID: prevID, Position: 1},
			{ArticleID: currentID, Position: 2},
			{ArticleID: nextID, Position: 3},
		},
	}
	seriesHead := &domain.Series{
		ID: seriesID, Slug: seriesSlug, Title: "S1", Status: domain.SeriesStatusPublishedOngoing,
		Articles: []domain.SeriesArticle{
			{ArticleID: currentID, Position: 1},
			{ArticleID: nextID, Position: 2},
		},
	}
	seriesTail := &domain.Series{
		ID: seriesID, Slug: seriesSlug, Title: "S1", Status: domain.SeriesStatusPublishedOngoing,
		Articles: []domain.SeriesArticle{
			{ArticleID: prevID, Position: 1},
			{ArticleID: currentID, Position: 2},
		},
	}
	seriesDraft := &domain.Series{
		ID: seriesID, Slug: seriesSlug, Title: "S1", Status: domain.SeriesStatusDraft,
		Articles: []domain.SeriesArticle{{ArticleID: currentID, Position: 1}},
	}
	seriesWithoutCurrent := &domain.Series{
		ID: seriesID, Slug: seriesSlug, Title: "S1", Status: domain.SeriesStatusPublishedOngoing,
		Articles: []domain.SeriesArticle{{ArticleID: prevID, Position: 1}},
	}

	wantErr := errors.New("boom")

	tests := []struct {
		name    string
		input   GetArticleUsecaseInput
		mock    func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository, sr *mock.MockSeriesRepository)
		want    GetArticleUsecaseOutput
		wantErr error
	}{
		{
			name:  "returns Article and Tags with nil InSeriesRef when series is not found",
			input: GetArticleUsecaseInput{Slug: articleSlug},
			mock: func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository, sr *mock.MockSeriesRepository) {
				ar.EXPECT().FindBySlug(ctx, articleSlug).Return(article, nil)
				tr.EXPECT().FindByIDs(ctx, tag1ID, tag2ID).Return([]*domain.Tag{tag1, tag2}, nil)
				sr.EXPECT().FindByArticleID(ctx, currentID).Return(nil, domain.ErrNotFound)
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
			mock: func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository, sr *mock.MockSeriesRepository) {
				ar.EXPECT().FindBySlug(ctx, articleSlug).Return(articleNoTags, nil)
				tr.EXPECT().FindByIDs(ctx).Return([]*domain.Tag{}, nil)
				sr.EXPECT().FindByArticleID(ctx, currentID).Return(nil, domain.ErrNotFound)
			},
			want: GetArticleUsecaseOutput{
				Article: *articleNoTags,
				Tags:    map[domain.TagID]domain.Tag{},
			},
		},
		{
			name:  "fills prev and next in InSeriesRef when current article is in the middle of a published series",
			input: GetArticleUsecaseInput{Slug: articleSlug},
			mock: func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository, sr *mock.MockSeriesRepository) {
				ar.EXPECT().FindBySlug(ctx, articleSlug).Return(article, nil)
				tr.EXPECT().FindByIDs(ctx, tag1ID, tag2ID).Return([]*domain.Tag{tag1, tag2}, nil)
				sr.EXPECT().FindByArticleID(ctx, currentID).Return(seriesMiddle, nil)
				ar.EXPECT().FindByIDs(ctx, prevID, nextID).Return([]*domain.Article{prevArticle, nextArticle}, nil)
			},
			want: GetArticleUsecaseOutput{
				Article: *article,
				Tags: map[domain.TagID]domain.Tag{
					tag1ID: *tag1,
					tag2ID: *tag2,
				},
				InSeriesRef: &GetArticleUsecaseSeriesRef{
					Slug:                   seriesSlug,
					Title:                  "S1",
					Status:                 domain.SeriesStatusPublishedOngoing,
					CurrentArticlePosition: 2,
					PrevArticleRef:         &GetArticleUsecaseSeriesNeighborArticleRef{Slug: "a1", Title: "A1", Position: 1},
					NextArticleRef:         &GetArticleUsecaseSeriesNeighborArticleRef{Slug: "a3", Title: "A3", Position: 3},
				},
			},
		},
		{
			name:  "leaves prev nil when current article is at the head of a series",
			input: GetArticleUsecaseInput{Slug: articleSlug},
			mock: func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository, sr *mock.MockSeriesRepository) {
				ar.EXPECT().FindBySlug(ctx, articleSlug).Return(article, nil)
				tr.EXPECT().FindByIDs(ctx, tag1ID, tag2ID).Return([]*domain.Tag{tag1, tag2}, nil)
				sr.EXPECT().FindByArticleID(ctx, currentID).Return(seriesHead, nil)
				ar.EXPECT().FindByIDs(ctx, nextID).Return([]*domain.Article{nextArticle}, nil)
			},
			want: GetArticleUsecaseOutput{
				Article: *article,
				Tags: map[domain.TagID]domain.Tag{
					tag1ID: *tag1,
					tag2ID: *tag2,
				},
				InSeriesRef: &GetArticleUsecaseSeriesRef{
					Slug:                   seriesSlug,
					Title:                  "S1",
					Status:                 domain.SeriesStatusPublishedOngoing,
					CurrentArticlePosition: 1,
					NextArticleRef:         &GetArticleUsecaseSeriesNeighborArticleRef{Slug: "a3", Title: "A3", Position: 2},
				},
			},
		},
		{
			name:  "leaves next nil when current article is at the tail of a series",
			input: GetArticleUsecaseInput{Slug: articleSlug},
			mock: func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository, sr *mock.MockSeriesRepository) {
				ar.EXPECT().FindBySlug(ctx, articleSlug).Return(article, nil)
				tr.EXPECT().FindByIDs(ctx, tag1ID, tag2ID).Return([]*domain.Tag{tag1, tag2}, nil)
				sr.EXPECT().FindByArticleID(ctx, currentID).Return(seriesTail, nil)
				ar.EXPECT().FindByIDs(ctx, prevID).Return([]*domain.Article{prevArticle}, nil)
			},
			want: GetArticleUsecaseOutput{
				Article: *article,
				Tags: map[domain.TagID]domain.Tag{
					tag1ID: *tag1,
					tag2ID: *tag2,
				},
				InSeriesRef: &GetArticleUsecaseSeriesRef{
					Slug:                   seriesSlug,
					Title:                  "S1",
					Status:                 domain.SeriesStatusPublishedOngoing,
					CurrentArticlePosition: 2,
					PrevArticleRef:         &GetArticleUsecaseSeriesNeighborArticleRef{Slug: "a1", Title: "A1", Position: 1},
				},
			},
		},
		{
			name:  "excludes draft neighbor article from InSeriesRef",
			input: GetArticleUsecaseInput{Slug: articleSlug},
			mock: func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository, sr *mock.MockSeriesRepository) {
				ar.EXPECT().FindBySlug(ctx, articleSlug).Return(article, nil)
				tr.EXPECT().FindByIDs(ctx, tag1ID, tag2ID).Return([]*domain.Tag{tag1, tag2}, nil)
				sr.EXPECT().FindByArticleID(ctx, currentID).Return(seriesMiddle, nil)
				ar.EXPECT().FindByIDs(ctx, prevID, nextID).Return([]*domain.Article{prevDraft, nextArticle}, nil)
			},
			want: GetArticleUsecaseOutput{
				Article: *article,
				Tags: map[domain.TagID]domain.Tag{
					tag1ID: *tag1,
					tag2ID: *tag2,
				},
				InSeriesRef: &GetArticleUsecaseSeriesRef{
					Slug:                   seriesSlug,
					Title:                  "S1",
					Status:                 domain.SeriesStatusPublishedOngoing,
					CurrentArticlePosition: 2,
					NextArticleRef:         &GetArticleUsecaseSeriesNeighborArticleRef{Slug: "a3", Title: "A3", Position: 3},
				},
			},
		},
		{
			name:  "returns nil InSeriesRef and skips neighbor fetch when series is draft",
			input: GetArticleUsecaseInput{Slug: articleSlug},
			mock: func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository, sr *mock.MockSeriesRepository) {
				ar.EXPECT().FindBySlug(ctx, articleSlug).Return(article, nil)
				tr.EXPECT().FindByIDs(ctx, tag1ID, tag2ID).Return([]*domain.Tag{tag1, tag2}, nil)
				sr.EXPECT().FindByArticleID(ctx, currentID).Return(seriesDraft, nil)
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
			name:  "returns ErrInternal when current article is missing from series.Articles (invariant violation)",
			input: GetArticleUsecaseInput{Slug: articleSlug},
			mock: func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository, sr *mock.MockSeriesRepository) {
				ar.EXPECT().FindBySlug(ctx, articleSlug).Return(article, nil)
				tr.EXPECT().FindByIDs(ctx, tag1ID, tag2ID).Return([]*domain.Tag{tag1, tag2}, nil)
				sr.EXPECT().FindByArticleID(ctx, currentID).Return(seriesWithoutCurrent, nil)
			},
			wantErr: domain.ErrInternal,
		},
		{
			name:  "wraps ErrNotFound with a user-facing message when FindBySlug returns ErrNotFound",
			input: GetArticleUsecaseInput{Slug: articleSlug},
			mock: func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository, sr *mock.MockSeriesRepository) {
				ar.EXPECT().FindBySlug(ctx, articleSlug).Return(nil, domain.ErrNotFound)
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name:  "propagates error from ArticleRepository.FindBySlug and skips subsequent calls",
			input: GetArticleUsecaseInput{Slug: articleSlug},
			mock: func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository, sr *mock.MockSeriesRepository) {
				ar.EXPECT().FindBySlug(ctx, articleSlug).Return(nil, wantErr)
			},
			wantErr: wantErr,
		},
		{
			name:  "propagates error from TagRepository.FindByIDs",
			input: GetArticleUsecaseInput{Slug: articleSlug},
			mock: func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository, sr *mock.MockSeriesRepository) {
				ar.EXPECT().FindBySlug(ctx, articleSlug).Return(article, nil)
				tr.EXPECT().FindByIDs(ctx, tag1ID, tag2ID).Return(nil, wantErr)
			},
			wantErr: wantErr,
		},
		{
			name:  "propagates error from SeriesRepository.FindByArticleID when it is not ErrNotFound",
			input: GetArticleUsecaseInput{Slug: articleSlug},
			mock: func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository, sr *mock.MockSeriesRepository) {
				ar.EXPECT().FindBySlug(ctx, articleSlug).Return(article, nil)
				tr.EXPECT().FindByIDs(ctx, tag1ID, tag2ID).Return([]*domain.Tag{tag1, tag2}, nil)
				sr.EXPECT().FindByArticleID(ctx, currentID).Return(nil, wantErr)
			},
			wantErr: wantErr,
		},
		{
			name:  "propagates error from ArticleRepository.FindByIDs during neighbor fetch",
			input: GetArticleUsecaseInput{Slug: articleSlug},
			mock: func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository, sr *mock.MockSeriesRepository) {
				ar.EXPECT().FindBySlug(ctx, articleSlug).Return(article, nil)
				tr.EXPECT().FindByIDs(ctx, tag1ID, tag2ID).Return([]*domain.Tag{tag1, tag2}, nil)
				sr.EXPECT().FindByArticleID(ctx, currentID).Return(seriesMiddle, nil)
				ar.EXPECT().FindByIDs(ctx, prevID, nextID).Return(nil, wantErr)
			},
			wantErr: wantErr,
		},
		{
			name:  "hides draft article from public caller",
			input: GetArticleUsecaseInput{Slug: articleSlug},
			mock: func(ar *mock.MockArticleRepository, tr *mock.MockTagRepository, sr *mock.MockSeriesRepository) {
				draft := &domain.Article{
					ID: currentID, Slug: articleSlug, Title: "Draft", Body: "",
					Status: domain.ArticleStatusDraft,
				}
				ar.EXPECT().FindBySlug(ctx, articleSlug).Return(draft, nil)
			},
			wantErr: domain.ErrNotFound,
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
