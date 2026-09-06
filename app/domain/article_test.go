package domain_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewArticle(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		body    string
		status  domain.ArticleStatus
		tagIDs  []domain.TagID
		wantErr error
	}{
		{
			name:    "success: draft",
			title:   "タイトル",
			body:    "本文",
			status:  domain.ArticleStatusDraft,
			wantErr: nil,
		},
		{
			name:    "success: published",
			title:   "タイトル",
			body:    "本文",
			status:  domain.ArticleStatusPublished,
			wantErr: nil,
		},
		{
			name:    "success: title at max length",
			title:   strings.Repeat("あ", domain.ArticleTitleMaxLength),
			body:    "本文",
			status:  domain.ArticleStatusDraft,
			wantErr: nil,
		},
		{
			name:    "success: body at max length",
			title:   "タイトル",
			body:    strings.Repeat("あ", domain.ArticleBodyMaxLength),
			status:  domain.ArticleStatusDraft,
			wantErr: nil,
		},
		{
			name:    "success: tags at max count",
			title:   "タイトル",
			body:    "本文",
			status:  domain.ArticleStatusDraft,
			tagIDs:  []domain.TagID{"t1", "t2", "t3", "t4", "t5"},
			wantErr: nil,
		},
		{
			name:    "error: empty title",
			title:   "",
			body:    "本文",
			status:  domain.ArticleStatusDraft,
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "error: title exceeds max length",
			title:   strings.Repeat("あ", domain.ArticleTitleMaxLength+1),
			body:    "本文",
			status:  domain.ArticleStatusDraft,
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "success: empty body allowed for draft",
			title:   "タイトル",
			body:    "",
			status:  domain.ArticleStatusDraft,
			wantErr: nil,
		},
		{
			name:    "error: empty body rejected for published",
			title:   "タイトル",
			body:    "",
			status:  domain.ArticleStatusPublished,
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "error: body exceeds max length",
			title:   "タイトル",
			body:    strings.Repeat("あ", domain.ArticleBodyMaxLength+1),
			status:  domain.ArticleStatusDraft,
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "error: invalid status",
			title:   "タイトル",
			body:    "本文",
			status:  domain.ArticleStatus("invalid"),
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "error: tags exceed max count",
			title:   "タイトル",
			body:    "本文",
			status:  domain.ArticleStatusDraft,
			tagIDs:  []domain.TagID{"t1", "t2", "t3", "t4", "t5", "t6"},
			wantErr: domain.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewArticle(
				domain.Slug("valid-slug"),
				tt.title,
				tt.body,
				tt.status,
				time.Time{},
				tt.tagIDs,
			)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.tagIDs, got.TagIDs)
				return
			}
			assert.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantErr))
			assert.Nil(t, got)
		})
	}
}

