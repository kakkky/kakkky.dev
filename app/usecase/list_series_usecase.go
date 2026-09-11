package usecase

import (
	"context"
	"time"

	"github.com/kakkky/kakkky.dev/domain"
)

type ListSeriesUsecase struct {
	seriesRepo domain.SeriesRepository
}

func (us *UseCase) NewListSeriesUsecase() *ListSeriesUsecase {
	return &ListSeriesUsecase{
		seriesRepo: us.repo.NewSeriesRepository(),
	}
}

type ListSeriesUsecaseCursor struct {
	AfterID        domain.SeriesID
	AfterCreatedAt time.Time
}

type ListSeriesUsecaseInput struct {
	Cursor ListSeriesUsecaseCursor
	Limit  int
}

type ListSeriesUsecaseOutput struct {
	Series     []domain.Series
	NextCursor ListSeriesUsecaseCursor
}

func (us *ListSeriesUsecase) Exec(ctx context.Context, in ListSeriesUsecaseInput) (ListSeriesUsecaseOutput, error) {
	fetchLimit := in.Limit
	if in.Limit > 0 {
		fetchLimit = in.Limit + 1
	}

	rows, err := us.seriesRepo.List(ctx, in.Cursor.AfterID, in.Cursor.AfterCreatedAt, fetchLimit)
	if err != nil {
		return ListSeriesUsecaseOutput{}, err
	}

	series := make([]domain.Series, len(rows))
	for i, s := range rows {
		series[i] = *s
	}

	var next ListSeriesUsecaseCursor
	if in.Limit > 0 && len(series) > in.Limit {
		series = series[:in.Limit]
		last := series[len(series)-1]
		next = ListSeriesUsecaseCursor{
			AfterID:        last.ID,
			AfterCreatedAt: last.CreatedAt,
		}
	}
	return ListSeriesUsecaseOutput{Series: series, NextCursor: next}, nil
}
