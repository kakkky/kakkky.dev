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

func (s *Series) AddArticle(articleID ArticleID, position int) error {
	if position <= 0 {
		return ErrInvalidArgument.With("position は 1 以上 です")
	}
	for _, a := range s.Articles {
		if a.ArticleID == articleID {
			return ErrInvalidArgument.With("この記事は既にこの連載に含まれています")
		}
		if a.Position == position {
			return ErrInvalidArgument.With(fmt.Sprintf("position %d は既に使用されています", position))
		}
	}
	s.Articles = append(s.Articles, SeriesArticle{ArticleID: articleID, Position: position})
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
