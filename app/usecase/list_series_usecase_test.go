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

func TestListSeriesUsecase_Exec(t *testing.T) {
	ctx := context.Background()
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	var (
		id1 domain.SeriesID = "cccccccc-cccc-cccc-cccc-ccccccccccc1"
		id2 domain.SeriesID = "cccccccc-cccc-cccc-cccc-ccccccccccc2"
		id3 domain.SeriesID = "cccccccc-cccc-cccc-cccc-ccccccccccc3"
	)

	s1 := domain.Series{ID: id1, Title: "S1", CreatedAt: baseTime.Add(3 * time.Hour)}
	s2 := domain.Series{ID: id2, Title: "S2", CreatedAt: baseTime.Add(2 * time.Hour)}
	s3 := domain.Series{ID: id3, Title: "S3", CreatedAt: baseTime.Add(1 * time.Hour)}

	wantErr := errors.New("boom")

	tests := []struct {
		name    string
		input   ListSeriesUsecaseInput
		mock    func(repo *mock.MockSeriesRepository)
		want    ListSeriesUsecaseOutput
		wantErr error
	}{
		{
			name:  "truncates to limit and sets NextCursor when repo returns more than limit (limit+1 signal)",
			input: ListSeriesUsecaseInput{Limit: 2},
			mock: func(repo *mock.MockSeriesRepository) {
				repo.EXPECT().
					List(ctx, domain.SeriesID(""), time.Time{}, 3).
					Return([]*domain.Series{&s1, &s2, &s3}, nil)
			},
			want: ListSeriesUsecaseOutput{
				Series:     []domain.Series{s1, s2},
				NextCursor: ListSeriesUsecaseCursor{AfterID: s2.ID, AfterCreatedAt: s2.CreatedAt},
			},
		},
		{
			name:  "leaves NextCursor zero when items exactly match the limit",
			input: ListSeriesUsecaseInput{Limit: 2},
			mock: func(repo *mock.MockSeriesRepository) {
				repo.EXPECT().
					List(ctx, domain.SeriesID(""), time.Time{}, 3).
					Return([]*domain.Series{&s1, &s2}, nil)
			},
			want: ListSeriesUsecaseOutput{Series: []domain.Series{s1, s2}},
		},
		{
			name:  "leaves NextCursor zero when items are fewer than the limit",
			input: ListSeriesUsecaseInput{Limit: 10},
			mock: func(repo *mock.MockSeriesRepository) {
				repo.EXPECT().
					List(ctx, domain.SeriesID(""), time.Time{}, 11).
					Return([]*domain.Series{&s1}, nil)
			},
			want: ListSeriesUsecaseOutput{Series: []domain.Series{s1}},
		},
		{
			name: "passes cursor through to repo.List",
			input: ListSeriesUsecaseInput{
				Cursor: ListSeriesUsecaseCursor{AfterID: id1, AfterCreatedAt: baseTime.Add(3 * time.Hour)},
				Limit:  10,
			},
			mock: func(repo *mock.MockSeriesRepository) {
				repo.EXPECT().
					List(ctx, id1, baseTime.Add(3*time.Hour), 11).
					Return([]*domain.Series{&s2}, nil)
			},
			want: ListSeriesUsecaseOutput{Series: []domain.Series{s2}},
		},
		{
			name:  "propagates error from repo.List",
			input: ListSeriesUsecaseInput{Limit: 10},
			mock: func(repo *mock.MockSeriesRepository) {
				repo.EXPECT().
					List(ctx, domain.SeriesID(""), time.Time{}, 11).
					Return(nil, wantErr)
			},
			wantErr: wantErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			sr := mock.NewMockSeriesRepository(ctrl)
			repo := mock.NewMockRepository(ctrl)
			repo.EXPECT().NewSeriesRepository().Return(sr)
			qs := mock.NewMockQueryService(ctrl)

			tt.mock(sr)

			gs := NewUseCase(repo, qs).NewListSeriesUsecase()
			got, err := gs.Exec(ctx, tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, ListSeriesUsecaseOutput{}, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
