package usecase

import "context"

type GetDashboardUsecase struct{}

func (us *UseCase) NewGetDashboardUsecase() *GetDashboardUsecase {
	return &GetDashboardUsecase{}
}

type GetDashboardUsecaseInput struct{}

type GetDashboardUsecaseOutput struct{}

func (us *GetDashboardUsecase) Exec(_ context.Context, _ GetDashboardUsecaseInput) (GetDashboardUsecaseOutput, error) {
	return GetDashboardUsecaseOutput{}, nil
}
