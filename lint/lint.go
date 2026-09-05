// Package lint implements the UHKM lint rules.
package lint

import (
	"bytes"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"go.hasteful.org/uhkm/config"
	"go.hasteful.org/uhkm/indent"
	"go.hasteful.org/uhkm/preamble"
)

// Severity classifies how serious an issue is. The zero value is
// SeverityError, so a rule opts in to warning level explicitly.
type Severity int

const (
	// SeverityError marks a violation of a hard specification requirement.
	SeverityError Severity = iota
	// SeverityWarning marks a violation that tooling can still work around,
	// such as a pragma key the specification does not define.
	SeverityWarning
)

// String returns the lowercase name of the severity.
func (s Severity) String() string {
	if s == SeverityWarning {
		return "warning"
	}
	return "error"
}

// Issue represents a single lint violation.
type Issue struct {
	Code     string
	Severity Severity
	File     string
	Line     int // 1-based; 0 means file-level
	Message  string
}

func (i Issue) String() string {
	if i.Line == 0 {
		return fmt.Sprintf("%s: %s: %s", i.File, i.Code, i.Message)
	}
	return fmt.Sprintf("%s:%d: %s: %s", i.File, i.Line, i.Code, i.Message)
}

// Check runs all lint rules against path with the given content and config.
// Rules are reported in code order: UHKM100 (indentation), UHKM200-UHKM201
// (file naming), UHKM300-UHKM305 (preamble and pragmas), then UHKM400
// (encoding).
func Check(path string, content []byte, cfg config.Config) []Issue {
	// A byte order mark is reported once, by UHKM400. Every other rule sees the
	// file as the specification says it should have been written, so that a BOM
	// cannot cascade into misleading "missing pragma" errors.
	body := preamble.TrimBOM(content)

	var issues []Issue
	issues = append(issues, checkUHKM100(path, body, cfg)...)
	issues = append(issues, checkUHKM200(path, cfg)...)
	issues = append(issues, checkUHKM201(path, body)...)
	issues = append(issues, checkPragmas(path, body)...)
	issues = append(issues, checkUHKM400(path, content)...)
	return issues
}

// Fix attempts to automatically repair all fixable issues in content.
// Returns the (possibly modified) content and whether any change was made.
// Non-fixable issues (e.g. UHKM200) are left for the user to resolve.
func Fix(content []byte, cfg config.Config) ([]byte, bool) {
	fixed := preamble.TrimBOM(content) // UHKM400
	if indented, ok := fixUHKM100(fixed, cfg); ok {
		fixed = indented
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
					Code:     "UHKM100",
					Severity: SeverityError,
					File:     path,
					Line:     lineNum,
					Message:  "tabs found; expected spaces",
				})
			} else if w := cfg.Lint.Indentation.Width; w > 0 && len(leading)%w != 0 {
				issues = append(issues, Issue{
					Code:     "UHKM100",
					Severity: SeverityError,
					File:     path,
					Line:     lineNum,
					Message:  fmt.Sprintf("indentation %d is not a multiple of %d", len(leading), w),
				})
			}
		case "tabs":
			if strings.ContainsRune(leading, ' ') {
				issues = append(issues, Issue{
					Code:     "UHKM100",
					Severity: SeverityError,
					File:     path,
					Line:     lineNum,
					Message:  "spaces found; expected tabs",
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
			Code:     "UHKM200",
			Severity: SeverityError,
			File:     path,
			Message:  fmt.Sprintf("filename %q does not match %q convention", base, cfg.Lint.Naming.Convention),
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

// --- UHKM201: Filename comment ---

// checkUHKM201 reports a missing first-line filename comment. The
// specification requires the first line of a .uhkm file to be a comment
// containing the filename, so that the name survives when a macro is viewed
// outside its original file. A pragma does not satisfy the rule: the filename
// comment precedes the preamble.
func checkUHKM201(path string, content []byte) []Issue {
	base := filepath.Base(path)
	first, _, _ := strings.Cut(string(content), "\n")
	first = strings.TrimSpace(first)

	if _, _, isPragma := preamble.ParseLine(first); !isPragma &&
		strings.HasPrefix(first, "//") && strings.Contains(first, base) {
		return nil
	}
	return []Issue{{
		Code:     "UHKM201",
		Severity: SeverityError,
		File:     path,
		Line:     1,
		Message:  fmt.Sprintf("first line must be a comment containing the filename %q", base),
	}}
}

// --- UHKM400: File encoding ---

// checkUHKM400 reports a leading UTF-8 byte order mark. The specification
// requires .uhkm files to be UTF-8 encoded with no BOM.
func checkUHKM400(path string, content []byte) []Issue {
	if !preamble.HasBOM(content) {
		return nil
	}
	return []Issue{{
		Code:     "UHKM400",
		Severity: SeverityError,
		File:     path,
		Line:     1,
		Message:  "file begins with a UTF-8 byte order mark; .uhkm files must be UTF-8 with no BOM",
	}}
}
