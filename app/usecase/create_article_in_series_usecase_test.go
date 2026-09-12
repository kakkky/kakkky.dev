package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/testhelper/mock"
)

func TestCreateArticleInSeriesUsecase_Exec(t *testing.T) {
	ctx := context.Background()

	var (
		seriesID     domain.SeriesID  = "cccccccc-cccc-cccc-cccc-ccccccccccc1"
		newArticleID domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1"
	)

	wantDBErr := errors.New("boom")

	tests := []struct {
		name         string
		input        CreateArticleInSeriesUsecaseInput
		mock         func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository)
		wantPosition int
		wantErr      error
	}{
		{
			name: "appends new article at position 1 for empty series",
			input: CreateArticleInSeriesUsecaseInput{
				SeriesSlug: "clean-arch",
				Title:      "First",
			},
			mock: func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewSeriesRepository().Return(sr)
				txRepo.EXPECT().NewArticleRepository().Return(ar)

				sr.EXPECT().FindBySlug(ctx, domain.Slug("clean-arch")).Return(&domain.Series{
					ID: seriesID, Slug: "clean-arch", Title: "S", Status: domain.SeriesStatusDraft,
				}, nil)
				ar.EXPECT().Store(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, a *domain.Article) error {
					assert.Equal(t, "First", a.Title)
					assert.Equal(t, domain.ArticleStatusDraft, a.Status)
					a.ID = newArticleID
					return nil
				})
				sr.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, s *domain.Series) error {
					assert.Equal(t, []domain.SeriesArticle{
						{ArticleID: newArticleID, Position: 1},
					}, s.Articles)
					return nil
				})
			},
			wantPosition: 1,
		},
		{
			name: "appends at max position + 1 (skips gaps)",
			input: CreateArticleInSeriesUsecaseInput{
				SeriesSlug: "clean-arch",
				Title:      "Next",
			},
			mock: func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewSeriesRepository().Return(sr)
				txRepo.EXPECT().NewArticleRepository().Return(ar)

				sr.EXPECT().FindBySlug(ctx, domain.Slug("clean-arch")).Return(&domain.Series{
					ID: seriesID, Slug: "clean-arch", Title: "S", Status: domain.SeriesStatusDraft,
					Articles: []domain.SeriesArticle{
						{ArticleID: "a1", Position: 1},
						{ArticleID: "a2", Position: 5},
					},
				}, nil)
				ar.EXPECT().Store(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, a *domain.Article) error {
					a.ID = newArticleID
					return nil
				})
				sr.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, s *domain.Series) error {
					assert.Equal(t, 6, s.Articles[len(s.Articles)-1].Position)
					assert.Equal(t, newArticleID, s.Articles[len(s.Articles)-1].ArticleID)
					return nil
				})
			},
			wantPosition: 6,
		},
		{
			name: "returns not found when series slug does not exist",
			input: CreateArticleInSeriesUsecaseInput{
				SeriesSlug: "missing",
				Title:      "T",
			},
			mock: func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewSeriesRepository().Return(sr)
				txRepo.EXPECT().NewArticleRepository().Return(ar)

				sr.EXPECT().FindBySlug(ctx, domain.Slug("missing")).Return(nil, domain.ErrNotFound)
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "translates slug conflict to ErrInvalidArgument",
			input: CreateArticleInSeriesUsecaseInput{
				SeriesSlug: "clean-arch",
				Title:      "Duplicated",
			},
			mock: func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewSeriesRepository().Return(sr)
				txRepo.EXPECT().NewArticleRepository().Return(ar)

				sr.EXPECT().FindBySlug(ctx, domain.Slug("clean-arch")).Return(&domain.Series{
					ID: seriesID, Slug: "clean-arch", Title: "S", Status: domain.SeriesStatusDraft,
				}, nil)
				ar.EXPECT().Store(ctx, gomock.Any()).Return(domain.ErrAlreadyExists)
			},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "propagates unexpected series Update error",
			input: CreateArticleInSeriesUsecaseInput{
				SeriesSlug: "clean-arch",
				Title:      "T",
			},
			mock: func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewSeriesRepository().Return(sr)
				txRepo.EXPECT().NewArticleRepository().Return(ar)

				sr.EXPECT().FindBySlug(ctx, domain.Slug("clean-arch")).Return(&domain.Series{
					ID: seriesID, Slug: "clean-arch", Title: "S", Status: domain.SeriesStatusDraft,
				}, nil)
				ar.EXPECT().Store(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, a *domain.Article) error {
					a.ID = newArticleID
					return nil
				})
				sr.EXPECT().Update(ctx, gomock.Any()).Return(wantDBErr)
			},
			wantErr: wantDBErr,
		},
		{
			name: "rejects empty series slug before tx",
			input: CreateArticleInSeriesUsecaseInput{
				SeriesSlug: "",
				Title:      "T",
			},
			mock:    func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository) {},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "rejects empty title before tx",
			input: CreateArticleInSeriesUsecaseInput{
				SeriesSlug: "clean-arch",
				Title:      "",
			},
			mock:    func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository) {},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "rejects too long title before tx",
			input: CreateArticleInSeriesUsecaseInput{
				SeriesSlug: "clean-arch",
				Title:      strings.Repeat("a", domain.ArticleTitleMaxLength+1),
			},
			mock:    func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository) {},
			wantErr: domain.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			repo := mock.NewMockRepository(ctrl)
			txRepo := mock.NewMockRepository(ctrl)
			sr := mock.NewMockSeriesRepository(ctrl)
			ar := mock.NewMockArticleRepository(ctrl)
			qs := mock.NewMockQueryService(ctrl)

			tt.mock(repo, txRepo, sr, ar)

			uc := NewUseCase(repo, qs).NewCreateArticleInSeriesUsecase()
			out, err := uc.Exec(ctx, tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, CreateArticleInSeriesUsecaseOutput{}, out)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, newArticleID, out.ArticleID)
			assert.Equal(t, tt.input.Title, out.ArticleTitle)
			assert.Equal(t, tt.wantPosition, out.Position)
		})
	}
}
