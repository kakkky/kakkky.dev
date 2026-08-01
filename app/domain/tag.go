package domain

import (
	"fmt"
	"unicode/utf8"
)

const (
	TagNameMaxLength = 30
)

type TagID string

type Tag struct {
	ID   TagID
	Slug Slug
	Name string
}

func NewTag(slug Slug, name string) (*Tag, error) {
	if name == "" {
		return nil, ErrInvalidArgument.With("タグ名 は 必須 です")
	}
	if utf8.RuneCountInString(name) > TagNameMaxLength {
		return nil, ErrInvalidArgument.With(fmt.Sprintf("タグ名 は %d 文字以内 です", TagNameMaxLength))
	}
	return &Tag{
		Slug: slug,
		Name: name,
	}, nil
}
