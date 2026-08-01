package domain_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewSeries(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
		status      domain.SeriesStatus
		wantErr     error
	}{
		{
			name:        "success: draft",
			title:       "タイトル",
			description: "説明",
			status:      domain.SeriesStatusDraft,
			wantErr:     nil,
		},
		{
			name:        "success: published_ongoing",
			title:       "タイトル",
			description: "説明",
			status:      domain.SeriesStatusPublishedOngoing,
			wantErr:     nil,
		},
		{
			name:        "success: published_completed",
			title:       "タイトル",
			description: "説明",
			status:      domain.SeriesStatusPublishedCompleted,
			wantErr:     nil,
		},
		{
			name:        "success: empty description",
			title:       "タイトル",
			description: "",
			status:      domain.SeriesStatusDraft,
			wantErr:     nil,
		},
		{
			name:        "success: title at max length",
			title:       strings.Repeat("あ", domain.SeriesTitleMaxLength),
			description: "説明",
			status:      domain.SeriesStatusDraft,
			wantErr:     nil,
		},
		{
			name:        "success: description at max length",
			title:       "タイトル",
			description: strings.Repeat("あ", domain.SeriesDescriptionMaxLength),
			status:      domain.SeriesStatusDraft,
			wantErr:     nil,
		},
		{
			name:        "error: empty title",
			title:       "",
			description: "説明",
			status:      domain.SeriesStatusDraft,
			wantErr:     domain.ErrInvalidArgument,
		},
		{
			name:        "error: title exceeds max length",
			title:       strings.Repeat("あ", domain.SeriesTitleMaxLength+1),
			description: "説明",
			status:      domain.SeriesStatusDraft,
			wantErr:     domain.ErrInvalidArgument,
		},
		{
			name:        "error: description exceeds max length",
			title:       "タイトル",
			description: strings.Repeat("あ", domain.SeriesDescriptionMaxLength+1),
			status:      domain.SeriesStatusDraft,
			wantErr:     domain.ErrInvalidArgument,
		},
		{
			name:        "error: invalid status",
			title:       "タイトル",
			description: "説明",
			status:      domain.SeriesStatus("invalid"),
			wantErr:     domain.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewSeries(
				domain.Slug("valid-slug"),
				tt.title,
				tt.description,
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

func TestSeriesAddArticle(t *testing.T) {
	tests := []struct {
		name         string
		original     []domain.SeriesArticle
		addArticleID domain.ArticleID
		addPosition  int
		wantArticles []domain.SeriesArticle
		wantErr      error
	}{
		{
			name:         "success: add first article",
			original:     nil,
			addArticleID: "a1",
			addPosition:  1,
			wantArticles: []domain.SeriesArticle{{ArticleID: "a1", Position: 1}},
			wantErr:      nil,
		},
		{
			name: "success: add second article with different position",
			original: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
			},
			addArticleID: "a2",
			addPosition:  2,
			wantArticles: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
				{ArticleID: "a2", Position: 2},
			},
			wantErr: nil,
		},
		{
			name: "success: allow position gap",
			original: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
			},
			addArticleID: "a2",
			addPosition:  5,
			wantArticles: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
				{ArticleID: "a2", Position: 5},
			},
			wantErr: nil,
		},
		{
			name: "error: duplicate article id",
			original: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
			},
			addArticleID: "a1",
			addPosition:  2,
			wantArticles: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
			},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "error: duplicate position",
			original: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
			},
			addArticleID: "a2",
			addPosition:  1,
			wantArticles: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
			},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:         "error: position zero",
			original:     nil,
			addArticleID: "a1",
			addPosition:  0,
			wantArticles: nil,
			wantErr:      domain.ErrInvalidArgument,
		},
		{
			name:         "error: position negative",
			original:     nil,
			addArticleID: "a1",
			addPosition:  -1,
			wantArticles: nil,
			wantErr:      domain.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := domain.NewSeries(
				domain.Slug("valid-slug"),
				"タイトル",
				"説明",
				domain.SeriesStatusDraft,
				time.Time{},
			)
			assert.NoError(t, err)
			s.Articles = tt.original

			err = s.AddArticle(tt.addArticleID, tt.addPosition)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			}
			assert.Equal(t, tt.wantArticles, s.Articles)
		})
	}
}

func TestSeriesAddTags(t *testing.T) {
	tests := []struct {
		name       string
		original   []domain.TagID
		addTagIDs  []domain.TagID
		wantTagIDs []domain.TagID
		wantErr    error
	}{
		{
			name:       "success: add one tag to empty",
			original:   nil,
			addTagIDs:  []domain.TagID{"t1"},
			wantTagIDs: []domain.TagID{"t1"},
			wantErr:    nil,
		},
		{
			name:       "success: add up to max from empty",
			original:   nil,
			addTagIDs:  []domain.TagID{"t1", "t2", "t3", "t4", "t5"},
			wantTagIDs: []domain.TagID{"t1", "t2", "t3", "t4", "t5"},
			wantErr:    nil,
		},
		{
			name:       "success: skip tag already in existing",
			original:   []domain.TagID{"t1", "t2"},
			addTagIDs:  []domain.TagID{"t1", "t3"},
			wantTagIDs: []domain.TagID{"t1", "t2", "t3"},
			wantErr:    nil,
		},
		{
			name:       "success: skip duplicates within input",
			original:   nil,
			addTagIDs:  []domain.TagID{"t1", "t1", "t2"},
			wantTagIDs: []domain.TagID{"t1", "t2"},
			wantErr:    nil,
		},
		{
			name:       "success: empty input keeps state unchanged",
			original:   []domain.TagID{"t1"},
			addTagIDs:  []domain.TagID{},
			wantTagIDs: []domain.TagID{"t1"},
			wantErr:    nil,
		},
		{
			name:       "success: re-adding existing tag when at max is no-op",
			original:   []domain.TagID{"t1", "t2", "t3", "t4", "t5"},
			addTagIDs:  []domain.TagID{"t1"},
			wantTagIDs: []domain.TagID{"t1", "t2", "t3", "t4", "t5"},
			wantErr:    nil,
		},
		{
			name:       "error: adding one to max exceeds limit",
			original:   []domain.TagID{"t1", "t2", "t3", "t4", "t5"},
			addTagIDs:  []domain.TagID{"t6"},
			wantTagIDs: []domain.TagID{"t1", "t2", "t3", "t4", "t5"},
			wantErr:    domain.ErrInvalidArgument,
		},
		{
			name:       "error: partial append then limit exceeded",
			original:   []domain.TagID{"t1", "t2", "t3", "t4"},
			addTagIDs:  []domain.TagID{"t5", "t6"},
			wantTagIDs: []domain.TagID{"t1", "t2", "t3", "t4", "t5"},
			wantErr:    domain.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := domain.NewSeries(
				domain.Slug("valid-slug"),
				"タイトル",
				"説明",
				domain.SeriesStatusDraft,
				time.Time{},
			)
			assert.NoError(t, err)
			s.TagIDs = tt.original

			err = s.AddTags(tt.addTagIDs)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
			}
			assert.Equal(t, tt.wantTagIDs, s.TagIDs)
		})
	}
}
