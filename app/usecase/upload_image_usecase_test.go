package usecase

import (
	"context"
	"github.com/kakkky/kakkky.dev/errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/testhelper/mock"
)

func TestUploadImageUsecase_Exec(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	uploadedURL := "https://images.example.com/abc.png"
	wantUploadErr := errors.New("boom")

	tests := []struct {
		name    string
		input   UploadImageUsecaseInput
		mock    func(c *mock.MockClient, s *mock.MockS3Client)
		wantURL string
		wantErr error
	}{
		{
			name: "success: uploads image and returns URL",
			input: UploadImageUsecaseInput{
				ContentType: "image/png",
				Size:        1024,
				Body:        strings.NewReader("dummy"),
			},
			mock: func(c *mock.MockClient, s *mock.MockS3Client) {
				c.EXPECT().NewS3Client().Return(s)
				s.EXPECT().Upload(ctx, "image/png", gomock.Any()).Return(uploadedURL, nil)
			},
			wantURL: uploadedURL,
		},
		{
			name: "error: rejects zero-size body",
			input: UploadImageUsecaseInput{
				ContentType: "image/png",
				Size:        0,
				Body:        strings.NewReader(""),
			},
			mock: func(c *mock.MockClient, _ *mock.MockS3Client) {
				c.EXPECT().NewS3Client().Return(nil) // factory は先に呼ばれる (constructor で取得)
			},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "error: rejects body exceeding 5MB",
			input: UploadImageUsecaseInput{
				ContentType: "image/png",
				Size:        uploadImageMaxSize + 1,
				Body:        strings.NewReader("dummy"),
			},
			mock: func(c *mock.MockClient, _ *mock.MockS3Client) {
				c.EXPECT().NewS3Client().Return(nil)
			},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "error: propagates upload error from S3Client",
			input: UploadImageUsecaseInput{
				ContentType: "image/png",
				Size:        1024,
				Body:        strings.NewReader("dummy"),
			},
			mock: func(c *mock.MockClient, s *mock.MockS3Client) {
				c.EXPECT().NewS3Client().Return(s)
				s.EXPECT().Upload(ctx, "image/png", gomock.Any()).Return("", wantUploadErr)
			},
			wantErr: wantUploadErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			c := mock.NewMockClient(ctrl)
			s := mock.NewMockS3Client(ctrl)
			tt.mock(c, s)

			uc := NewUseCase(nil, nil, c, nil).NewUploadImageUsecase()
			out, err := uc.Exec(ctx, tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, UploadImageUsecaseOutput{}, out)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantURL, out.URL)
		})
	}
}
