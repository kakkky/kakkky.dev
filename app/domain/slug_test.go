package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/stretchr/testify/assert"
)

func TestNewSlug(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "success: lowercase letters",
			input:   "hello",
			wantErr: nil,
		},
		{
			name:    "success: digits only",
			input:   "12345",
			wantErr: nil,
		},
		{
			name:    "success: letters, digits, and hyphens",
			input:   "hello-world-123",
			wantErr: nil,
		},
		{
			name:    "success: at max length",
			input:   strings.Repeat("a", domain.SlugMaxLength),
			wantErr: nil,
		},
		{
			name:    "error: empty",
			input:   "",
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "error: exceeds max length",
			input:   strings.Repeat("a", domain.SlugMaxLength+1),
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "error: uppercase letters",
			input:   "Hello",
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "error: contains space",
			input:   "hello world",
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "error: contains underscore",
			input:   "hello_world",
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "error: contains slash",
			input:   "hello/world",
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "error: contains multibyte character",
			input:   "こんにちは",
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "error: leading hyphen",
			input:   "-hello",
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "error: trailing hyphen",
			input:   "hello-",
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "error: consecutive hyphens",
			input:   "hello--world",
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "error: only hyphens",
			input:   "---",
			wantErr: domain.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewSlug(tt.input)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, domain.Slug(tt.input), got)
				return
			}
			assert.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantErr))
			assert.Empty(t, got)
		})
	}
}

func TestDeriveSlug(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    domain.Slug
		wantAny bool // fallback ランダム slug なので値は問わない
	}{
		{
			name:  "lowercase ascii kept",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "uppercase folded to lowercase",
			input: "Hello",
			want:  "hello",
		},
		{
			name:  "space becomes hyphen",
			input: "Hello World",
			want:  "hello-world",
		},
		{
			name:  "underscore becomes hyphen",
			input: "next_js",
			want:  "next-js",
		},
		{
			name:  "japanese converted to romaji, ascii kept",
			input: "Go における Context",
			// kagome tokenize: "Go", "における", "Context" → "go", "niokeru", "context"
			want: "go-niokeru-context",
		},
		{
			name:  "pure katakana to romaji",
			input: "テスト",
			want:  "tesuto",
		},
		{
			name:  "long vowel repeats previous vowel",
			input: "ケーキ",
			want:  "keeki",
		},
		{
			name:  "small tsu doubles next consonant",
			input: "キット",
			want:  "kitto",
		},
		{
			name:  "contraction",
			input: "シャツ",
			want:  "shatsu",
		},
		{
			name:  "hiragana kept via katakana normalization",
			input: "きょう",
			want:  "kyou",
		},
		{
			name:  "symbols dropped",
			input: "C++ / Rust!",
			want:  "c-rust",
		},
		{
			name:  "consecutive spaces compact to single hyphen",
			input: "a   b",
			want:  "a-b",
		},
		{
			name:  "trim leading/trailing hyphens and spaces",
			input: "  hello  ",
			want:  "hello",
		},
		{
			name:  "truncate to SlugMaxLength",
			input: strings.Repeat("a", domain.SlugMaxLength+5),
			want:  domain.Slug(strings.Repeat("a", domain.SlugMaxLength)),
		},
		{
			name:  "pure japanese to romaji",
			input: "こんにちは",
			want:  "konnichiha",
		},
		{
			name:    "empty title falls back to random slug",
			input:   "",
			wantAny: true,
		},
		{
			name:    "symbols only falls back to random slug",
			input:   "!!!///",
			wantAny: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.DeriveSlug(tt.input)

			// 返り値は必ず NewSlug の検証を通過する
			_, err := domain.NewSlug(string(got))
			assert.NoError(t, err)

			if tt.wantAny {
				assert.NotEmpty(t, got)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
