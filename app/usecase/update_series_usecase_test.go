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

func TestUpdateSeriesUsecase_Exec(t *testing.T) {
	ctx := context.Background()

	var (
		seriesID     domain.SeriesID  = "cccccccc-cccc-cccc-cccc-ccccccccccc1"
		existingTag1 domain.TagID     = "11111111-1111-1111-1111-111111111111"
		existingTag2 domain.TagID     = "22222222-2222-2222-2222-222222222222"
		newTagID     domain.TagID     = "99999999-9999-9999-9999-999999999999"
		article1     domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1"
		article2     domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa2"
		newArticleID domain.ArticleID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa9"
	)

	wantDBErr := errors.New("boom")

	tests := []struct {
		name    string
		input   UpdateSeriesUsecaseInput
		mock    func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository)
		wantErr error
	}{
		{
			name: "updates fields and replaces tags in single tx",
			input: UpdateSeriesUsecaseInput{
				Slug:           "clean-arch",
				Title:          "New Title",
				Description:    "new desc",
				Status:         domain.SeriesStatusPublishedOngoing,
				ExistingTagIDs: []domain.TagID{existingTag2},
				NewTagNames:    []string{"DDD"},
			},
			mock: func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewSeriesRepository().Return(sr)
				txRepo.EXPECT().NewArticleRepository().Return(ar)
				txRepo.EXPECT().NewTagRepository().Return(tr)

				sr.EXPECT().FindBySlug(ctx, domain.Slug("clean-arch")).Return(&domain.Series{
					ID: seriesID, Slug: "clean-arch", Title: "Old", Description: "old",
					Status: domain.SeriesStatusDraft,
					TagIDs: []domain.TagID{existingTag1},
				}, nil)
				tr.EXPECT().Store(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, tag *domain.Tag) error {
					assert.Equal(t, "DDD", tag.Name)
					tag.ID = newTagID
					return nil
				})
				sr.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, s *domain.Series) error {
					assert.Equal(t, seriesID, s.ID)
					assert.Equal(t, "New Title", s.Title)
					assert.Equal(t, "new desc", s.Description)
					assert.Equal(t, domain.SeriesStatusPublishedOngoing, s.Status)
					assert.ElementsMatch(t, []domain.TagID{existingTag2, newTagID}, s.TagIDs)
					assert.False(t, s.PublishedAt.IsZero(), "draft→published transition should set PublishedAt")
					return nil
				})
			},
		},
		{
			name: "reorders existing articles by OrderedArticleIDs order",
			input: UpdateSeriesUsecaseInput{
				Slug:              "clean-arch",
				Title:             "T",
				Status:            domain.SeriesStatusDraft,
				OrderedArticleIDs: []domain.ArticleID{article2, article1},
			},
			mock: func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewSeriesRepository().Return(sr)
				txRepo.EXPECT().NewArticleRepository().Return(ar)
				txRepo.EXPECT().NewTagRepository().Return(tr)

				sr.EXPECT().FindBySlug(ctx, domain.Slug("clean-arch")).Return(&domain.Series{
					ID: seriesID, Slug: "clean-arch", Title: "Old",
					Status: domain.SeriesStatusDraft,
					Articles: []domain.SeriesArticle{
						{ArticleID: article1, Position: 1},
						{ArticleID: article2, Position: 2},
					},
				}, nil)
				sr.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, s *domain.Series) error {
					assert.Equal(t, []domain.SeriesArticle{
						{ArticleID: article2, Position: 1},
						{ArticleID: article1, Position: 2},
					}, s.Articles)
					return nil
				})
			},
		},
		{
			name: "deletes articles listed in DeleteArticleIDs",
			input: UpdateSeriesUsecaseInput{
				Slug:              "clean-arch",
				Title:             "T",
				Status:            domain.SeriesStatusDraft,
				OrderedArticleIDs: []domain.ArticleID{article1},
				DeleteArticleIDs:  []domain.ArticleID{article2},
			},
			mock: func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewSeriesRepository().Return(sr)
				txRepo.EXPECT().NewArticleRepository().Return(ar)
				txRepo.EXPECT().NewTagRepository().Return(tr)

				sr.EXPECT().FindBySlug(ctx, domain.Slug("clean-arch")).Return(&domain.Series{
					ID: seriesID, Slug: "clean-arch", Title: "T",
					Status: domain.SeriesStatusDraft,
					Articles: []domain.SeriesArticle{
						{ArticleID: article1, Position: 1},
						{ArticleID: article2, Position: 2},
					},
				}, nil)
				ar.EXPECT().Delete(ctx, article2).Return(nil)
				sr.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, s *domain.Series) error {
					assert.Equal(t, []domain.SeriesArticle{
						{ArticleID: article1, Position: 1},
					}, s.Articles)
					return nil
				})
			},
		},
		{
			name: "creates new articles from NewArticleTitles and appends at tail",
			input: UpdateSeriesUsecaseInput{
				Slug:              "clean-arch",
				Title:             "T",
				Status:            domain.SeriesStatusDraft,
				OrderedArticleIDs: []domain.ArticleID{article1},
				NewArticleTitles:  []string{"New Chapter"},
			},
			mock: func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewSeriesRepository().Return(sr)
				txRepo.EXPECT().NewArticleRepository().Return(ar)
				txRepo.EXPECT().NewTagRepository().Return(tr)

				sr.EXPECT().FindBySlug(ctx, domain.Slug("clean-arch")).Return(&domain.Series{
					ID: seriesID, Slug: "clean-arch", Title: "T",
					Status: domain.SeriesStatusDraft,
					Articles: []domain.SeriesArticle{
						{ArticleID: article1, Position: 1},
					},
				}, nil)
				ar.EXPECT().Store(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, a *domain.Article) error {
					assert.Equal(t, "New Chapter", a.Title)
					assert.Equal(t, domain.ArticleStatusDraft, a.Status)
					a.ID = newArticleID
					return nil
				})
				sr.EXPECT().Update(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, s *domain.Series) error {
					assert.Equal(t, []domain.SeriesArticle{
						{ArticleID: article1, Position: 1},
						{ArticleID: newArticleID, Position: 2},
					}, s.Articles)
					return nil
				})
			},
		},
		{
			name: "translates new article slug conflict to ErrInvalidArgument",
			input: UpdateSeriesUsecaseInput{
				Slug:             "clean-arch",
				Title:            "T",
				Status:           domain.SeriesStatusDraft,
				NewArticleTitles: []string{"Duplicated"},
			},
			mock: func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewSeriesRepository().Return(sr)
				txRepo.EXPECT().NewArticleRepository().Return(ar)
				txRepo.EXPECT().NewTagRepository().Return(tr)

				sr.EXPECT().FindBySlug(ctx, domain.Slug("clean-arch")).Return(&domain.Series{
					ID: seriesID, Slug: "clean-arch", Title: "T", Status: domain.SeriesStatusDraft,
				}, nil)
				ar.EXPECT().Store(ctx, gomock.Any()).Return(domain.ErrAlreadyExists)
			},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "returns not found when slug does not exist",
			input: UpdateSeriesUsecaseInput{
				Slug:   "missing",
				Title:  "New",
				Status: domain.SeriesStatusDraft,
			},
			mock: func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewSeriesRepository().Return(sr)
				txRepo.EXPECT().NewArticleRepository().Return(ar)
				txRepo.EXPECT().NewTagRepository().Return(tr)

				sr.EXPECT().FindBySlug(ctx, domain.Slug("missing")).Return(nil, domain.ErrNotFound)
			},
			wantErr: domain.ErrNotFound,
		},
		{
			name: "translates tag ErrAlreadyExists to ErrInvalidArgument",
			input: UpdateSeriesUsecaseInput{
				Slug:        "clean-arch",
				Title:       "New",
				Status:      domain.SeriesStatusDraft,
				NewTagNames: []string{"Go"},
			},
			mock: func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewSeriesRepository().Return(sr)
				txRepo.EXPECT().NewArticleRepository().Return(ar)
				txRepo.EXPECT().NewTagRepository().Return(tr)

				sr.EXPECT().FindBySlug(ctx, domain.Slug("clean-arch")).Return(&domain.Series{
					ID: seriesID, Slug: "clean-arch", Title: "Old",
					Status: domain.SeriesStatusDraft,
				}, nil)
				tr.EXPECT().Store(ctx, gomock.Any()).Return(domain.ErrAlreadyExists)
			},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "propagates unexpected series Update error",
			input: UpdateSeriesUsecaseInput{
				Slug:   "clean-arch",
				Title:  "New",
				Status: domain.SeriesStatusDraft,
			},
			mock: func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewSeriesRepository().Return(sr)
				txRepo.EXPECT().NewArticleRepository().Return(ar)
				txRepo.EXPECT().NewTagRepository().Return(tr)

				sr.EXPECT().FindBySlug(ctx, domain.Slug("clean-arch")).Return(&domain.Series{
					ID: seriesID, Slug: "clean-arch", Title: "Old",
					Status: domain.SeriesStatusDraft,
				}, nil)
				sr.EXPECT().Update(ctx, gomock.Any()).Return(wantDBErr)
			},
			wantErr: wantDBErr,
		},
		{
			name: "rejects empty slug before tx",
			input: UpdateSeriesUsecaseInput{
				Slug:   "",
				Title:  "New",
				Status: domain.SeriesStatusDraft,
			},
			mock:    func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "rejects duplicated new tag names before tx",
			input: UpdateSeriesUsecaseInput{
				Slug:        "clean-arch",
				Title:       "New",
				Status:      domain.SeriesStatusDraft,
				NewTagNames: []string{"DDD", "DDD"},
			},
			mock:    func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "rejects empty new tag name before tx",
			input: UpdateSeriesUsecaseInput{
				Slug:        "clean-arch",
				Title:       "New",
				Status:      domain.SeriesStatusDraft,
				NewTagNames: []string{""},
			},
			mock:    func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "rejects empty new article title before tx",
			input: UpdateSeriesUsecaseInput{
				Slug:             "clean-arch",
				Title:            "T",
				Status:           domain.SeriesStatusDraft,
				NewArticleTitles: []string{""},
			},
			mock:    func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "propagates invalid title validation from domain",
			input: UpdateSeriesUsecaseInput{
				Slug:   "clean-arch",
				Title:  "",
				Status: domain.SeriesStatusDraft,
			},
			mock: func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewSeriesRepository().Return(sr)
				txRepo.EXPECT().NewArticleRepository().Return(ar)
				txRepo.EXPECT().NewTagRepository().Return(tr)

				sr.EXPECT().FindBySlug(ctx, domain.Slug("clean-arch")).Return(&domain.Series{
					ID: seriesID, Slug: "clean-arch", Title: "Old",
					Status: domain.SeriesStatusDraft,
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
			sr := mock.NewMockSeriesRepository(ctrl)
			ar := mock.NewMockArticleRepository(ctrl)
			tr := mock.NewMockTagRepository(ctrl)
			qs := mock.NewMockQueryService(ctrl)

			tt.mock(repo, txRepo, sr, ar, tr)

			uc := NewUseCase(repo, qs).NewUpdateSeriesUsecase()
			out, err := uc.Exec(ctx, tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, UpdateSeriesUsecaseOutput{}, out)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.input.Slug, out.SeriesSlug)
		})
	}
}
