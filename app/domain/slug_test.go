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
