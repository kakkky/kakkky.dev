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

func TestCreateSeriesUsecase_Exec(t *testing.T) {
	ctx := context.Background()

	var (
		existingTag1 domain.TagID = "11111111-1111-1111-1111-111111111111"
		existingTag2 domain.TagID = "22222222-2222-2222-2222-222222222222"
	)

	wantDBErr := errors.New("boom")

	tests := []struct {
		name    string
		input   CreateSeriesUsecaseInput
		mock    func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, tr *mock.MockTagRepository)
		wantErr error
	}{
		{
			name: "creates draft series with existing + new tags in single tx",
			input: CreateSeriesUsecaseInput{
				Title:          "Clean Arch",
				Description:    "learn clean arch",
				ExistingTagIDs: []domain.TagID{existingTag1, existingTag2},
				NewTagNames:    []string{"DDD"},
			},
			mock: func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewTagRepository().Return(tr)
				txRepo.EXPECT().NewSeriesRepository().Return(sr)

				tr.EXPECT().Store(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, tag *domain.Tag) error {
					assert.Equal(t, "DDD", tag.Name)
					tag.ID = "99999999-9999-9999-9999-999999999999"
					return nil
				})
				sr.EXPECT().Store(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, s *domain.Series) error {
					assert.Equal(t, "Clean Arch", s.Title)
					assert.Equal(t, "learn clean arch", s.Description)
					assert.Equal(t, domain.SeriesStatusDraft, s.Status)
					assert.ElementsMatch(t, []domain.TagID{existingTag1, existingTag2, "99999999-9999-9999-9999-999999999999"}, s.TagIDs)
					s.ID = "cccccccc-cccc-cccc-cccc-ccccccccccc1"
					return nil
				})
			},
		},
		{
			name: "translates tag ErrAlreadyExists to ErrInvalidArgument with user-facing message",
			input: CreateSeriesUsecaseInput{
				Title:       "Clean Arch",
				NewTagNames: []string{"Go"},
			},
			mock: func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewTagRepository().Return(tr)
				txRepo.EXPECT().NewSeriesRepository().Return(sr)

				tr.EXPECT().Store(ctx, gomock.Any()).Return(domain.ErrAlreadyExists)
			},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "translates series ErrAlreadyExists (slug conflict) to ErrInvalidArgument",
			input: CreateSeriesUsecaseInput{
				Title: "Clean Arch",
			},
			mock: func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewTagRepository().Return(tr)
				txRepo.EXPECT().NewSeriesRepository().Return(sr)

				sr.EXPECT().Store(ctx, gomock.Any()).Return(domain.ErrAlreadyExists)
			},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "propagates unexpected series Store error",
			input: CreateSeriesUsecaseInput{
				Title: "Clean Arch",
			},
			mock: func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(func(ctx context.Context, fn func(domain.Repository) error) error {
					return fn(txRepo)
				})
				txRepo.EXPECT().NewTagRepository().Return(tr)
				txRepo.EXPECT().NewSeriesRepository().Return(sr)

				sr.EXPECT().Store(ctx, gomock.Any()).Return(wantDBErr)
			},
			wantErr: wantDBErr,
		},
		{
			name: "rejects empty title before tx",
			input: CreateSeriesUsecaseInput{
				Title: "",
			},
			mock:    func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, tr *mock.MockTagRepository) {},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "rejects duplicated new tag names before tx",
			input: CreateSeriesUsecaseInput{
				Title:       "Clean Arch",
				NewTagNames: []string{"DDD", "DDD"},
			},
			mock:    func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, tr *mock.MockTagRepository) {},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "rejects empty new tag name before tx",
			input: CreateSeriesUsecaseInput{
				Title:       "Clean Arch",
				NewTagNames: []string{""},
			},
			mock:    func(repo, txRepo *mock.MockRepository, sr *mock.MockSeriesRepository, tr *mock.MockTagRepository) {},
			wantErr: domain.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			repo := mock.NewMockRepository(ctrl)
			txRepo := mock.NewMockRepository(ctrl)
			sr := mock.NewMockSeriesRepository(ctrl)
			tr := mock.NewMockTagRepository(ctrl)
			qs := mock.NewMockQueryService(ctrl)

			tt.mock(repo, txRepo, sr, tr)

			uc := NewUseCase(repo, qs).NewCreateSeriesUsecase()
			out, err := uc.Exec(ctx, tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, CreateSeriesUsecaseOutput{}, out)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, out.SeriesSlug)
		})
	}
}
