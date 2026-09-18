package usecase

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"

	"github.com/kakkky/kakkky.dev/domain"
)

const (
	uploadImageMaxSize = 5 * 1024 * 1024
	uploadImageKeyDir  = "articles"
)

var uploadImageAllowedTypes = map[string]string{
	"image/png":  "png",
	"image/jpeg": "jpg",
	"image/webp": "webp",
	"image/gif":  "gif",
}

type UploadImageUsecase struct {
	uploader domain.ImageUploader
}

func (us *UseCase) NewUploadImageUsecase() *UploadImageUsecase {
	return &UploadImageUsecase{uploader: us.client.NewImageUploader()}
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
	ext, ok := uploadImageAllowedTypes[in.ContentType]
	if !ok {
		return UploadImageUsecaseOutput{}, domain.ErrInvalidArgument.With(
			"png / jpeg / webp / gif のみ アップロード できます",
		)
	}
	if in.Size <= 0 {
		return UploadImageUsecaseOutput{}, domain.ErrInvalidArgument.With("空 の 画像 は アップロード できません")
	}
	if in.Size > uploadImageMaxSize {
		return UploadImageUsecaseOutput{}, domain.ErrInvalidArgument.With(
			fmt.Sprintf("画像 サイズ は %dMB 以下 に してください", uploadImageMaxSize/(1024*1024)),
		)
	}

	key := fmt.Sprintf("%s/%s.%s", uploadImageKeyDir, uuid.NewString(), ext)

	url, err := us.uploader.Upload(ctx, key, in.ContentType, in.Body)
	if err != nil {
		return UploadImageUsecaseOutput{}, err
	}
	return UploadImageUsecaseOutput{URL: url}, nil
}
