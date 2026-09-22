package usecase

import (
	"context"
	"github.com/kakkky/kakkky.dev/errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kakkky/kakkky.dev/adapter/cache"
	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/testhelper/mock"
)

func TestGetLinkPreviewUsecase_Exec(t *testing.T) {
	ctx := context.Background()

	url := "https://example.com/article"
	fetched := domain.OGPData{
		Host:        "example.com",
		Title:       "Example",
		Description: "desc",
		Image:       "https://example.com/ogp.png",
	}
	wantFetchErr := errors.New("boom")

	tests := []struct {
		name      string
		mock      func(c *mock.MockClient, f *mock.MockOGPFetcher)
		execTimes int
		wantData  domain.OGPData
		wantErr   error
	}{
		{
			name: "success: cache miss triggers fetch and returns OGP data",
			mock: func(c *mock.MockClient, f *mock.MockOGPFetcher) {
				c.EXPECT().NewOGPFetcher().Return(f)
				f.EXPECT().Fetch(ctx, url).Return(fetched, nil)
			},
			wantData: fetched,
		},
		{
			name: "error: propagates fetcher error",
			mock: func(c *mock.MockClient, f *mock.MockOGPFetcher) {
				c.EXPECT().NewOGPFetcher().Return(f)
				f.EXPECT().Fetch(ctx, url).Return(domain.OGPData{}, wantFetchErr)
			},
			wantErr: wantFetchErr,
		},
		{
			name: "success: cache hit skips fetcher on repeated URL",
			mock: func(c *mock.MockClient, f *mock.MockOGPFetcher) {
				c.EXPECT().NewOGPFetcher().Return(f)
				// Times(1) で 2 回目の Exec が cache から返ることを保証
				f.EXPECT().Fetch(ctx, url).Return(fetched, nil).Times(1)
			},
			execTimes: 2,
			wantData:  fetched,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			c := mock.NewMockClient(ctrl)
			f := mock.NewMockOGPFetcher(ctrl)
			tt.mock(c, f)

			uc := NewUseCase(nil, nil, c, cache.NewCache()).NewGetLinkPreviewUsecase()

			n := tt.execTimes
			if n <= 0 {
				n = 1
			}
			var (
				out GetLinkPreviewUsecaseOutput
				err error
			)
			for range n {
				out, err = uc.Exec(ctx, GetLinkPreviewUsecaseInput{URL: url})
			}

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, GetLinkPreviewUsecaseOutput{}, out)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantData, out.Data)
		})
	}
}
