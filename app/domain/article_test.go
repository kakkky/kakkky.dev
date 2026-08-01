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

func TestArticleAddTags(t *testing.T) {
	tests := []struct {
		name       string
		original    []domain.TagID
		addTagIDs  []domain.TagID
		wantTagIDs []domain.TagID
		wantErr    error
	}{
		{
			name:       "success: add one tag to empty",
			original:    nil,
			addTagIDs:  []domain.TagID{"t1"},
			wantTagIDs: []domain.TagID{"t1"},
			wantErr:    nil,
		},
		{
			name:       "success: add up to max from empty",
			original:    nil,
			addTagIDs:  []domain.TagID{"t1", "t2", "t3", "t4", "t5"},
			wantTagIDs: []domain.TagID{"t1", "t2", "t3", "t4", "t5"},
			wantErr:    nil,
		},
		{
			name:       "success: skip tag already in existing",
			original:    []domain.TagID{"t1", "t2"},
			addTagIDs:  []domain.TagID{"t1", "t3"},
			wantTagIDs: []domain.TagID{"t1", "t2", "t3"},
			wantErr:    nil,
		},
		{
			name:       "success: skip duplicates within input",
			original:    nil,
			addTagIDs:  []domain.TagID{"t1", "t1", "t2"},
			wantTagIDs: []domain.TagID{"t1", "t2"},
			wantErr:    nil,
		},
		{
			name:       "success: empty input keeps state unchanged",
			original:    []domain.TagID{"t1"},
			addTagIDs:  []domain.TagID{},
			wantTagIDs: []domain.TagID{"t1"},
			wantErr:    nil,
		},
		{
			name:       "success: re-adding existing tag when at max is no-op",
			original:    []domain.TagID{"t1", "t2", "t3", "t4", "t5"},
			addTagIDs:  []domain.TagID{"t1"},
			wantTagIDs: []domain.TagID{"t1", "t2", "t3", "t4", "t5"},
			wantErr:    nil,
		},
		{
			name:       "error: adding one to max exceeds limit",
			original:    []domain.TagID{"t1", "t2", "t3", "t4", "t5"},
			addTagIDs:  []domain.TagID{"t6"},
			wantTagIDs: []domain.TagID{"t1", "t2", "t3", "t4", "t5"},
			wantErr:    domain.ErrInvalidArgument,
		},
		{
			name:       "error: partial append then limit exceeded",
			original:    []domain.TagID{"t1", "t2", "t3", "t4"},
			addTagIDs:  []domain.TagID{"t5", "t6"},
			wantTagIDs: []domain.TagID{"t1", "t2", "t3", "t4", "t5"},
			wantErr:    domain.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := domain.NewArticle(
				domain.Slug("valid-slug"),
				"タイトル",
				"本文",
				"要約",
				domain.ArticleStatusDraft,
				time.Time{},
			)
			assert.NoError(t, err)
			a.TagIDs = tt.original

			err = a.AddTags(tt.addTagIDs)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			}
			assert.Equal(t, tt.wantTagIDs, a.TagIDs)
		})
	}
}
