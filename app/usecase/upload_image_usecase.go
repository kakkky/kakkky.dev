package usecase

import (
	"context"
	"fmt"
	"io"

	"github.com/kakkky/kakkky.dev/domain"
)

const uploadImageMaxSize = 5 * 1024 * 1024

type UploadImageUsecase struct {
	s3Client domain.S3Client
}

func (us *UseCase) NewUploadImageUsecase() *UploadImageUsecase {
	return &UploadImageUsecase{s3Client: us.client.NewS3Client()}
}

type UploadImageUsecaseInput struct {
	ContentType string
	Size        int64
	Body        io.Reader
}

type UploadImageUsecaseOutput struct {
	URL string
}

func (us *UploadImageUsecase) Exec(ctx context.Context, in UploadImageUsecaseInput) (UploadImageUsecaseOutput, error) {
	if in.Size <= 0 {
		return UploadImageUsecaseOutput{}, domain.ErrInvalidArgument.With("空 の 画像 は アップロード できません")
	}
	if in.Size > uploadImageMaxSize {
		return UploadImageUsecaseOutput{}, domain.ErrInvalidArgument.With(
			fmt.Sprintf("画像 サイズ は %dMB 以下 に してください", uploadImageMaxSize/(1024*1024)),
		)
	}

	url, err := us.s3Client.Upload(ctx, in.ContentType, in.Body)
	if err != nil {
		return UploadImageUsecaseOutput{}, err
	}
	return UploadImageUsecaseOutput{URL: url}, nil
}
