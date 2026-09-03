// Package preamble parses the pragma preamble of a .uhkm file.
//
// A preamble is the leading block of "// @key: value" pragma lines, plain
// "//" comments, and blank lines at the top of a file. It ends at the first
// line that is none of those. Tooling honours only the pragmas inside that
// region; pragmas below it are ignored.
//
// The package also owns the byte order mark helpers, since a BOM sits ahead
// of the preamble and would otherwise hide the first pragma from every
// caller.
package preamble

import (
	"bytes"
	"regexp"
	"strings"
)

// BOM is the UTF-8 byte order mark. The specification requires .uhkm files to
// be UTF-8 encoded without one.
const BOM = "\ufeff"

// rePragma matches a pragma line, tolerating the spacing variations that
// formatting normalizes away. Keys may contain hyphens, as in "uhkm-name".
var rePragma = regexp.MustCompile(`^//[ \t]*@([A-Za-z0-9_-]+)[ \t]*:(.*)$`)

// Pragma is a single "// @key: value" line.
type Pragma struct {
	Key        string // key without the leading "@", for example "uhkm-name"
	Value      string // value with surrounding whitespace trimmed
	Line       int    // 1-based line number
	InPreamble bool   // whether the pragma sits inside the preamble region
}

// HasBOM reports whether content begins with a UTF-8 byte order mark.
func HasBOM(content []byte) bool {
	return bytes.HasPrefix(content, []byte(BOM))
}

// TrimBOM returns content without a leading UTF-8 byte order mark.
func TrimBOM(content []byte) []byte {
	return bytes.TrimPrefix(content, []byte(BOM))
}

// ParseLine reports whether line is a pragma and returns its key and value.
// The value is empty for a pragma written without one ("// @uhkm-name:").
func ParseLine(line string) (key, value string, ok bool) {
	m := rePragma.FindStringSubmatch(line)
	if m == nil {
		return "", "", false
	}
	return m[1], strings.TrimSpace(m[2]), true
}

// Canonical rewrites a pragma line as "// @key: value" and reports whether
// line was a pragma at all. A valueless pragma canonicalizes to "// @key:"
// so that canonicalization never introduces trailing whitespace.
func Canonical(line string) (string, bool) {
	key, value, ok := ParseLine(line)
	if !ok {
		return line, false
	}
	if value == "" {
		return "// @" + key + ":", true
	}
	return "// @" + key + ": " + value, true
}

// Parse returns every pragma in content, in file order. A leading byte order
// mark is ignored so that it cannot mask the first pragma.
func Parse(content []byte) []Pragma {
	return ParseLines(strings.Split(string(TrimBOM(content)), "\n"))
}

// ParseLines is Parse for content that is already split into lines.
func ParseLines(lines []string) []Pragma {
	end := regionEnd(lines)

	var pragmas []Pragma
	for i, line := range lines {
		key, value, ok := ParseLine(line)
		if !ok {
			continue
		}
		pragmas = append(pragmas, Pragma{
			Key:        key,
			Value:      value,
			Line:       i + 1,
			InPreamble: i < end,
		})
	}
	return pragmas
}

// regionEnd returns the number of leading lines that form the preamble
// region: pragmas, plain "//" comments, and blank lines.
//
// Surrounding whitespace is ignored when classifying a line, so an indented
// comment does not end the region. An indented pragma is still not a pragma,
// because the specification anchors "//" to the start of the line.
func regionEnd(lines []string) int {
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		return i
	}
	return len(lines)
}