func TestArticleUpdate(t *testing.T) {
	oldPublishedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	// publishedAtCheck: PublishedAt 検証をカスタマイズする。nil の場合は wantPublishedAt と Equal 比較。
	tests := []struct {
		name               string
		initial            *domain.Article
		title              string
		body               string
		status             domain.ArticleStatus
		tagIDs             []domain.TagID
		wantTitle          string
		wantBody           string
		wantStatus         domain.ArticleStatus
		wantPublishedAt    time.Time
		publishedAtCheck   func(t *testing.T, got time.Time)
		wantTagIDs         []domain.TagID
		wantErr            error
	}{
		{
			name: "success: draft→draft update fields, published_at stays zero",
			initial: &domain.Article{
				Title: "旧", Body: "旧本文", Status: domain.ArticleStatusDraft,
			},
			title:           "新",
			body:            "新本文",
			status:          domain.ArticleStatusDraft,
			wantTitle:       "新",
			wantBody:        "新本文",
			wantStatus:      domain.ArticleStatusDraft,
			wantPublishedAt: time.Time{},
		},
		{
			name: "success: draft→published sets published_at to now",
			initial: &domain.Article{
				Title: "旧", Body: "旧本文", Status: domain.ArticleStatusDraft,
			},
			title:      "新",
			body:       "新本文",
			status:     domain.ArticleStatusPublished,
			wantTitle:  "新",
			wantBody:   "新本文",
			wantStatus: domain.ArticleStatusPublished,
			publishedAtCheck: func(t *testing.T, got time.Time) {
				assert.False(t, got.IsZero(), "PublishedAt should be set to non-zero")
			},
		},
		{
			name: "success: published→published keeps original published_at",
			initial: &domain.Article{
				Title: "旧", Body: "旧本文", Status: domain.ArticleStatusPublished, PublishedAt: oldPublishedAt,
			},
			title:           "新",
			body:            "新本文",
			status:          domain.ArticleStatusPublished,
			wantTitle:       "新",
			wantBody:        "新本文",
			wantStatus:      domain.ArticleStatusPublished,
			wantPublishedAt: oldPublishedAt,
		},
		{
			name: "success: published→draft keeps published_at (再公開時に再利用)",
			initial: &domain.Article{
				Title: "旧", Body: "旧本文", Status: domain.ArticleStatusPublished, PublishedAt: oldPublishedAt,
			},
			title:           "新",
			body:            "新本文",
			status:          domain.ArticleStatusDraft,
			wantTitle:       "新",
			wantBody:        "新本文",
			wantStatus:      domain.ArticleStatusDraft,
			wantPublishedAt: oldPublishedAt,
		},
		{
			name: "success: replaces tag ids",
			initial: &domain.Article{
				Title: "旧", Body: "旧本文", Status: domain.ArticleStatusDraft,
				TagIDs: []domain.TagID{"t1", "t2"},
			},
			title:           "新",
			body:            "新本文",
			status:          domain.ArticleStatusDraft,
			tagIDs:          []domain.TagID{"t2", "t3"},
			wantTitle:       "新",
			wantBody:        "新本文",
			wantStatus:      domain.ArticleStatusDraft,
			wantPublishedAt: time.Time{},
			wantTagIDs:      []domain.TagID{"t2", "t3"},
		},
		{
			name: "success: empty tags detaches all",
			initial: &domain.Article{
				Title: "旧", Body: "旧本文", Status: domain.ArticleStatusDraft,
				TagIDs: []domain.TagID{"t1"},
			},
			title:           "新",
			body:            "新本文",
			status:          domain.ArticleStatusDraft,
			tagIDs:          nil,
			wantTitle:       "新",
			wantBody:        "新本文",
			wantStatus:      domain.ArticleStatusDraft,
			wantPublishedAt: time.Time{},
			wantTagIDs:      nil,
		},
		{
			name: "error: empty title",
			initial: &domain.Article{
				Title: "旧", Body: "旧本文", Status: domain.ArticleStatusDraft,
			},
			title:   "",
			body:    "新本文",
			status:  domain.ArticleStatusDraft,
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "error: title exceeds max length",
			initial: &domain.Article{
				Title: "旧", Body: "旧本文", Status: domain.ArticleStatusDraft,
			},
			title:   strings.Repeat("あ", domain.ArticleTitleMaxLength+1),
			body:    "新本文",
			status:  domain.ArticleStatusDraft,
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "error: published requires body",
			initial: &domain.Article{
				Title: "旧", Body: "旧本文", Status: domain.ArticleStatusDraft,
			},
			title:   "新",
			body:    "",
			status:  domain.ArticleStatusPublished,
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "error: body exceeds max length",
			initial: &domain.Article{
				Title: "旧", Body: "旧本文", Status: domain.ArticleStatusDraft,
			},
			title:   "新",
			body:    strings.Repeat("あ", domain.ArticleBodyMaxLength+1),
			status:  domain.ArticleStatusDraft,
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "error: invalid status",
			initial: &domain.Article{
				Title: "旧", Body: "旧本文", Status: domain.ArticleStatusDraft,
			},
			title:   "新",
			body:    "新本文",
			status:  domain.ArticleStatus("invalid"),
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "error: tags exceed max count",
			initial: &domain.Article{
				Title: "旧", Body: "旧本文", Status: domain.ArticleStatusDraft,
			},
			title:   "新",
			body:    "新本文",
			status:  domain.ArticleStatusDraft,
			tagIDs:  []domain.TagID{"t1", "t2", "t3", "t4", "t5", "t6"},
			wantErr: domain.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.initial.Update(tt.title, tt.body, tt.status, tt.tagIDs)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantTitle, tt.initial.Title)
			assert.Equal(t, tt.wantBody, tt.initial.Body)
			assert.Equal(t, tt.wantStatus, tt.initial.Status)
			if tt.publishedAtCheck != nil {
				tt.publishedAtCheck(t, tt.initial.PublishedAt)
			} else {
				assert.Equal(t, tt.wantPublishedAt, tt.initial.PublishedAt)
			}
			assert.Equal(t, tt.wantTagIDs, tt.initial.TagIDs)
		})
	}
}
