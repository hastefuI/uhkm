package preamble_test

import (
	"testing"

	"github.com/hastefuI/uhkm/preamble"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantKey   string
		wantValue string
		wantOK    bool
	}{
		{name: "canonical", line: "// @uhkm-name: my-macro", wantKey: "uhkm-name", wantValue: "my-macro", wantOK: true},
		{name: "no space after slashes", line: "//@uhkm-name:my-macro", wantKey: "uhkm-name", wantValue: "my-macro", wantOK: true},
		{name: "extra spacing", line: "//   @uhkm-name  :   my-macro   ", wantKey: "uhkm-name", wantValue: "my-macro", wantOK: true},
		{name: "underscore key", line: "// @uhkm_name: x", wantKey: "uhkm_name", wantValue: "x", wantOK: true},
		{name: "empty value", line: "// @uhkm-name:", wantKey: "uhkm-name", wantValue: "", wantOK: true},
		{name: "whitespace value", line: "// @uhkm-name:    ", wantKey: "uhkm-name", wantValue: "", wantOK: true},
		{name: "carriage return trimmed", line: "// @uhkm-name: my-macro\r", wantKey: "uhkm-name", wantValue: "my-macro", wantOK: true},
		{name: "value keeps inner spaces", line: "// @uhkm-author: A Name <a@b.c>", wantKey: "uhkm-author", wantValue: "A Name <a@b.c>", wantOK: true},
		{name: "plain comment", line: "// not a pragma"},
		{name: "no colon", line: "// @uhkm-name"},
		{name: "indented", line: "    // @uhkm-name: x"},
		{name: "statement", line: "tapKey enter"},
		{name: "blank", line: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, value, ok := preamble.ParseLine(tt.line)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if key != tt.wantKey || value != tt.wantValue {
				t.Errorf("got (%q, %q), want (%q, %q)", key, value, tt.wantKey, tt.wantValue)
			}
		})
	}
}

func TestCanonical(t *testing.T) {
	tests := []struct {
		name   string
		line   string
		want   string
		wantOK bool
	}{
		{name: "tightens spacing", line: "//@uhkm-name:   my-macro  ", want: "// @uhkm-name: my-macro", wantOK: true},
		{name: "already canonical", line: "// @uhkm-version: 1.0.0", want: "// @uhkm-version: 1.0.0", wantOK: true},
		{name: "empty value has no trailing space", line: "//  @uhkm-name:  ", want: "// @uhkm-name:", wantOK: true},
		{name: "non-pragma untouched", line: "tapKey enter", want: "tapKey enter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := preamble.Canonical(tt.line)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("got (%q, %v), want (%q, %v)", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestParseRegion(t *testing.T) {
	content := []byte(
		"// a plain comment\n" +
			"// @uhkm-name: my-macro\n" +
			"\n" +
			"// @uhkm-version: 1.0.0\n" +
			"\n" +
			"tapKey enter\n" +
			"// @uhkm-os: linux\n",
	)

	got := preamble.Parse(content)
	want := []preamble.Pragma{
		{Key: "uhkm-name", Value: "my-macro", Line: 2, InPreamble: true},
		{Key: "uhkm-version", Value: "1.0.0", Line: 4, InPreamble: true},
		{Key: "uhkm-os", Value: "linux", Line: 7, InPreamble: false},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d pragmas (%+v), want %d", len(got), got, len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("pragma %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestParseCommentOnlyFileIsAllPreamble(t *testing.T) {
	content := []byte("// @uhkm-name: my-macro\n// @uhkm-version: 1.0.0\n")
	for _, p := range preamble.Parse(content) {
		if !p.InPreamble {
			t.Errorf("pragma %q on line %d: InPreamble = false, want true", p.Key, p.Line)
		}
	}
}

func TestParseNoPragmas(t *testing.T) {
	if got := preamble.Parse([]byte("tapKey enter\n")); got != nil {
		t.Errorf("got %+v, want nil", got)
	}
}

func TestBOMHelpers(t *testing.T) {
	withBOM := []byte("\ufeff// @uhkm-name: x\n")
	if !preamble.HasBOM(withBOM) {
		t.Error("HasBOM = false, want true")
	}
	without := preamble.TrimBOM(withBOM)
	if preamble.HasBOM(without) {
		t.Error("TrimBOM left a byte order mark")
	}
	if got := string(without); got != "// @uhkm-name: x\n" {
		t.Errorf("got %q", got)
	}
	if got := preamble.TrimBOM(without); string(got) != string(without) {
		t.Error("TrimBOM changed content that has no byte order mark")
	}
}

// A byte order mark must not hide the first pragma.
func TestParseIgnoresBOM(t *testing.T) {
	got := preamble.Parse([]byte("\ufeff// @uhkm-name: x\n// @uhkm-version: 1.0.0\n"))
	if len(got) != 2 {
		t.Fatalf("got %d pragmas (%+v), want 2", len(got), got)
	}
	for _, p := range got {
		if !p.InPreamble {
			t.Errorf("pragma %q: InPreamble = false, want true", p.Key)
		}
	}
}

// An indented comment is still a comment, so it does not end the region.
func TestParseIndentedCommentStaysInRegion(t *testing.T) {
	got := preamble.Parse([]byte("    // a note\n// @uhkm-name: x\n\ntapKey enter\n"))
	if len(got) != 1 {
		t.Fatalf("got %d pragmas (%+v), want 1", len(got), got)
	}
	if !got[0].InPreamble {
		t.Error("indented comment ended the preamble region")
	}
}
