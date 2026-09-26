//go:build integration

package testhelper

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
)

type LocalStack struct {
	Endpoint     string
	Region       string
	AccessKey    string
	SecretKey    string
	Bucket       string
	UsePathStyle bool
}

func SetupLocalStack(ctx context.Context, bucket string) (*LocalStack, func(), error) {
	ls, err := localstack.Run(ctx, "localstack/localstack:3")
	if err != nil {
		return nil, nil, fmt.Errorf("run localstack: %w", err)
	}

	mapped, err := ls.PortEndpoint(ctx, "4566/tcp", "http")
	if err != nil {
		_ = ls.Terminate(ctx)
		return nil, nil, fmt.Errorf("port endpoint: %w", err)
	}

	l := &LocalStack{
		Endpoint:     mapped,
		Region:       "us-east-1",
		AccessKey:    "test",
		SecretKey:    "test",
		Bucket:       bucket,
		UsePathStyle: true,
	}
	if _, err := l.newS3Client().CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)}); err != nil {
		_ = ls.Terminate(ctx)
		return nil, nil, fmt.Errorf("create bucket: %w", err)
	}

	return l, func() { _ = ls.Terminate(ctx) }, nil
}

// HeadObject は bucket 内 の object 存在確認 と metadata 取得 に使う helper。
func (l *LocalStack) HeadObject(ctx context.Context, key string) (*s3.HeadObjectOutput, error) {
	c := l.newS3Client()
	return c.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(l.Bucket),
		Key:    aws.String(key),
	})
}

func (l *LocalStack) newS3Client() *s3.Client {
	return s3.NewFromConfig(aws.Config{
		Region:      l.Region,
		Credentials: credentials.NewStaticCredentialsProvider(l.AccessKey, l.SecretKey, ""),
	}, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(l.Endpoint)
		o.UsePathStyle = l.UsePathStyle
	})
}
