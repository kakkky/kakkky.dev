package usecase

import "github.com/kakkky/kakkky.dev/domain"

type UseCase struct {
	repo domain.Repository
	qs   domain.QueryService
}

func NewUseCase(repo domain.Repository, qs domain.QueryService) *UseCase {
	return &UseCase{
		repo: repo,
		qs:   qs,
	}
}
