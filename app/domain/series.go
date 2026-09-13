package domain

import (
	"fmt"
	"slices"
	"time"
	"unicode/utf8"
)

const (
	SeriesTitleMaxLength       = 100
	SeriesDescriptionMaxLength = 500
	SeriesMaxTags              = 5
)

type SeriesID string

type SeriesStatus string

const (
	SeriesStatusDraft              SeriesStatus = "draft"
	SeriesStatusPublishedOngoing   SeriesStatus = "published_ongoing"
	SeriesStatusPublishedCompleted SeriesStatus = "published_completed"
)

type SeriesArticle struct {
	ArticleID ArticleID
	Position  int
}

type Series struct {
	ID          SeriesID
	Slug        Slug
	Title       string
	Description string
	Status      SeriesStatus
	PublishedAt time.Time
	CreatedAt   time.Time
	TagIDs      []TagID
	Articles    []SeriesArticle
}

func NewSeries(slug Slug, title string, description string, status SeriesStatus, publishedAt time.Time) (*Series, error) {
	if title == "" {
		return nil, ErrInvalidArgument.With("タイトル は 必須 です")
	}
	if utf8.RuneCountInString(title) > SeriesTitleMaxLength {
		return nil, ErrInvalidArgument.With(fmt.Sprintf("タイトル は %d 文字以内 です", SeriesTitleMaxLength))
	}
	if utf8.RuneCountInString(description) > SeriesDescriptionMaxLength {
		return nil, ErrInvalidArgument.With(fmt.Sprintf("説明 は %d 文字以内 です", SeriesDescriptionMaxLength))
	}
	if status != SeriesStatusDraft &&
		status != SeriesStatusPublishedOngoing &&
		status != SeriesStatusPublishedCompleted {
		return nil, ErrInvalidArgument.With("ステータス は draft, published_ongoing, published_completed のいずれか です")
	}
	return &Series{
		Slug:        slug,
		Title:       title,
		Description: description,
		Status:      status,
		PublishedAt: publishedAt,
	}, nil
}

func (s *Series) Update(title string, description string, status SeriesStatus, tagIDs []TagID) error {
	if title == "" {
		return ErrInvalidArgument.With("タイトル は 必須 です")
	}
	if utf8.RuneCountInString(title) > SeriesTitleMaxLength {
		return ErrInvalidArgument.With(fmt.Sprintf("タイトル は %d 文字以内 です", SeriesTitleMaxLength))
	}
	if utf8.RuneCountInString(description) > SeriesDescriptionMaxLength {
		return ErrInvalidArgument.With(fmt.Sprintf("説明 は %d 文字以内 です", SeriesDescriptionMaxLength))
	}
	if status != SeriesStatusDraft &&
		status != SeriesStatusPublishedOngoing &&
		status != SeriesStatusPublishedCompleted {
		return ErrInvalidArgument.With("ステータス は draft, published_ongoing, published_completed のいずれか です")
	}
	if len(tagIDs) > SeriesMaxTags {
		return ErrInvalidArgument.With(fmt.Sprintf("タグ は 最大 %d 個 です", SeriesMaxTags))
	}

	if (status == SeriesStatusPublishedOngoing || status == SeriesStatusPublishedCompleted) && s.PublishedAt.IsZero() {
		s.PublishedAt = time.Now().UTC()
	}

	s.Title = title
	s.Description = description
	s.Status = status
	s.TagIDs = tagIDs
	return nil
}

func (s *Series) AddArticle(articleID ArticleID) error {
	for _, a := range s.Articles {
		if a.ArticleID == articleID {
			return ErrInvalidArgument.With("この article は既にこの series に含まれています")
		}
	}
	// position は 1 始まり
	nextPosition := 1
	if n := len(s.Articles); n > 0 {
		nextPosition = s.Articles[n-1].Position + 1
	}
	s.Articles = append(s.Articles, SeriesArticle{ArticleID: articleID, Position: nextPosition})
	return nil
}

func (s *Series) RemoveArticle(articleID ArticleID) error {
	for i, a := range s.Articles {
		if a.ArticleID == articleID {
			s.Articles = append(s.Articles[:i], s.Articles[i+1:]...)
			return nil
		}
	}
	return ErrInvalidArgument.With(fmt.Sprintf("article %s は この series に 含まれていません", articleID))
}

func (s *Series) ReorderArticles(articleIDs []ArticleID) error {
	if len(articleIDs) != len(s.Articles) {
		return ErrInvalidArgument.With("articles の 個数 が 一致 しません")
	}
	seen := make(map[ArticleID]struct{}, len(articleIDs))
	for _, id := range articleIDs {
		if _, ok := seen[id]; ok {
			return ErrInvalidArgument.With(fmt.Sprintf("article %s が 重複 しています", id))
		}
		seen[id] = struct{}{}
	}
	current := make(map[ArticleID]struct{}, len(s.Articles))
	for _, a := range s.Articles {
		current[a.ArticleID] = struct{}{}
	}
	for _, id := range articleIDs {
		if _, ok := current[id]; !ok {
			return ErrInvalidArgument.With(fmt.Sprintf("article %s は この series に 含まれていません", id))
		}
	}
	next := make([]SeriesArticle, len(articleIDs))
	for i, id := range articleIDs {
		next[i] = SeriesArticle{ArticleID: id, Position: i + 1}
	}
	s.Articles = next
	return nil
}

func (s *Series) AddTags(tagIDs []TagID) error {
	for _, id := range tagIDs {
		if slices.Contains(s.TagIDs, id) {
			continue
		}
		if len(s.TagIDs) >= SeriesMaxTags {
			return ErrInvalidArgument.With(fmt.Sprintf("タグ は 最大 %d 個 です", SeriesMaxTags))
		}
		s.TagIDs = append(s.TagIDs, id)
	}
	return nil
}
