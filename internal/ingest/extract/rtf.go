package extract

import (
	"context"
	"strings"
	"unicode/utf8"
)

// RTF extracts text from RTF documents by stripping control words,
// groups, and formatting codes. Handles standard RTF constructs including
// escaped characters, unicode sequences, and paragraph breaks.
func RTF(_ context.Context, data []byte) (string, error) {
	return stripRTF(string(data)), nil
}

// stripRTF removes RTF control structures and returns plain text.
func stripRTF(input string) string {
	var out strings.Builder
	out.Grow(len(input) / 2)

	depth := 0
	skipDepth := -1 // depth at which skip started, -1 = not skipping
	i := 0

	for i < len(input) {
		ch := input[i]

		switch ch {
		case '{':
			depth++
			// Skip known non-text groups.
			if skipDepth < 0 && i+1 < len(input) && input[i+1] == '\\' {
				rest := input[i+1:]
				for _, prefix := range skipPrefixes {
					if strings.HasPrefix(rest, prefix) {
						skipDepth = depth
						break
					}
				}
			}
			i++

		case '}':
			if skipDepth == depth {
				skipDepth = -1
			}
			depth--
			if depth < 0 {
				depth = 0
			}
			i++

		case '\\':
			if skipDepth >= 0 {
				i++
				skipControlWord(input, &i)
				continue
			}
			i++
			if i >= len(input) {
				break
			}
			next := input[i]
			switch next {
			case '\\', '{', '}':
				out.WriteByte(next)
				i++
			case '~':
				out.WriteString("\u00A0")
				i++
			case 'u':
				// Unicode escape: \uN?
				i++
				r, end := parseUnicode(input, i)
				if r > 0 {
					out.WriteRune(r)
					i = end
					// Skip replacement character.
					if i < len(input) && input[i] == '?' {
						i++
					}
				}
			case '\n', '\r':
				i++
			default:
				word, end := readControlWord(input, i)
				i = end
				if text, ok := controlWordText(word); ok {
					out.WriteString(text)
				}
				// Skip optional space delimiter after control word.
				if i < len(input) && input[i] == ' ' {
					i++
				}
			}

		case '\n', '\r':
			i++

		default:
			if skipDepth < 0 {
				out.WriteByte(ch)
			}
			i++
		}
	}

	return strings.TrimSpace(out.String())
}

// skipPrefixes are RTF group prefixes whose content should be skipped entirely.
var skipPrefixes = []string{
	"\\fonttbl", "\\colortbl", "\\stylesheet", "\\info",
	"\\header", "\\footer", "\\pict", "\\*",
}

// readControlWord reads an alphabetic control word starting at pos.
func readControlWord(input string, pos int) (string, int) {
	start := pos
	for pos < len(input) && ((input[pos] >= 'a' && input[pos] <= 'z') || (input[pos] >= 'A' && input[pos] <= 'Z')) {
		pos++
	}
	// Skip optional numeric parameter.
	if pos < len(input) && (input[pos] == '-' || (input[pos] >= '0' && input[pos] <= '9')) {
		if input[pos] == '-' {
			pos++
		}
		for pos < len(input) && input[pos] >= '0' && input[pos] <= '9' {
			pos++
		}
	}
	return input[start:pos], pos
}

// skipControlWord advances past a control word without capturing it.
func skipControlWord(input string, pos *int) {
	for *pos < len(input) && input[*pos] != ' ' && input[*pos] != '\\' && input[*pos] != '{' && input[*pos] != '}' {
		*pos++
	}
	if *pos < len(input) && input[*pos] == ' ' {
		*pos++
	}
}

// parseUnicode parses a decimal code point after \u.
func parseUnicode(input string, pos int) (rune, int) {
	neg := false
	if pos < len(input) && input[pos] == '-' {
		neg = true
		pos++
	}
	start := pos
	for pos < len(input) && input[pos] >= '0' && input[pos] <= '9' {
		pos++
	}
	if pos == start {
		return 0, start
	}
	n := 0
	for _, ch := range input[start:pos] {
		n = n*10 + int(ch-'0')
	}
	if neg {
		n = 65536 - n
	}
	r := rune(n)
	if r == 0 || !utf8.ValidRune(r) {
		return 0, start
	}
	return r, pos
}

// controlWordText maps RTF control words to their text equivalents.
func controlWordText(word string) (string, bool) {
	// Strip numeric suffix for matching.
	base := word
	for i := len(word) - 1; i >= 0; i-- {
		if word[i] < '0' || word[i] > '9' {
			if word[i] == '-' && i > 0 {
				base = word[:i]
			} else {
				base = word[:i+1]
			}
			break
		}
	}

	switch base {
	case "par", "line":
		return "\n", true
	case xmlElemTab:
		return "\t", true
	case "emdash":
		return "\u2014", true
	case "endash":
		return "\u2013", true
	case "lquote":
		return "\u2018", true
	case "rquote":
		return "\u2019", true
	case "ldblquote":
		return "\u201C", true
	case "rdblquote":
		return "\u201D", true
	case "bullet":
		return "\u2022", true
	}
	return "", false
}
