package domain_test

import (
	"errors"
	"regexp"
	"strings"
	"testing"

	"github.com/kakkky/kakkky.dev/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestGenerateSlug(t *testing.T) {
	fallbackRe := regexp.MustCompile(`^d-[0-9a-f]{6}$`)
	tests := []struct {
		name        string
		input       string
		want        string
		wantMatches *regexp.Regexp
	}{
		{
			name:  "success: lowercase ASCII passthrough",
			input: "hello",
			want:  "hello",
		},
		{
			name:  "success: uppercase to lowercase",
			input: "Hello",
			want:  "hello",
		},
		{
			name:  "success: hyphens preserved",
			input: "hello-world",
			want:  "hello-world",
		},
		{
			name:  "success: space to hyphen",
			input: "hello world",
			want:  "hello-world",
		},
		{
			name:  "success: consecutive symbols collapsed to single hyphen",
			input: "hello   world!!!",
			want:  "hello-world",
		},
		{
			name:  "success: leading/trailing symbols trimmed",
			input: "!!!hello!!!",
			want:  "hello",
		},
		{
			name:  "success: digits only",
			input: "12345",
			want:  "12345",
		},
		{
			name:  "success: hiragana to romaji",
			input: "こんにちは",
			want:  "konnnichiha",
		},
		{
			name:  "success: kanji to romaji",
			input: "世界",
			want:  "sekai",
		},
		{
			name:  "success: mixed hiragana and kanji",
			input: "こんにちは世界",
			want:  "konnnichihasekai",
		},
		{
			name:  "success: katakana with long vowel dropped",
			input: "東京タワー",
			want:  "toukyoutawa",
		},
		{
			name:  "success: mixed English and Japanese",
			input: "Go入門",
			want:  "gonyuumon",
		},
		{
			name:  "success: truncated to max length",
			input: strings.Repeat("a", domain.SlugMaxLength+5),
			want:  strings.Repeat("a", domain.SlugMaxLength),
		},
		{
			name:  "success: truncation ending in hyphen is trimmed",
			input: strings.Repeat("a", domain.SlugMaxLength-1) + "-b",
			want:  strings.Repeat("a", domain.SlugMaxLength-1),
		},
		{
			name:        "success: empty input falls back to random",
			input:       "",
			wantMatches: fallbackRe,
		},
		{
			name:        "success: symbols only falls back to random",
			input:       "!!!",
			wantMatches: fallbackRe,
		},
		{
			name:        "success: emoji only falls back to random",
			input:       "🎉🎉🎉",
			wantMatches: fallbackRe,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.GenerateSlug(tt.input)
			require.NoError(t, err)
			if tt.wantMatches != nil {
				assert.Regexp(t, tt.wantMatches, string(got))
				return
			}
			assert.Equal(t, domain.Slug(tt.want), got)
		})
	}
}
