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

func TestSeriesUpdate(t *testing.T) {
	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		originalStatus domain.SeriesStatus
		originalPubAt  time.Time
		originalTagIDs []domain.TagID
		title          string
		description    string
		status         domain.SeriesStatus
		tagIDs         []domain.TagID
		wantPubAtZero  bool
		wantPubAtNow   bool // now と 一致 する か (draft → published の 初回 セット)
		wantPubAt      time.Time
		wantErr        error
	}{
		{
			name:           "success: draft to published_ongoing sets PublishedAt to now",
			originalStatus: domain.SeriesStatusDraft,
			originalPubAt:  time.Time{},
			title:          "new title",
			description:    "new desc",
			status:         domain.SeriesStatusPublishedOngoing,
			tagIDs:         []domain.TagID{"t1"},
			wantPubAtNow:   true,
		},
		{
			name:           "success: draft to published_completed sets PublishedAt to now",
			originalStatus: domain.SeriesStatusDraft,
			originalPubAt:  time.Time{},
			title:          "t",
			status:         domain.SeriesStatusPublishedCompleted,
			wantPubAtNow:   true,
		},
		{
			name:           "success: published_ongoing to published_completed preserves PublishedAt",
			originalStatus: domain.SeriesStatusPublishedOngoing,
			originalPubAt:  baseTime,
			title:          "t",
			status:         domain.SeriesStatusPublishedCompleted,
			wantPubAt:      baseTime,
		},
		{
			name:           "success: published to draft preserves PublishedAt",
			originalStatus: domain.SeriesStatusPublishedOngoing,
			originalPubAt:  baseTime,
			title:          "t",
			status:         domain.SeriesStatusDraft,
			wantPubAt:      baseTime,
		},
		{
			name:           "success: replaces tag ids",
			originalStatus: domain.SeriesStatusDraft,
			originalTagIDs: []domain.TagID{"t1", "t2"},
			title:          "t",
			status:         domain.SeriesStatusDraft,
			tagIDs:         []domain.TagID{"t3"},
			wantPubAtZero:  true,
		},
		{
			name:           "success: empty tag ids clears tags",
			originalStatus: domain.SeriesStatusDraft,
			originalTagIDs: []domain.TagID{"t1"},
			title:          "t",
			status:         domain.SeriesStatusDraft,
			tagIDs:         nil,
			wantPubAtZero:  true,
		},
		{
			name:           "error: empty title",
			originalStatus: domain.SeriesStatusDraft,
			title:          "",
			status:         domain.SeriesStatusDraft,
			wantErr:        domain.ErrInvalidArgument,
		},
		{
			name:           "error: title exceeds max length",
			originalStatus: domain.SeriesStatusDraft,
			title:          strings.Repeat("あ", domain.SeriesTitleMaxLength+1),
			status:         domain.SeriesStatusDraft,
			wantErr:        domain.ErrInvalidArgument,
		},
		{
			name:           "error: description exceeds max length",
			originalStatus: domain.SeriesStatusDraft,
			title:          "t",
			description:    strings.Repeat("あ", domain.SeriesDescriptionMaxLength+1),
			status:         domain.SeriesStatusDraft,
			wantErr:        domain.ErrInvalidArgument,
		},
		{
			name:           "error: invalid status",
			originalStatus: domain.SeriesStatusDraft,
			title:          "t",
			status:         domain.SeriesStatus("invalid"),
			wantErr:        domain.ErrInvalidArgument,
		},
		{
			name:           "error: tag ids exceed max",
			originalStatus: domain.SeriesStatusDraft,
			title:          "t",
			status:         domain.SeriesStatusDraft,
			tagIDs:         []domain.TagID{"t1", "t2", "t3", "t4", "t5", "t6"},
			wantErr:        domain.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := domain.NewSeries(
				domain.Slug("valid-slug"),
				"original",
				"original desc",
				tt.originalStatus,
				tt.originalPubAt,
			)
			assert.NoError(t, err)
			s.TagIDs = tt.originalTagIDs

			before := time.Now().UTC()
			err = s.Update(tt.title, tt.description, tt.status, tt.tagIDs)
			after := time.Now().UTC()

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.title, s.Title)
			assert.Equal(t, tt.description, s.Description)
			assert.Equal(t, tt.status, s.Status)
			assert.Equal(t, tt.tagIDs, s.TagIDs)

			switch {
			case tt.wantPubAtNow:
				assert.False(t, s.PublishedAt.Before(before), "PublishedAt should be >= before")
				assert.False(t, s.PublishedAt.After(after), "PublishedAt should be <= after")
			case tt.wantPubAtZero:
				assert.True(t, s.PublishedAt.IsZero())
			default:
				assert.Equal(t, tt.wantPubAt, s.PublishedAt)
			}
		})
	}
}

