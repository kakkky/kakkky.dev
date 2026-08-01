package domain

import (
	"fmt"
	"regexp"
)

type Slug string

const SlugMaxLength = 20

var slugRegexp = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func NewSlug(slug string) (Slug, error) {
	if slug == "" {
		return "", ErrInvalidArgument.With("slug は 必須 です")
	}
	if len(slug) > SlugMaxLength {
		return "", ErrInvalidArgument.With(fmt.Sprintf("slug は %d 文字以内 です", SlugMaxLength))
	}
	if !slugRegexp.MatchString(slug) {
		return "", ErrInvalidArgument.With("slug は 英小文字、数字、ハイフン(-) のみ です")
	}

	return Slug(slug), nil
}
