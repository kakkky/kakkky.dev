package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/kakkky/kakkky.dev/testhelper/mock"
)

func TestCreateArticleUsecase_Exec(t *testing.T) {
	ctx := context.Background()

	var (
		existingTagID domain.TagID = "11111111-1111-1111-1111-111111111111"
		newTagID      domain.TagID = "22222222-2222-2222-2222-222222222222"
	)
	existingTag := &domain.Tag{ID: existingTagID, Slug: "go", Name: "Go"}

	wantErr := errors.New("boom")

	tests := []struct {
		name     string
		input    CreateArticleUsecaseInput
		mock     func(repo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository)
		wantSlug domain.Slug
		wantErr  error
	}{
		{
			name: "creates article with existing and new tags, slug derived from title",
			input: CreateArticleUsecaseInput{
				Title:    "Hello World",
				Body:     "body",
				Status:   string(domain.ArticleStatusDraft),
				TagNames: []string{"Go", "設計"},
			},
			mock: func(repo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(withTxRunner(repo))
				repo.EXPECT().NewArticleRepository().Return(ar).Times(2)
				repo.EXPECT().NewTagRepository().Return(tr)

				// unique 判定: slug 空きあり
				ar.EXPECT().FindBySlug(ctx, domain.Slug("hello-world")).Return(nil, domain.ErrNotFound)

				tr.EXPECT().FindByNames(ctx, "Go", "設計").Return([]*domain.Tag{existingTag}, nil)
				tr.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, tag *domain.Tag) error {
					assert.Equal(t, "設計", tag.Name)
					assert.Equal(t, domain.Slug("設計"), tag.Slug)
					tag.ID = newTagID
					return nil
				})
				ar.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, a *domain.Article) error {
					assert.Equal(t, domain.Slug("hello-world"), a.Slug)
					assert.Equal(t, "Hello World", a.Title)
					assert.Equal(t, domain.ArticleStatusDraft, a.Status)
					assert.True(t, a.PublishedAt.IsZero())
					assert.Equal(t, []domain.TagID{existingTagID, newTagID}, a.TagIDs)
					return nil
				})
			},
			wantSlug: "hello-world",
		},
		{
			name: "appends suffix on slug collision",
			input: CreateArticleUsecaseInput{
				Title:  "Hello",
				Body:   "body",
				Status: string(domain.ArticleStatusDraft),
			},
			mock: func(repo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(withTxRunner(repo))
				repo.EXPECT().NewArticleRepository().Return(ar).Times(2)
				repo.EXPECT().NewTagRepository().Return(tr)

				// 1 回目衝突、2 回目空き
				ar.EXPECT().FindBySlug(ctx, domain.Slug("hello")).Return(&domain.Article{}, nil)
				ar.EXPECT().FindBySlug(ctx, domain.Slug("hello-2")).Return(nil, domain.ErrNotFound)

				ar.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, a *domain.Article) error {
					assert.Equal(t, domain.Slug("hello-2"), a.Slug)
					return nil
				})
			},
			wantSlug: "hello-2",
		},
		{
			name: "sets published_at when status is published",
			input: CreateArticleUsecaseInput{
				Title:  "Hello",
				Body:   "body",
				Status: string(domain.ArticleStatusPublished),
			},
			mock: func(repo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(withTxRunner(repo))
				repo.EXPECT().NewArticleRepository().Return(ar).Times(2)
				repo.EXPECT().NewTagRepository().Return(tr)
				ar.EXPECT().FindBySlug(ctx, domain.Slug("hello")).Return(nil, domain.ErrNotFound)
				ar.EXPECT().Create(ctx, gomock.Any()).DoAndReturn(func(_ context.Context, a *domain.Article) error {
					assert.False(t, a.PublishedAt.IsZero())
					return nil
				})
			},
			wantSlug: "hello",
		},
		{
			name: "returns error on invalid article (empty title)",
			input: CreateArticleUsecaseInput{
				Title:  "",
				Body:   "body",
				Status: string(domain.ArticleStatusDraft),
			},
			mock:    func(repo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "propagates error from Create",
			input: CreateArticleUsecaseInput{
				Title:  "Hello",
				Body:   "body",
				Status: string(domain.ArticleStatusDraft),
			},
			mock: func(repo *mock.MockRepository, ar *mock.MockArticleRepository, tr *mock.MockTagRepository) {
				repo.EXPECT().WithTx(ctx, gomock.Any()).DoAndReturn(withTxRunner(repo))
				repo.EXPECT().NewArticleRepository().Return(ar).Times(2)
				repo.EXPECT().NewTagRepository().Return(tr)
				ar.EXPECT().FindBySlug(ctx, domain.Slug("hello")).Return(nil, domain.ErrNotFound)
				ar.EXPECT().Create(ctx, gomock.Any()).Return(wantErr)
			},
			wantErr: wantErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			repo := mock.NewMockRepository(ctrl)
			ar := mock.NewMockArticleRepository(ctrl)
			tr := mock.NewMockTagRepository(ctrl)
			qs := mock.NewMockQueryService(ctrl)

			tt.mock(repo, ar, tr)

			us := NewUseCase(repo, qs).NewCreateArticleUsecase()
			got, err := us.Exec(ctx, tt.input)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Equal(t, CreateArticleUsecaseOutput{}, got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantSlug, got.Slug)
		})
	}
}

// withTxRunner は MockRepository.WithTx を「fn を実行する」動作にする helper。
// 本来 mock は fn を呼ばないため。
func withTxRunner(repo domain.Repository) func(context.Context, func(domain.Repository) error) error {
	return func(_ context.Context, fn func(domain.Repository) error) error {
		return fn(repo)
	}
}
