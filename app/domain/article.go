package domain

import (
	"fmt"
	"slices"
	"time"
	"unicode/utf8"
)

const (
	ArticleTitleMaxLength   = 100
	ArticleBodyMaxLength    = 50000
	ArticleSummaryMaxLength = 200
	ArticleMaxTags          = 5
)

type ArticleID string

type ArticleStatus string

const (
	ArticleStatusDraft     ArticleStatus = "draft"
	ArticleStatusPublished ArticleStatus = "published"
)

type Article struct {
	ID          ArticleID
	Slug        Slug
	Title       string
	Body        string
	Summary     string
	Status      ArticleStatus
	PublishedAt time.Time
	TagIDs      []TagID
}

func NewArticle(slug Slug, title string, body string, summary string, status ArticleStatus, publishedAt time.Time) (*Article, error) {
	if title == "" {
		return nil, ErrInvalidArgument.With("タイトル は 必須 です")
	}
	if utf8.RuneCountInString(title) > ArticleTitleMaxLength {
		return nil, ErrInvalidArgument.With(fmt.Sprintf("タイトル は %d 文字以内 です", ArticleTitleMaxLength))
	}
	if body == "" {
		return nil, ErrInvalidArgument.With("本文 は 必須 です")
	}
	if utf8.RuneCountInString(body) > ArticleBodyMaxLength {
		return nil, ErrInvalidArgument.With(fmt.Sprintf("本文 は %d 文字以内 です", ArticleBodyMaxLength))
	}
	if utf8.RuneCountInString(summary) > ArticleSummaryMaxLength {
		return nil, ErrInvalidArgument.With(fmt.Sprintf("要約 は %d 文字以内 です", ArticleSummaryMaxLength))
	}
	if status != ArticleStatusDraft && status != ArticleStatusPublished {
		return nil, ErrInvalidArgument.With("ステータス は draft または published です")
	}
	return &Article{
		Slug:        slug,
		Title:       title,
		Body:        body,
		Summary:     summary,
		Status:      status,
		PublishedAt: publishedAt,
	}, nil
}

func (a *Article) AddTags(tagIDs []TagID) error {
	for _, id := range tagIDs {
		if slices.Contains(a.TagIDs, id) {
			continue
		}
		if len(a.TagIDs) >= ArticleMaxTags {
			return ErrInvalidArgument.With(
				fmt.Sprintf("タグ は 最大 %d 個 です", ArticleMaxTags),
			)
		}
		a.TagIDs = append(a.TagIDs, id)
	}
	return nil
}
