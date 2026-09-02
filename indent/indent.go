// Package indent centralizes indentation normalization shared by lint and format.
package indent

import (
	"strings"

	"github.com/hastefuI/uhkm/config"
)

// Mode controls how unfixable indentation is handled.
type Mode int

const (
	// BestEffort leaves unfixable lines unchanged.
	BestEffort Mode = iota
	// Strict fails when a line cannot be normalized safely.
	Strict
)

// NormalizeLines rewrites leading indentation for each line according to cfg.
// In Strict mode, it returns ok=false if any line cannot be normalized.
func NormalizeLines(lines []string, cfg config.Config, mode Mode) ([]string, bool, bool) {
	out := make([]string, len(lines))
	changed := false
	for i, line := range lines {
		normalized, lineChanged, ok := normalizeLine(line, cfg, mode)
		if !ok {
			return nil, false, false
		}
		out[i] = normalized
		changed = changed || lineChanged
	}
	return out, changed, true
}

func normalizeLine(line string, cfg config.Config, mode Mode) (string, bool, bool) {
	leading := leadingWhitespace(line)
	rest := line[len(leading):]
	if leading == "" {
		return line, false, true
	}

	tabs := strings.Count(leading, "\t")
	spaces := strings.Count(leading, " ")
	if tabs > 0 && spaces > 0 {
		if mode == Strict {
			return "", false, false
		}
		return line, false, true
	}

	width := cfg.Lint.Indentation.Width
	if width <= 0 {
		width = 4
	}

	switch cfg.Lint.Indentation.Style {
	case "spaces":
		if tabs > 0 {
			normalized := strings.Repeat(" ", tabs*width) + rest
			return normalized, normalized != line, true
		}
		return line, false, true
	case "tabs":
		if spaces > 0 {
			if spaces%width != 0 {
				if mode == Strict {
					return "", false, false
				}
				return line, false, true
			}
			normalized := strings.Repeat("\t", spaces/width) + rest
			return normalized, normalized != line, true
		}
		return line, false, true
	default:
		return line, false, true
	}
}

func leadingWhitespace(s string) string {
	for i, r := range s {
		if r != ' ' && r != '\t' {
			return s[:i]
		}
	}
	return s
}
