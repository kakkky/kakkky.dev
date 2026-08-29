package domain_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kakkky/kakkky.dev/domain"
)

func TestNewTag(t *testing.T) {
	tests := []struct {
		name    string
		tagName string
		wantErr error
	}{
		{
			name:    "success: ascii name",
			tagName: "Go",
			wantErr: nil,
		},
		{
			name:    "success: name at max length",
			tagName: strings.Repeat("あ", domain.TagNameMaxLength),
			wantErr: nil,
		},
		{
			name:    "error: empty name",
			tagName: "",
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "error: name exceeds max length",
			tagName: strings.Repeat("あ", domain.TagNameMaxLength+1),
			wantErr: domain.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewTag(domain.Slug("valid-slug"), tt.tagName)
			if tt.wantErr == nil {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				return
			}
			assert.Error(t, err)
			assert.True(t, errors.Is(err, tt.wantErr))
			assert.Nil(t, got)
		})
	}
}

func TestNewTagSlug(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    domain.Slug
		wantErr error
	}{
		{
			name:  "lowercases ascii letters",
			input: "Go",
			want:  "go",
		},
		{
			name:  "replaces spaces with hyphen",
			input: "React Hooks",
			want:  "react-hooks",
		},
		{
			name:  "replaces underscore with hyphen",
			input: "next_js",
			want:  "next-js",
		},
		{
			name:  "keeps hyphen",
			input: "go-lang",
			want:  "go-lang",
		},
		{
			name:  "strips ascii symbols",
			input: "C++",
			want:  "c",
		},
		{
			name:  "preserves japanese characters",
			input: "設計",
			want:  "設計",
		},
		{
			name:  "japanese with space becomes hyphen",
			input: "設計 パターン",
			want:  "設計-パターン",
		},
		{
			name:  "compacts consecutive hyphens",
			input: "a  b",
			want:  "a-b",
		},
		{
			name:  "trims leading/trailing hyphens",
			input: "  hello  ",
			want:  "hello",
		},
		{
			name:  "truncates to SlugMaxLength runes",
			input: strings.Repeat("あ", domain.SlugMaxLength+5),
			want:  domain.Slug(strings.Repeat("あ", domain.SlugMaxLength)),
		},
		{
			name:  "keeps digits",
			input: "es2020",
			want:  "es2020",
		},
		{
			name:    "error: empty input",
			input:   "",
			wantErr: domain.ErrInvalidArgument,
		},
		{
			name:    "error: only symbols",
			input:   "!!!",
			wantErr: domain.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.NewTagSlug(tt.input)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				assert.Empty(t, got)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
