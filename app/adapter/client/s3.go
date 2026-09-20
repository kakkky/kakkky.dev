package client

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"github.com/kakkky/kakkky.dev/domain"
)

var s3ExtByContentType = map[string]string{
	"image/png":  "png",
	"image/jpeg": "jpg",
	"image/webp": "webp",
	"image/gif":  "gif",
}

type S3Client struct {
	client       *s3.Client
	bucket       string
	imageBaseURL string
}

func (c *Client) NewS3Client() domain.S3Client {
	awsCfg := aws.Config{
		Region: c.cfg.S3Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			c.cfg.S3AccessKeyID,
			c.cfg.S3SecretAccessKey,
			"",
		),
	}

	s3c := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(c.cfg.S3Endpoint)
		o.UsePathStyle = c.cfg.S3UsePathStyle
	})

	return &S3Client{
		client:       s3c,
		bucket:       c.cfg.S3Bucket,
		imageBaseURL: strings.TrimRight(c.cfg.ImageBaseURL, "/"),
	}
}

func (c *S3Client) Upload(ctx context.Context, prefix, contentType string, body io.Reader) (string, error) {
	ext, ok := s3ExtByContentType[contentType]
	if !ok {
		return "", domain.ErrInvalidArgument.With("png / jpeg / webp / gif のみ アップロード できます")
	}

	key := fmt.Sprintf("%s/%s.%s", strings.Trim(prefix, "/"), uuid.NewString(), ext)

	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		Body:        body,
	})
	if err != nil {
		return "", domain.ErrInternal.Wrap(err, "画像 の アップロード に 失敗 しました")
	}
	return c.imageBaseURL + "/" + key, nil
}
