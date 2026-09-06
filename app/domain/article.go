package domain

import (
	"fmt"
	"time"
	"unicode/utf8"
)

const (
	ArticleTitleMaxLength = 100
	ArticleBodyMaxLength  = 50000
	ArticleMaxTags        = 5
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
	Status      ArticleStatus
	PublishedAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	TagIDs      []TagID
}

func NewArticle(slug Slug, title string, body string, status ArticleStatus, publishedAt time.Time, tagIDs []TagID) (*Article, error) {
	if title == "" {
		return nil, ErrInvalidArgument.With("タイトル は 必須 です")
	}
	if utf8.RuneCountInString(title) > ArticleTitleMaxLength {
		return nil, ErrInvalidArgument.With(fmt.Sprintf("タイトル は %d 文字以内 です", ArticleTitleMaxLength))
	}
	// draft は edit page で後から本文を書けるので空を許容する。published は必須。
	if status == ArticleStatusPublished && body == "" {
		return nil, ErrInvalidArgument.With("本文 は 必須 です")
	}
	if utf8.RuneCountInString(body) > ArticleBodyMaxLength {
		return nil, ErrInvalidArgument.With(fmt.Sprintf("本文 は %d 文字以内 です", ArticleBodyMaxLength))
	}
	if status != ArticleStatusDraft && status != ArticleStatusPublished {
		return nil, ErrInvalidArgument.With("ステータス は draft または published です")
	}
	if len(tagIDs) > ArticleMaxTags {
		return nil, ErrInvalidArgument.With(fmt.Sprintf("タグ は 最大 %d 個 です", ArticleMaxTags))
	}
	return &Article{
		Slug:        slug,
		Title:       title,
		Body:        body,
		Status:      status,
		PublishedAt: publishedAt,
		TagIDs:      tagIDs,
	}, nil
}

func (a *Article) Update(title string, body string, status ArticleStatus, tagIDs []TagID) error {
	if title == "" {
		return ErrInvalidArgument.With("タイトル は 必須 です")
	}
	if utf8.RuneCountInString(title) > ArticleTitleMaxLength {
		return ErrInvalidArgument.With(fmt.Sprintf("タイトル は %d 文字以内 です", ArticleTitleMaxLength))
	}
	if status == ArticleStatusPublished && body == "" {
		return ErrInvalidArgument.With("本文 は 必須 です")
	}
	if utf8.RuneCountInString(body) > ArticleBodyMaxLength {
		return ErrInvalidArgument.With(fmt.Sprintf("本文 は %d 文字以内 です", ArticleBodyMaxLength))
	}
	if status != ArticleStatusDraft && status != ArticleStatusPublished {
		return ErrInvalidArgument.With("ステータス は draft または published です")
	}
	if len(tagIDs) > ArticleMaxTags {
		return ErrInvalidArgument.With(fmt.Sprintf("タグ は 最大 %d 個 です", ArticleMaxTags))
	}

	if status == ArticleStatusPublished && a.PublishedAt.IsZero() {
		a.PublishedAt = time.Now().UTC()
	}

	a.Title = title
	a.Body = body
	a.Status = status
	a.TagIDs = tagIDs
	return nil
}
