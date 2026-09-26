//go:build integration

package client

import (
	"context"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kakkky/kakkky.dev/config"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/testhelper"
)

func TestS3Client_Upload(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	const bucket = "test-bucket"
	ls, cleanup, err := testhelper.SetupLocalStack(ctx, bucket)
	require.NoError(t, err)
	t.Cleanup(cleanup)

	imageBaseURL := ls.Endpoint + "/" + bucket
	c := (&Client{cfg: &config.Config{
		S3Endpoint:        ls.Endpoint,
		S3Region:          ls.Region,
		S3Bucket:          ls.Bucket,
		S3AccessKeyID:     ls.AccessKey,
		S3SecretAccessKey: ls.SecretKey,
		S3UsePathStyle:    ls.UsePathStyle,
		ImageBaseURL:      imageBaseURL,
	}}).NewS3Client()

	tests := []struct {
		name        string
		contentType string
		body        string
		wantExt     string
		wantErr     error
	}{
		{
			name:        "success: png",
			contentType: "image/png",
			body:        "pngbody",
			wantExt:     "png",
		},
		{
			name:        "success: jpeg",
			contentType: "image/jpeg",
			body:        "jpgbody",
			wantExt:     "jpg",
		},
		{
			name:        "success: webp",
			contentType: "image/webp",
			body:        "webpbody",
			wantExt:     "webp",
		},
		{
			name:        "success: gif",
			contentType: "image/gif",
			body:        "gifbody",
			wantExt:     "gif",
		},
		{
			name:        "error: rejects unsupported content-type",
			contentType: "image/svg+xml",
			body:        "svg",
			wantErr:     domain.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			url, err := c.Upload(ctx, tt.contentType, strings.NewReader(tt.body))

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, url)
				return
			}
			require.NoError(t, err)

			// URL prefix と ext が想定通り
			assert.True(t, strings.HasPrefix(url, imageBaseURL+"/"), "url=%s", url)
			assert.Equal(t, "."+tt.wantExt, path.Ext(url))

			// object が実際に bucket に存在し content-type / body が保存されている
			key := strings.TrimPrefix(url, imageBaseURL+"/")
			head, err := ls.HeadObject(ctx, key)
			require.NoError(t, err)
			require.NotNil(t, head.ContentType)
			assert.Equal(t, tt.contentType, *head.ContentType)
			assert.Equal(t, int64(len(tt.body)), *head.ContentLength)
		})
	}

}
