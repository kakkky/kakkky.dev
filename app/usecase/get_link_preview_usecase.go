package usecase

import (
	"context"
	"time"

	"github.com/kakkky/kakkky.dev/domain"
)

const linkPreviewCacheTTL = 6 * time.Hour

type GetLinkPreviewUsecase struct {
	ogpFetcher domain.OGPFetcher
	ttl        time.Duration
	cache      domain.CacheClient
}

func (us *UseCase) NewGetLinkPreviewUsecase() *GetLinkPreviewUsecase {
	return &GetLinkPreviewUsecase{
		ogpFetcher: us.client.NewOGPFetcher(),
		ttl:        linkPreviewCacheTTL,
		cache:      us.cache.NewInMemoryCacheClient(),
	}
}

type GetLinkPreviewUsecaseInput struct {
	URL string
}

type GetLinkPreviewUsecaseOutput struct {
	Data domain.OGPData
}

func (us *GetLinkPreviewUsecase) Exec(ctx context.Context, in GetLinkPreviewUsecaseInput) (GetLinkPreviewUsecaseOutput, error) {
	if v, ok := us.cache.Get(in.URL); ok {
		return GetLinkPreviewUsecaseOutput{Data: v.(domain.OGPData)}, nil
	}
	d, err := us.ogpFetcher.Fetch(ctx, in.URL)
	if err != nil {
		return GetLinkPreviewUsecaseOutput{}, err
	}
	us.cache.Set(in.URL, d, time.Now().Add(us.ttl))
	return GetLinkPreviewUsecaseOutput{Data: d}, nil
}
