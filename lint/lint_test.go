package lint_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"go.hasteful.org/uhkm/config"
	"go.hasteful.org/uhkm/lint"
)

// --- UHKM100 ---

func TestCheckUHKM100_SpacesValid(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "spaces-valid.uhkm"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Default() // spaces, width=4
	issues := lint.Check("testdata/spaces-valid.uhkm", content, cfg)
	for _, iss := range issues {
		if iss.Code == "UHKM100" {
			t.Errorf("unexpected UHKM100: %s", iss)
		}
	}
}

func TestCheckUHKM100_TabsValid(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "tabs-valid.uhkm"))
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		Lint: config.Lint{
			Indentation: config.Indentation{Style: "tabs"},
			Naming:      config.Naming{Convention: "kebab"},
		},
	}
	issues := lint.Check("testdata/tabs-valid.uhkm", content, cfg)
	for _, iss := range issues {
		if iss.Code == "UHKM100" {
			t.Errorf("unexpected UHKM100: %s", iss)
		}
	}
}

func TestCheckUHKM100_TabsInSpacesMode(t *testing.T) {
	content := []byte("action {\n\tkey A\n}\n")
	cfg := config.Default() // expects spaces
	issues := lint.Check("f.uhkm", content, cfg)
	if !hasCode(issues, "UHKM100") {
		t.Error("expected UHKM100 for tab in spaces mode, got none")
	}
}

func TestCheckUHKM100_SpacesInTabsMode(t *testing.T) {
	content := []byte("action {\n    key A\n}\n")
	cfg := config.Config{
		Lint: config.Lint{
			Indentation: config.Indentation{Style: "tabs"},
			Naming:      config.Naming{Convention: "kebab"},
		},
	}
	issues := lint.Check("f.uhkm", content, cfg)
	if !hasCode(issues, "UHKM100") {
		t.Error("expected UHKM100 for space in tabs mode, got none")
	}
}

func TestCheckUHKM100_NotMultipleOfWidth(t *testing.T) {
	// 3-space indent with width=4 is a violation.
	content := []byte("action {\n   key A\n}\n")
	cfg := config.Default() // width=4
	issues := lint.Check("f.uhkm", content, cfg)
	if !hasCode(issues, "UHKM100") {
		t.Error("expected UHKM100 for non-multiple indentation, got none")
	}
}

func TestCheckUHKM100_EmptyLines(t *testing.T) {
	content := []byte("action {\n\n    key A\n}\n")
	cfg := config.Default()
	issues := lint.Check("my-macro.uhkm", content, cfg)
	for _, iss := range issues {
		if iss.Code == "UHKM100" {
			t.Errorf("unexpected UHKM100 on blank line: %s", iss)
		}
	}
}

// --- Fix UHKM100 ---

func TestFixUHKM100_TabsToSpaces(t *testing.T) {
	input := []byte("action {\n\tkey A\n\tkey B\n}\n")
	cfg := config.Default() // spaces, width=4
	fixed, changed := lint.Fix(input, cfg)
	if !changed {
		t.Fatal("expected Fix to report a change")
	}
	want := "action {\n    key A\n    key B\n}\n"
	if string(fixed) != want {
		t.Errorf("got %q, want %q", fixed, want)
	}
}

func TestFixUHKM100_SpacesToTabs(t *testing.T) {
	input := []byte("action {\n    key A\n    key B\n}\n")
	cfg := config.Config{
		Lint: config.Lint{
			Indentation: config.Indentation{Style: "tabs", Width: 4},
			Naming:      config.Naming{Convention: "kebab"},
		},
	}
	fixed, changed := lint.Fix(input, cfg)
	if !changed {
		t.Fatal("expected Fix to report a change")
	}
	want := "action {\n\tkey A\n\tkey B\n}\n"
	if string(fixed) != want {
		t.Errorf("got %q, want %q", fixed, want)
	}
}

func TestFixUHKM100_MixedCannotFix(t *testing.T) {
	// Mixed tabs and spaces on the same line cannot be fixed cleanly.
	input := []byte("action {\n\t   key A\n}\n")
	cfg := config.Default()
	_, changed := lint.Fix(input, cfg)
	if changed {
		t.Error("expected Fix to report no change for mixed indentation")
	}
}

