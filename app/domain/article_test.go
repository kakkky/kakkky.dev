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
		summary string
		status  domain.ArticleStatus
		wantErr error
	}{
		{
			name:    "success: draft",
			title:   "タイトル",
			body:    "本文",
			summary: "要約",
			status:  domain.ArticleStatusDraft,
			wantErr: nil,
		},
		{
			name:    "success: published",
			title:   "タイトル",
			body:    "本文",
			summary: "要約",
			status:  domain.ArticleStatusPublished,
			wantErr: nil,
		},
		{
			name:    "success: empty summary",
			title:   "タイトル",
			body:    "本文",
			summary: "",
			status:  domain.ArticleStatusDraft,
			wantErr: nil,
		},
		{
			name:    "success: title at max length",
			title:   strings.Repeat("あ", domain.ArticleTitleMaxLength),
			body:    "本文",
			summary: "要約",
			status:  domain.ArticleStatusDraft,
			wantErr: nil,
		},
		{
			name:    "success: body at max length",
			title:   "タイトル",
			body:    strings.Repeat("あ", domain.ArticleBodyMaxLength),
			summary: "要約",
			status:  domain.ArticleStatusDraft,
			wantErr: nil,
		},
		{
			name:    "success: summary at max length",
			title:   "タイトル",
			body:    "本文",
			summary: strings.Repeat("あ", domain.ArticleSummaryMaxLength),
			status:  domain.ArticleStatusDraft,
			wantErr: nil,
		},
		{
			name:    "error: empty title",
			title:   "",
			body:    "本文",
			summary: "要約",
			status:  domain.ArticleStatusDraft,
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "error: title exceeds max length",
			title:   strings.Repeat("あ", domain.ArticleTitleMaxLength+1),
			body:    "本文",
			summary: "要約",
			status:  domain.ArticleStatusDraft,
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "error: empty body",
			title:   "タイトル",
			body:    "",
			summary: "要約",
			status:  domain.ArticleStatusDraft,
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "error: body exceeds max length",
			title:   "タイトル",
			body:    strings.Repeat("あ", domain.ArticleBodyMaxLength+1),
			summary: "要約",
			status:  domain.ArticleStatusDraft,
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "error: summary exceeds max length",
			title:   "タイトル",
			body:    "本文",
			summary: strings.Repeat("あ", domain.ArticleSummaryMaxLength+1),
			status:  domain.ArticleStatusDraft,
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "error: invalid status",
			title:   "タイトル",
			body:    "本文",
			summary: "要約",
			status:  domain.ArticleStatus("invalid"),
			wantErr: domain.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewArticle(
				domain.Slug("valid-slug"),
				tt.title,
				tt.body,
				tt.summary,
				tt.status,
				time.Time{},
			)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				return
			}
			assert.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantErr))
			assert.Nil(t, got)
		})
	}
}
