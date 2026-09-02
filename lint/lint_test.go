package lint_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hastefuI/uhkm/config"
	"github.com/hastefuI/uhkm/lint"
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
