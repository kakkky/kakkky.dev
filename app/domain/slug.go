package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/gojp/kana"
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

type Slug string

const SlugMaxLength = 200

var (
	slugRegexp    = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	nonSlugRegexp = regexp.MustCompile(`[^a-z0-9]+`)
	jaTokenizer   = sync.OnceValue(func() *tokenizer.Tokenizer {
		t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
		if err != nil {
			panic(fmt.Errorf("kagome tokenizer init: %w", err))
		}
		return t
	})
)

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

func GenerateSlug(str string) (Slug, error) {
	s := strToRomaji(str)
	s = strings.ToLower(s)
	s = nonSlugRegexp.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > SlugMaxLength {
		s = s[:SlugMaxLength]
		s = strings.Trim(s[:SlugMaxLength], "-")
	}
	// fallback
	if s == "" {
		b := make([]byte, 3)
		_, _ = rand.Read(b)
		s = "d-" + hex.EncodeToString(b)
	}
	return NewSlug(s)
}

// strToRomaji は str を token 化し、読み (カタカナ) をローマ字に変換して連結する。
// 読みが取れない token (英数字/記号など) は Surface をそのまま使う。
func strToRomaji(str string) string {
	tokens := jaTokenizer().Tokenize(str)
	var b strings.Builder
	for _, tk := range tokens {
		reading, ok := tk.Reading()
		if !ok || reading == "" || reading == "*" {
			b.WriteString(tk.Surface)
			continue
		}
		b.WriteString(kana.KanaToRomaji(reading))
	}
	return b.String()
}