func TestFixUHKM100_AlreadyCorrect(t *testing.T) {
	input := []byte("action {\n    key A\n}\n")
	cfg := config.Default() // spaces, width=4
	_, changed := lint.Fix(input, cfg)
	if changed {
		t.Error("expected Fix to report no change for already-correct content")
	}
}

// --- UHKM200 ---

func TestCheckUHKM200_KebabValid(t *testing.T) {
	cases := []string{"my-macro.uhkm", "foo.uhkm", "a1-b2.uhkm"}
	cfg := config.Default()
	for _, name := range cases {
		issues := lint.Check(name, []byte{}, cfg)
		for _, iss := range issues {
			if iss.Code == "UHKM200" {
				t.Errorf("%q: unexpected UHKM200", name)
			}
		}
	}
}

func TestCheckUHKM200_KebabInvalid(t *testing.T) {
	cases := []string{"MyMacro.uhkm", "my_macro.uhkm", "MY-MACRO.uhkm", "macro.txt"}
	cfg := config.Default()
	for _, name := range cases {
		issues := lint.Check(name, []byte{}, cfg)
		if !hasCode(issues, "UHKM200") {
			t.Errorf("%q: expected UHKM200, got none", name)
		}
	}
}

func TestCheckUHKM200_SnakeValid(t *testing.T) {
	cases := []string{"my_macro.uhkm", "foo.uhkm", "a1_b2.uhkm"}
	cfg := config.Config{
		Lint: config.Lint{
			Indentation: config.Indentation{Style: "spaces", Width: 4},
			Naming:      config.Naming{Convention: "snake"},
		},
	}
	for _, name := range cases {
		issues := lint.Check(name, []byte{}, cfg)
		for _, iss := range issues {
			if iss.Code == "UHKM200" {
				t.Errorf("%q: unexpected UHKM200", name)
			}
		}
	}
}

func TestCheckUHKM200_SnakeInvalid(t *testing.T) {
	cases := []string{"my-macro.uhkm", "MyMacro.uhkm"}
	cfg := config.Config{
		Lint: config.Lint{
			Indentation: config.Indentation{Style: "spaces", Width: 4},
			Naming:      config.Naming{Convention: "snake"},
		},
	}
	for _, name := range cases {
		issues := lint.Check(name, []byte{}, cfg)
		if !hasCode(issues, "UHKM200") {
			t.Errorf("%q: expected UHKM200, got none", name)
		}
	}
}

func TestCheckUHKM200_PascalValid(t *testing.T) {
	cases := []string{"MyMacro.uhkm", "Foo.uhkm", "ABC.uhkm"}
	cfg := config.Config{
		Lint: config.Lint{
			Indentation: config.Indentation{Style: "spaces", Width: 4},
			Naming:      config.Naming{Convention: "pascal"},
		},
	}
	for _, name := range cases {
		issues := lint.Check(name, []byte{}, cfg)
		for _, iss := range issues {
			if iss.Code == "UHKM200" {
				t.Errorf("%q: unexpected UHKM200", name)
			}
		}
	}
}

func TestCheckUHKM200_PascalInvalid(t *testing.T) {
	cases := []string{"my-macro.uhkm", "myMacro.uhkm", "my_macro.uhkm"}
	cfg := config.Config{
		Lint: config.Lint{
			Indentation: config.Indentation{Style: "spaces", Width: 4},
			Naming:      config.Naming{Convention: "pascal"},
		},
	}
	for _, name := range cases {
		issues := lint.Check(name, []byte{}, cfg)
		if !hasCode(issues, "UHKM200") {
			t.Errorf("%q: expected UHKM200, got none", name)
		}
	}
}

