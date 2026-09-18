package client

import (
	"context"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/kakkky/kakkky.dev/config"
	"github.com/kakkky/kakkky.dev/domain"
)

type S3ImageUploader struct {
	client       *s3.Client
	bucket       string
	imageBaseURL string
}

func NewS3ImageUploader(cfg *config.Config) *S3ImageUploader {
	awsCfg := aws.Config{
		Region: cfg.S3Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.S3AccessKeyID,
			cfg.S3SecretAccessKey,
			"",
		),
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.S3Endpoint)
		o.UsePathStyle = cfg.S3UsePathStyle
	})

	return &S3ImageUploader{
		client:       client,
		bucket:       cfg.S3Bucket,
		imageBaseURL: strings.TrimRight(cfg.ImageBaseURL, "/"),
	}
}

func (u *S3ImageUploader) Upload(ctx context.Context, key, contentType string, body io.Reader) (string, error) {
	_, err := u.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(u.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		Body:        body,
	})
	if err != nil {
		return "", domain.ErrInternal.Wrap(err, "画像 の アップロード に 失敗 しました")
	}
	return u.imageBaseURL + "/" + key, nil
}
