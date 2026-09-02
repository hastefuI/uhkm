// Package format implements canonical formatting for .uhkm files.
package format

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/hastefuI/uhkm/config"
	"github.com/hastefuI/uhkm/indent"
)

var rePragma = regexp.MustCompile(`^//\s*@(\w+)\s*:\s*(.+?)\s*$`)

// Format applies canonical formatting to content and returns the result.
//
// Canonical form:
//   - Pragmas are normalized to "// @key: value".
//   - Trailing whitespace is stripped from every line.
//   - CRLF line endings are converted to LF.
//   - Exactly one blank line separates the preamble from the body.
//   - The file ends with exactly one newline.
//   - Indentation is normalized to the configured style.
func Format(content []byte, cfg config.Config) []byte {
	lines := strings.Split(string(content), "\n")

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
		if m := rePragma.FindStringSubmatch(line); m != nil {
			line = "// @" + m[1] + ": " + m[2] // canonicalize pragma
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
// and the body. The preamble is the initial contiguous block of comment lines
// (lines starting with "//") at the top of the file. Body comments that
// appear after the first non-comment line are not part of the preamble.
func separatePreambleBody(lines []string) []string {
	// Walk from the top: the preamble ends at the last consecutive "//" line
	// before any non-comment line is encountered.
	preambleEnd := -1
	for i, l := range lines {
		if strings.HasPrefix(l, "//") {
			preambleEnd = i
		} else {
			break
		}
	}
	if preambleEnd < 0 || preambleEnd >= len(lines)-1 {
		return lines
	}

	// Collect the body: skip blank lines immediately after the preamble.
	rest := lines[preambleEnd+1:]
	for len(rest) > 0 && strings.TrimSpace(rest[0]) == "" {
		rest = rest[1:]
	}
	if len(rest) == 0 {
		return lines[:preambleEnd+1]
	}

	out := make([]string, 0, preambleEnd+2+len(rest))
	out = append(out, lines[:preambleEnd+1]...)
	out = append(out, "") // exactly one blank line
	out = append(out, rest...)
	return out
}
