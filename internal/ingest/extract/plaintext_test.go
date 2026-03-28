package extract

import (
	"context"
	"testing"
)

func TestPlaintext(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple text",
			input: "hello world",
			want:  "hello world",
		},
		{
			name:  "trims whitespace",
			input: "  hello world  \n\n",
			want:  "hello world",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
		{
			name:  "multiline preserves inner newlines",
			input: "line one\nline two\nline three",
			want:  "line one\nline two\nline three",
		},
		{
			name:  "unicode content",
			input: "日本語テスト",
			want:  "日本語テスト",
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Plaintext(ctx, []byte(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
