// Package lint implements the UHKM lint rules.
package lint

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hastefuI/uhkm/config"
	"github.com/hastefuI/uhkm/indent"
)

// Issue represents a single lint violation.
type Issue struct {
	Code    string
	File    string
	Line    int // 1-based; 0 means file-level
	Message string
}

func (i Issue) String() string {
	if i.Line == 0 {
		return fmt.Sprintf("%s: %s: %s", i.File, i.Code, i.Message)
	}
	return fmt.Sprintf("%s:%d: %s: %s", i.File, i.Line, i.Code, i.Message)
}

// Check runs all lint rules against path with the given content and config.
func Check(path string, content []byte, cfg config.Config) []Issue {
	var issues []Issue
	issues = append(issues, checkUHKM100(path, content, cfg)...)
	issues = append(issues, checkUHKM200(path, cfg)...)
	return issues
}

// Fix attempts to automatically repair all fixable issues in content.
// Returns the (possibly modified) content and whether any change was made.
// Non-fixable issues (e.g. UHKM200) are left for the user to resolve.
func Fix(content []byte, cfg config.Config) ([]byte, bool) {
	fixed, ok := fixUHKM100(content, cfg)
	if !ok {
		return content, false
	}
	return fixed, !bytes.Equal(content, fixed)
}

// --- UHKM100: Indentation ---

func checkUHKM100(path string, content []byte, cfg config.Config) []Issue {
	var issues []Issue
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		leading := leadingWhitespace(line)
		if leading == "" {
			continue
		}
		lineNum := i + 1
		switch cfg.Lint.Indentation.Style {
		case "spaces":
			if strings.ContainsRune(leading, '\t') {
				issues = append(issues, Issue{
					Code:    "UHKM100",
					File:    path,
					Line:    lineNum,
					Message: "tabs found; expected spaces",
				})
			} else if w := cfg.Lint.Indentation.Width; w > 0 && len(leading)%w != 0 {
				issues = append(issues, Issue{
					Code:    "UHKM100",
					File:    path,
					Line:    lineNum,
					Message: fmt.Sprintf("indentation %d is not a multiple of %d", len(leading), w),
				})
			}
		case "tabs":
			if strings.ContainsRune(leading, ' ') {
				issues = append(issues, Issue{
					Code:    "UHKM100",
					File:    path,
					Line:    lineNum,
					Message: "spaces found; expected tabs",
				})
			}
		}
	}
	return issues
}

// fixUHKM100 converts indentation to the configured style.
// Returns (content, true) on success, (nil, false) if the fix cannot be applied.
func fixUHKM100(content []byte, cfg config.Config) ([]byte, bool) {
	lines := strings.Split(string(content), "\n")
	out, _, ok := indent.NormalizeLines(lines, cfg, indent.Strict)
	if !ok {
		return nil, false
	}
	return []byte(strings.Join(out, "\n")), true
}

// --- UHKM200: File Naming ---

var (
	reKebab  = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*\.uhkm$`)
	reSnake  = regexp.MustCompile(`^[a-z0-9]+(_[a-z0-9]+)*\.uhkm$`)
	rePascal = regexp.MustCompile(`^[A-Z][a-zA-Z0-9]*\.uhkm$`)
)

func checkUHKM200(path string, cfg config.Config) []Issue {
	base := filepath.Base(path)
	var ok bool
	switch cfg.Lint.Naming.Convention {
	case "kebab":
		ok = reKebab.MatchString(base)
	case "snake":
		ok = reSnake.MatchString(base)
	case "pascal":
		ok = rePascal.MatchString(base)
	default:
		ok = reKebab.MatchString(base)
	}
	if !ok {
		return []Issue{{
			Code:    "UHKM200",
			File:    path,
			Message: fmt.Sprintf("filename %q does not match %q convention", base, cfg.Lint.Naming.Convention),
		}}
	}
	return nil
}

// leadingWhitespace returns the leading whitespace prefix of s.
func leadingWhitespace(s string) string {
	for i, r := range s {
		if r != ' ' && r != '\t' {
			return s[:i]
		}
	}
	return s
}
