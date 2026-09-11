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

func TestListArticlesUsecase_Exec(t *testing.T) {
	ctx := context.Background()
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	var (
		id1 domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1"
		id2 domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2"
		id3 domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa3"
	)

	a1 := domain.Article{ID: id1, Title: "A1", CreatedAt: baseTime.Add(3 * time.Hour)}
	a2 := domain.Article{ID: id2, Title: "A2", CreatedAt: baseTime.Add(2 * time.Hour)}
	a3 := domain.Article{ID: id3, Title: "A3", CreatedAt: baseTime.Add(1 * time.Hour)}

	wantErr := errors.New("boom")

	tests := []struct {
		name    string
		input   ListArticlesUsecaseInput
		mock    func(repo *mock.MockArticleRepository)
		want    ListArticlesUsecaseOutput
		wantErr error
	}{
		{
			name:  "truncates to limit and sets NextCursor when repo returns more than limit (limit+1 signal)",
			input: ListArticlesUsecaseInput{Limit: 2},
			mock: func(repo *mock.MockArticleRepository) {
				repo.EXPECT().
					List(ctx, domain.ArticleID(""), time.Time{}, 3).
					Return([]*domain.Article{&a1, &a2, &a3}, nil)
			},
			want: ListArticlesUsecaseOutput{
				Articles:   []domain.Article{a1, a2},
				NextCursor: ListArticlesUsecaseCursor{AfterID: a2.ID, AfterCreatedAt: a2.CreatedAt},
			},
		},
		{
			name:  "leaves NextCursor zero when items exactly match the limit",
			input: ListArticlesUsecaseInput{Limit: 2},
			mock: func(repo *mock.MockArticleRepository) {
				repo.EXPECT().
					List(ctx, domain.ArticleID(""), time.Time{}, 3).
					Return([]*domain.Article{&a1, &a2}, nil)
			},
			want: ListArticlesUsecaseOutput{Articles: []domain.Article{a1, a2}},
		},
		{
			name:  "leaves NextCursor zero when items are fewer than the limit",
			input: ListArticlesUsecaseInput{Limit: 10},
			mock: func(repo *mock.MockArticleRepository) {
				repo.EXPECT().
					List(ctx, domain.ArticleID(""), time.Time{}, 11).
					Return([]*domain.Article{&a1}, nil)
			},
			want: ListArticlesUsecaseOutput{Articles: []domain.Article{a1}},
		},
		{
			name: "passes cursor through to repo.List",
			input: ListArticlesUsecaseInput{
				Cursor: ListArticlesUsecaseCursor{AfterID: id1, AfterCreatedAt: baseTime.Add(3 * time.Hour)},
				Limit:  10,
			},
			mock: func(repo *mock.MockArticleRepository) {
				repo.EXPECT().
					List(ctx, id1, baseTime.Add(3*time.Hour), 11).
					Return([]*domain.Article{&a2}, nil)
			},
			want: ListArticlesUsecaseOutput{Articles: []domain.Article{a2}},
		},
		{
			name:  "propagates error from repo.List",
			input: ListArticlesUsecaseInput{Limit: 10},
			mock: func(repo *mock.MockArticleRepository) {
				repo.EXPECT().
					List(ctx, domain.ArticleID(""), time.Time{}, 11).
					Return(nil, wantErr)
			},
			wantErr: wantErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			ar := mock.NewMockArticleRepository(ctrl)
			repo := mock.NewMockRepository(ctrl)
			repo.EXPECT().NewArticleRepository().Return(ar)
			qs := mock.NewMockQueryService(ctrl)

			tt.mock(ar)

			ga := NewUseCase(repo, qs).NewListArticlesUsecase()
			got, err := ga.Exec(ctx, tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, ListArticlesUsecaseOutput{}, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
