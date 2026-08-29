package domain

import (
	"fmt"
	"strings"
	"unicode"
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

// NewTagSlug は tag name から URL slug を導出する。
// article slug (NewSlug) は ASCII 限定だが tag は日本語も許可するため、
// 別関数として提供する。
//   - ASCII 英字は小文字化
//   - 空白 / `_` は `-` に変換
//   - ASCII 記号 (英数字以外) は削除
//   - Unicode letter/digit はそのまま残す
//   - 連続する `-` は 1 個に圧縮し、先頭末尾の `-` を除去
//   - 最大 SlugMaxLength rune で truncate
func NewTagSlug(name string) (Slug, error) {
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(unicode.ToLower(r))
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case unicode.IsSpace(r), r == '_', r == '-':
			b.WriteRune('-')
		case r < 128:
			// その他 ASCII 記号は削除
		default:
			// 非 ASCII (日本語含む Unicode) はそのまま残す
			b.WriteRune(r)
		}
	}

	// 連続する `-` を 1 個に圧縮
	compact := make([]rune, 0, b.Len())
	prevDash := false
	for _, r := range b.String() {
		if r == '-' {
			if prevDash {
				continue
			}
			prevDash = true
		} else {
			prevDash = false
		}
		compact = append(compact, r)
	}

	// 先頭末尾の `-` を除去
	result := strings.Trim(string(compact), "-")

	// 最大 SlugMaxLength rune で truncate
	if utf8.RuneCountInString(result) > SlugMaxLength {
		runes := []rune(result)
		result = string(runes[:SlugMaxLength])
		result = strings.TrimRight(result, "-")
	}

	if result == "" {
		return "", ErrInvalidArgument.With("タグ名 から slug を生成できませんでした")
	}
	return Slug(result), nil
}