func TestSeriesAddArticle(t *testing.T) {
	tests := []struct {
		name         string
		original     []domain.SeriesArticle
		addArticleID domain.ArticleID
		wantArticles []domain.SeriesArticle
		wantErr      error
	}{
		{
			name:         "success: first article gets position 1",
			original:     nil,
			addArticleID: "a1",
			wantArticles: []domain.SeriesArticle{{ArticleID: "a1", Position: 1}},
		},
		{
			name: "success: appends after existing max position",
			original: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
				{ArticleID: "a2", Position: 2},
			},
			addArticleID: "a3",
			wantArticles: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
				{ArticleID: "a2", Position: 2},
				{ArticleID: "a3", Position: 3},
			},
		},
		{
			name: "success: uses last position + 1 (gaps preserved)",
			original: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
				{ArticleID: "a2", Position: 5},
			},
			addArticleID: "a3",
			wantArticles: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
				{ArticleID: "a2", Position: 5},
				{ArticleID: "a3", Position: 6},
			},
		},
		{
			name: "error: duplicate article id",
			original: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
			},
			addArticleID: "a1",
			wantArticles: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
			},
			wantErr: domain.ErrInvalidArgument,
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

			err = s.AddArticle(tt.addArticleID)
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

func TestSeriesRemoveArticle(t *testing.T) {
	tests := []struct {
		name         string
		original     []domain.SeriesArticle
		removeID     domain.ArticleID
		wantArticles []domain.SeriesArticle
		wantErr      error
	}{
		{
			name: "success: removes middle article and preserves positions",
			original: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
				{ArticleID: "a2", Position: 2},
				{ArticleID: "a3", Position: 3},
			},
			removeID: "a2",
			wantArticles: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
				{ArticleID: "a3", Position: 3},
			},
		},
		{
			name: "success: removes single article",
			original: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
			},
			removeID:     "a1",
			wantArticles: []domain.SeriesArticle{},
		},
		{
			name: "error: article not in series",
			original: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
			},
			removeID: "a2",
			wantArticles: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
			},
			wantErr: domain.ErrInvalidArgument,
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

			err = s.RemoveArticle(tt.removeID)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantArticles, s.Articles)
			} else {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				assert.Equal(t, tt.wantArticles, s.Articles)
			}
		})
	}
}

func TestSeriesReorderArticles(t *testing.T) {
	tests := []struct {
		name         string
		original     []domain.SeriesArticle
		input        []domain.ArticleID
		wantArticles []domain.SeriesArticle
		wantErr      error
	}{
		{
			name: "success: reorder assigns 1..n in input order",
			original: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
				{ArticleID: "a2", Position: 2},
				{ArticleID: "a3", Position: 3},
			},
			input: []domain.ArticleID{"a3", "a1", "a2"},
			wantArticles: []domain.SeriesArticle{
				{ArticleID: "a3", Position: 1},
				{ArticleID: "a1", Position: 2},
				{ArticleID: "a2", Position: 3},
			},
		},
		{
			name: "success: same order compacts gaps",
			original: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
				{ArticleID: "a2", Position: 5},
			},
			input: []domain.ArticleID{"a1", "a2"},
			wantArticles: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
				{ArticleID: "a2", Position: 2},
			},
		},
		{
			name:         "success: empty when originally empty",
			original:     nil,
			input:        nil,
			wantArticles: []domain.SeriesArticle{},
		},
		{
			name: "error: count mismatch (missing)",
			original: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
				{ArticleID: "a2", Position: 2},
			},
			input: []domain.ArticleID{"a1"},
			wantArticles: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
				{ArticleID: "a2", Position: 2},
			},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "error: count mismatch (extra)",
			original: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
			},
			input: []domain.ArticleID{"a1", "a2"},
			wantArticles: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
			},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "error: duplicate article id in input",
			original: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
				{ArticleID: "a2", Position: 2},
			},
			input: []domain.ArticleID{"a1", "a1"},
			wantArticles: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
				{ArticleID: "a2", Position: 2},
			},
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name: "error: unknown article id",
			original: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
				{ArticleID: "a2", Position: 2},
			},
			input: []domain.ArticleID{"a1", "a3"},
			wantArticles: []domain.SeriesArticle{
				{ArticleID: "a1", Position: 1},
				{ArticleID: "a2", Position: 2},
			},
			wantErr: domain.ErrInvalidArgument,
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

			err = s.ReorderArticles(tt.input)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantArticles, s.Articles)
			} else {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				assert.Equal(t, tt.wantArticles, s.Articles)
			}
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
