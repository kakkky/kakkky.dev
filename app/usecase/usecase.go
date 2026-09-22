package usecase

import "github.com/kakkky/kakkky.dev/domain"

type UseCase struct {
	repo   domain.Repository
	qs     domain.QueryService
	client domain.Client
	cache  domain.Cache
}

func NewUseCase(repo domain.Repository, qs domain.QueryService, client domain.Client, cache domain.Cache) *UseCase {
	return &UseCase{
		repo:   repo,
		qs:     qs,
		client: client,
		cache:  cache,
	}
}