func TestIssueString_WithLine(t *testing.T) {
	iss := lint.Issue{Code: "UHKM100", File: "f.uhkm", Line: 3, Message: "tabs found"}
	want := "f.uhkm:3: UHKM100: tabs found"
	if got := iss.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIssueString_FileLevel(t *testing.T) {
	iss := lint.Issue{Code: "UHKM200", File: "f.uhkm", Line: 0, Message: `filename "f.uhkm" does not match "kebab" convention`}
	want := `f.uhkm: UHKM200: filename "f.uhkm" does not match "kebab" convention`
	if got := iss.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func hasCode(issues []lint.Issue, code string) bool {
	for _, iss := range issues {
		if iss.Code == code {
			return true
		}
	}
	return false
}

// --- UHKM400 ---

func TestCheckUHKM400_BOM(t *testing.T) {
	content := []byte("\ufeff// bom.uhkm\n// @uhkm-name: my-macro\n// @uhkm-version: 1.0.0\n\ntapKey enter\n")
	issues := lint.Check("bom.uhkm", content, config.Default())

	if !hasCode(issues, "UHKM400") {
		t.Errorf("expected UHKM400, got %v", issues)
	}
	// A byte order mark must not cascade: the pragmas below it are plainly
	// inside the preamble, so reporting them as missing or stray is wrong.
	for _, iss := range issues {
		if iss.Code == "UHKM300" || iss.Code == "UHKM304" {
			t.Errorf("byte order mark cascaded into a preamble issue: %s", iss)
		}
	}
}

func TestCheckUHKM400_CleanFile(t *testing.T) {
	content := []byte("// clean.uhkm\n// @uhkm-name: my-macro\n// @uhkm-version: 1.0.0\n\ntapKey enter\n")
	if issues := lint.Check("clean.uhkm", content, config.Default()); len(issues) != 0 {
		t.Errorf("got %d issues, want none: %v", len(issues), issues)
	}
}

func TestFixStripsBOM(t *testing.T) {
	content := []byte("\ufeff// @uhkm-name: my-macro\n")
	fixed, changed := lint.Fix(content, config.Default())
	if !changed {
		t.Fatal("Fix reported no change for a file with a byte order mark")
	}
	if bytes.HasPrefix(fixed, []byte("\ufeff")) {
		t.Errorf("byte order mark not removed: %q", fixed)
	}
}

// An indented comment at the top of a file is still a comment, so it must not
// end the preamble region and strand the pragmas below it.
func TestCheckIndentedCommentDoesNotEndPreamble(t *testing.T) {
	content := []byte("    // a note\n// @uhkm-name: my-macro\n// @uhkm-version: 1.0.0\n\ntapKey enter\n")
	for _, iss := range lint.Check("my-macro.uhkm", content, config.Default()) {
		if iss.Code == "UHKM300" || iss.Code == "UHKM304" {
			t.Errorf("indented comment ended the preamble: %s", iss)
		}
	}
}

// --- UHKM201 ---

func TestCheckUHKM201(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		content string
		want    bool // whether UHKM201 is expected
	}{
		{name: "filename comment", path: "on-init.uhkm", content: "// on-init.uhkm\n// @uhkm-name: x\n// @uhkm-version: 1.0.0\n"},
		{name: "comment with trailing prose", path: "on-init.uhkm", content: "// on-init.uhkm, the entry point\n// @uhkm-name: x\n// @uhkm-version: 1.0.0\n"},
		{name: "missing", path: "on-init.uhkm", content: "// @uhkm-name: x\n// @uhkm-version: 1.0.0\n", want: true},
		{name: "names a different file", path: "on-init.uhkm", content: "// other.uhkm\n// @uhkm-name: x\n// @uhkm-version: 1.0.0\n", want: true},
		{name: "pragma does not satisfy it", path: "on-init.uhkm", content: "// @uhkm-name: on-init.uhkm\n// @uhkm-version: 1.0.0\n", want: true},
		{name: "not a comment", path: "on-init.uhkm", content: "tapKey enter\n", want: true},
		{name: "empty file", path: "on-init.uhkm", content: "", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasCode(lint.Check(tt.path, []byte(tt.content), config.Default()), "UHKM201")
			if got != tt.want {
				t.Errorf("UHKM201 reported = %v, want %v", got, tt.want)
			}
		})
	}
}

// The filename comment is matched against the base name, not the whole path.
func TestCheckUHKM201UsesBaseName(t *testing.T) {
	content := []byte("// on-init.uhkm\n// @uhkm-name: x\n// @uhkm-version: 1.0.0\n")
	if hasCode(lint.Check("macros/events/on-init.uhkm", content, config.Default()), "UHKM201") {
		t.Error("unexpected UHKM201 for a nested path whose base name matches")
	}
}

// A byte order mark must not hide the filename comment either.
func TestCheckUHKM201IgnoresBOM(t *testing.T) {
	content := []byte("\ufeff// on-init.uhkm\n// @uhkm-name: x\n// @uhkm-version: 1.0.0\n")
	if hasCode(lint.Check("on-init.uhkm", content, config.Default()), "UHKM201") {
		t.Error("byte order mark hid the filename comment from UHKM201")
	}
}
