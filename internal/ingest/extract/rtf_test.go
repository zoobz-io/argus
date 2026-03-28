package extract

import (
	"context"
	"testing"
)

func TestRTF(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain text passthrough",
			input: `{\rtf1 Hello world}`,
			want:  "Hello world",
		},
		{
			name:  "paragraph breaks",
			input: `{\rtf1 First\par Second\par Third}`,
			want:  "First\nSecond\nThird",
		},
		{
			name:  "tab control word",
			input: `{\rtf1 Column A\tab Column B}`,
			want:  "Column A\tColumn B",
		},
		{
			name:  "escaped braces",
			input: `{\rtf1 Open \{ and close \}}`,
			want:  "Open { and close }",
		},
		{
			name:  "escaped backslash",
			input: `{\rtf1 Path C:\\Users\\test}`,
			want:  `Path C:\Users\test`,
		},
		{
			name:  "font table skipped",
			input: `{\rtf1 {\fonttbl{\f0 Times New Roman;}}Hello}`,
			want:  "Hello",
		},
		{
			name:  "color table skipped",
			input: `{\rtf1 {\colortbl;\red255\green0\blue0;}Hello}`,
			want:  "Hello",
		},
		{
			name:  "stylesheet skipped",
			input: `{\rtf1 {\stylesheet{\s0 Normal;}}Hello}`,
			want:  "Hello",
		},
		{
			name:  "info group skipped",
			input: `{\rtf1 {\info{\author Test}}Hello}`,
			want:  "Hello",
		},
		{
			name:  "bold and italic formatting ignored",
			input: `{\rtf1 \b Bold\b0  and \i italic\i0  text}`,
			want:  "Bold and italic text",
		},
		{
			name:  "font size ignored",
			input: `{\rtf1 \fs24 Normal size text}`,
			want:  "Normal size text",
		},
		{
			name:  "unicode escape",
			input: `{\rtf1 caf\u233?}`,
			want:  "café",
		},
		{
			name:  "unicode negative value",
			input: `{\rtf1 \u-4064?}`,
			want:  "\uF020",
		},
		{
			name:  "line break control word",
			input: `{\rtf1 Line one\line Line two}`,
			want:  "Line one\nLine two",
		},
		{
			name:  "emdash",
			input: `{\rtf1 word\emdash word}`,
			want:  "word\u2014word",
		},
		{
			name:  "endash",
			input: `{\rtf1 1\endash 10}`,
			want:  "1\u201310",
		},
		{
			name:  "smart quotes",
			input: `{\rtf1 \lquote Hello\rquote  and \ldblquote World\rdblquote}`,
			want:  "\u2018Hello\u2019 and \u201CWorld\u201D",
		},
		{
			name:  "bullet",
			input: `{\rtf1 \bullet  Item one}`,
			want:  "\u2022 Item one",
		},
		{
			name:  "empty document",
			input: `{\rtf1 }`,
			want:  "",
		},
		{
			name:  "nested groups",
			input: `{\rtf1 Before {\b bold {\i bold-italic} bold} after}`,
			want:  "Before bold bold-italic bold after",
		},
		{
			name:  "star group skipped",
			input: `{\rtf1 {\*\generator Some Tool;}Hello}`,
			want:  "Hello",
		},
		{
			name:  "non-breaking space",
			input: `{\rtf1 word\~word}`,
			want:  "word\u00A0word",
		},
		// Edge cases for coverage.
		{
			name:  "negative depth guard",
			input: "text}}}",
			want:  "text",
		},
		{
			name:  "backslash at end of input",
			input: `{\rtf1 text\`,
			want:  "text",
		},
		{
			name:  "bare newlines stripped",
			input: "{\\rtf1 line1\nline2\rline3}",
			want:  "line1line2line3",
		},
		{
			name:  "newline after backslash",
			input: "{\\rtf1 text\\\nmore}",
			want:  "textmore",
		},
		{
			name:  "control word with negative numeric param",
			input: `{\rtf1 \li-360 text}`,
			want:  "text",
		},
		{
			name:  "unicode with no digits",
			input: `{\rtf1 \ux text}`,
			want:  "x text",
		},
		{
			name:  "unicode null codepoint falls through",
			input: `{\rtf1 \u0? text}`,
			want:  "0? text",
		},
		{
			name:  "control word base strip with negative suffix",
			input: `{\rtf1 before\par-1 after}`,
			want:  "before\nafter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RTF(ctx, []byte(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStripRTF_RawString(t *testing.T) {
	// Test the unexported function directly for edge cases.
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "bare text no wrapper",
			input: "Hello world",
			want:  "Hello world",
		},
		{
			name:  "unbalanced braces handled gracefully",
			input: `{\rtf1 unbalanced`,
			want:  "unbalanced",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripRTF(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
