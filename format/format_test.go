package format_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hastefuI/uhkm/config"
	"github.com/hastefuI/uhkm/format"
)

func defaultCfg() config.Config { return config.Default() }

// normalizeNewlines rewrites CR and CRLF to LF so golden-file comparisons
// do not fail when Git checks out testdata with core.autocrlf=true.
func normalizeNewlines(b []byte) []byte {
	b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
	return bytes.ReplaceAll(b, []byte("\r"), []byte("\n"))
}

// TestFormatGolden runs every *.input.uhkm file through Format and
// compares the result to the corresponding *.golden.uhkm file.
func TestFormatGolden(t *testing.T) {
	inputs, err := filepath.Glob(filepath.Join("testdata", "*.input.uhkm"))
	if err != nil {
		t.Fatal(err)
	}
	if len(inputs) == 0 {
		t.Fatal("no golden input files found")
	}
	for _, input := range inputs {
		golden := strings.Replace(input, ".input.uhkm", ".golden.uhkm", 1)
		t.Run(filepath.Base(input), func(t *testing.T) {
			content, err := os.ReadFile(input)
			if err != nil {
				t.Fatalf("read input: %v", err)
			}
			expected, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("read golden: %v", err)
			}
			// Windows checkouts may rewrite testdata to CRLF; Format always emits LF.
			expected = normalizeNewlines(expected)
			got := format.Format(content, defaultCfg())
			if !bytes.Equal(got, expected) {
				t.Errorf("Format(%q):\ngot:\n%s\nwant:\n%s\ngot (quoted): %q\nwant (quoted): %q",
					filepath.Base(input), got, expected, got, expected)
			}
		})
	}
}

func TestNormalizeNewlines(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		want []byte
	}{
		{name: "lf unchanged", in: []byte("a\nb\n"), want: []byte("a\nb\n")},
		{name: "crlf", in: []byte("a\r\nb\r\n"), want: []byte("a\nb\n")},
		{name: "bare cr", in: []byte("a\rb\r"), want: []byte("a\nb\n")},
		{name: "mixed", in: []byte("a\r\nb\rc\n"), want: []byte("a\nb\nc\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeNewlines(tt.in)
			if !bytes.Equal(got, tt.want) {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// TestFormatIdempotent verifies that formatting twice produces the same result.
func TestFormatIdempotent(t *testing.T) {
	inputs, err := filepath.Glob(filepath.Join("testdata", "*.input.uhkm"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := defaultCfg()
	for _, input := range inputs {
		t.Run(filepath.Base(input), func(t *testing.T) {
			content, err := os.ReadFile(input)
			if err != nil {
				t.Fatalf("read: %v", err)
			}
			once := format.Format(content, cfg)
			twice := format.Format(once, cfg)
			if !bytes.Equal(once, twice) {
				t.Errorf("Format not idempotent for %q:\nfirst:\n%s\nsecond:\n%s",
					filepath.Base(input), once, twice)
			}
		})
	}
}

func TestIsFormatted_True(t *testing.T) {
	content := []byte("// @name: my-macro\n\naction {\n    key A\n}\n")
	if !format.IsFormatted(content, defaultCfg()) {
		t.Error("expected IsFormatted=true for already-canonical content")
	}
}

func TestIsFormatted_False(t *testing.T) {
	content := []byte("// @name:  my-macro\n\naction {\n\tkey A\n}\n") // tab indent + extra space in pragma
	if format.IsFormatted(content, defaultCfg()) {
		t.Error("expected IsFormatted=false for unformatted content")
	}
}

func TestFormatPragmaCanonical(t *testing.T) {
	input := []byte("//@version:1\n\nbody\n")
	got := format.Format(input, defaultCfg())
	if !bytes.Contains(got, []byte("// @version: 1")) {
		t.Errorf("pragma not canonicalized: %s", got)
	}
}

func TestFormatEndsWithNewline(t *testing.T) {
	input := []byte("// @name: test\n\nbody")
	got := format.Format(input, defaultCfg())
	if !bytes.HasSuffix(got, []byte("\n")) {
		t.Errorf("output does not end with newline: %q", got)
	}
}

func TestFormatOneBlankLineBetweenPreambleAndBody(t *testing.T) {
	input := []byte("// @name: test\n\n\n\nbody\n")
	got := string(format.Format(input, defaultCfg()))
	if !strings.Contains(got, "// @name: test\n\nbody") {
		t.Errorf("expected exactly one blank line between preamble and body:\n%s", got)
	}
}

// TestFormatBodyCommentsNotTreatedAsPreamble guards against a regression where
// "//" lines inside the body were mistaken for preamble, causing spurious blank
// lines to be inserted before statements that followed a body comment.
func TestFormatBodyCommentsNotTreatedAsPreamble(t *testing.T) {
	input := []byte(
		"// @name: test\n" +
			"// @version: 1\n" +
			"\n" +
			"// first body comment\n" +
			"stmtA\n" +
			"\n" +
			"// second body comment\n" +
			"stmtB\n",
	)
	got := string(format.Format(input, defaultCfg()))
	// stmtB must immediately follow its comment with no injected blank line.
	if !strings.Contains(got, "// second body comment\nstmtB\n") {
		t.Errorf("body comment caused spurious blank line before stmtB:\n%s", got)
	}
}

func TestFormatStripsTrailingWhitespace(t *testing.T) {
	input := []byte("// @name: test\n\nbody   \n")
	got := format.Format(input, defaultCfg())
	for i, line := range strings.Split(string(got), "\n") {
		if strings.TrimRight(line, " \t") != line {
			t.Errorf("line %d has trailing whitespace: %q", i+1, line)
		}
	}
}

// TestFormatBlankLineInsidePreamble guards a case where format and lint used to
// disagree. A blank line between two pragmas ended format's own notion of the
// preamble, so no blank line was inserted before the body, even though lint
// (and the specification) treat both pragmas as part of the preamble.
func TestFormatBlankLineInsidePreamble(t *testing.T) {
	input := []byte("// @uhkm-name: my-macro\n\n// @uhkm-version: 1.0.0\ntapKey enter\n")
	want := "// @uhkm-name: my-macro\n\n// @uhkm-version: 1.0.0\n\ntapKey enter\n"
	if got := string(format.Format(input, defaultCfg())); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestFormatStripsBOM(t *testing.T) {
	input := []byte("\ufeff// @uhkm-name: my-macro\n\ntapKey enter\n")
	got := format.Format(input, defaultCfg())
	if bytes.HasPrefix(got, []byte("\ufeff")) {
		t.Errorf("byte order mark not removed: %q", got)
	}
	if want := "// @uhkm-name: my-macro\n\ntapKey enter\n"; string(got) != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
