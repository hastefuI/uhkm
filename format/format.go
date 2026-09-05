// Package format implements canonical formatting for .uhkm files.
package format

import (
	"bytes"
	"strings"

	"go.hasteful.org/uhkm/config"
	"go.hasteful.org/uhkm/indent"
	"go.hasteful.org/uhkm/preamble"
)

// Format applies canonical formatting to content and returns the result.
//
// Canonical form:
//   - A leading UTF-8 byte order mark is removed.
//   - Pragmas are normalized to "// @key: value".
//   - Trailing whitespace is stripped from every line.
//   - CRLF line endings are converted to LF.
//   - Exactly one blank line separates the preamble from the body.
//   - The file ends with exactly one newline.
//   - Indentation is normalized to the configured style.
func Format(content []byte, cfg config.Config) []byte {
	lines := strings.Split(string(preamble.TrimBOM(content)), "\n")

	// Normalize CRLF → LF.
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, "\r")
	}

	// Strip trailing empty lines.
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}

	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimRight(line, " \t") // strip trailing whitespace
		if canonical, ok := preamble.Canonical(line); ok {
			line = canonical // canonicalize pragma
		}
		out = append(out, line)
	}

	out = normalizeIndentation(out, cfg)
	out = separatePreambleBody(out)

	return []byte(strings.Join(out, "\n") + "\n")
}

// IsFormatted reports whether content is already in canonical form.
func IsFormatted(content []byte, cfg config.Config) bool {
	return bytes.Equal(content, Format(content, cfg))
}

// normalizeIndentation rewrites leading whitespace on each line to match cfg.
func normalizeIndentation(lines []string, cfg config.Config) []string {
	out, _, _ := indent.NormalizeLines(lines, cfg, indent.BestEffort)
	return out
}

// separatePreambleBody ensures exactly one blank line between the preamble
// and the body.
//
// The preamble ends at the last pragma inside the region that the preamble
// package identifies, not at the end of that region: a plain comment trailing
// the pragmas introduces the body (for example a comment documenting the first
// statement), so it stays attached to what follows. A file with no pragma has
// no preamble to separate.
func separatePreambleBody(lines []string) []string {
	end := -1
	for _, p := range preamble.ParseLines(lines) {
		if p.InPreamble {
			end = p.Line - 1 // Pragma.Line is 1-based
		}
	}
	if end < 0 {
		return lines
	}

	head, body := lines[:end+1], lines[end+1:]
	for len(body) > 0 && strings.TrimSpace(body[0]) == "" {
		body = body[1:]
	}
	if len(body) == 0 {
		return head
	}

	out := make([]string, 0, len(head)+1+len(body))
	out = append(out, head...)
	out = append(out, "") // exactly one blank line
	out = append(out, body...)
	return out
}
