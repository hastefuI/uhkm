package lint_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hastefuI/uhkm/config"
	"github.com/hastefuI/uhkm/lint"
)

const validPreamble = "// @uhkm-name: my-macro\n// @uhkm-version: 1.0.0\n"

// --- UHKM300-UHKM304 ---

func TestCheckPragmas(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		wantCode string // "" means no pragma issue is expected
		wantLine int
	}{
		{
			name:    "valid preamble",
			content: validPreamble + "\ntapKey enter\n",
		},
		{
			name:    "all optional pragmas",
			content: "// @uhkm-name: my-macro\n// @uhkm-version: 1.0.0\n// @uhkm-spec: 1\n// @uhkm-firmware: >=8.0.0\n// @uhkm-author: A Name <a@b.c>\n// @uhkm-license: MIT\n// @uhkm-os: linux\n// @uhkm-description: does a thing\n\ntapKey enter\n",
		},
		{
			name:     "missing both required pragmas",
			content:  "tapKey enter\n",
			wantCode: "UHKM300",
		},
		{
			name:     "missing version",
			content:  "// @uhkm-name: my-macro\n\ntapKey enter\n",
			wantCode: "UHKM300",
		},
		{
			name:     "empty required value",
			content:  "// @uhkm-name:\n// @uhkm-version: 1.0.0\n\ntapKey enter\n",
			wantCode: "UHKM301",
			wantLine: 1,
		},
		{
			name:     "whitespace-only required value",
			content:  "// @uhkm-name: my-macro\n// @uhkm-version:   \n\ntapKey enter\n",
			wantCode: "UHKM301",
			wantLine: 2,
		},
		{
			name:     "duplicate key",
			content:  validPreamble + "// @uhkm-name: other\n\ntapKey enter\n",
			wantCode: "UHKM302",
			wantLine: 3,
		},
		{
			name:     "unknown key",
			content:  validPreamble + "// @name: my-macro\n\ntapKey enter\n",
			wantCode: "UHKM303",
			wantLine: 3,
		},
		{
			name:     "pragma below the preamble",
			content:  validPreamble + "\ntapKey enter\n// @uhkm-os: linux\n",
			wantCode: "UHKM304",
			wantLine: 5,
		},
	}

	cfg := config.Default()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := lint.Check("my-macro.uhkm", []byte(tt.content), cfg)

			if tt.wantCode == "" {
				for _, iss := range issues {
					if iss.Code >= "UHKM300" {
						t.Errorf("unexpected pragma issue: %s", iss)
					}
				}
				return
			}

			var found bool
			for _, iss := range issues {
				if iss.Code != tt.wantCode {
					continue
				}
				found = true
				if iss.Line != tt.wantLine {
					t.Errorf("%s reported on line %d, want %d", iss.Code, iss.Line, tt.wantLine)
				}
			}
			if !found {
				t.Errorf("expected %s, got %v", tt.wantCode, issues)
			}
		})
	}
}

func TestCheckPragmasMissingReportsEachRequiredKey(t *testing.T) {
	issues := lint.Check("my-macro.uhkm", []byte("tapKey enter\n"), config.Default())

	var missing []string
	for _, iss := range issues {
		if iss.Code == "UHKM300" {
			if iss.Line != 0 {
				t.Errorf("UHKM300 should be file-level, got line %d", iss.Line)
			}
			missing = append(missing, iss.Message)
		}
	}
	if len(missing) != 2 {
		t.Fatalf("got %d UHKM300 issues (%v), want 2", len(missing), missing)
	}
}

// A pragma below the preamble is ignored by tooling, so it must not satisfy a
// required key: the file gets both the warning and the missing-pragma error.
func TestCheckPragmasBelowPreambleDoesNotSatisfyRequired(t *testing.T) {
	content := "// @uhkm-name: my-macro\n\ntapKey enter\n// @uhkm-version: 1.0.0\n"
	issues := lint.Check("my-macro.uhkm", []byte(content), config.Default())

	if !hasCode(issues, "UHKM304") {
		t.Error("expected UHKM304 for the pragma below the preamble")
	}
	if !hasCode(issues, "UHKM300") {
		t.Error("expected UHKM300: a pragma below the preamble does not satisfy the requirement")
	}
}

// Duplicate detection covers only the preamble, where tooling reads pragmas.
func TestCheckPragmasDuplicateBelowPreambleIsNotADuplicate(t *testing.T) {
	content := validPreamble + "\ntapKey enter\n// @uhkm-name: other\n"
	issues := lint.Check("my-macro.uhkm", []byte(content), config.Default())

	if hasCode(issues, "UHKM302") {
		t.Error("unexpected UHKM302 for a pragma outside the preamble")
	}
	if !hasCode(issues, "UHKM304") {
		t.Error("expected UHKM304 for the pragma below the preamble")
	}
}

func TestCheckPragmasSeverities(t *testing.T) {
	want := map[string]lint.Severity{
		"UHKM300": lint.SeverityError,
		"UHKM301": lint.SeverityError,
		"UHKM302": lint.SeverityError,
		"UHKM303": lint.SeverityWarning,
		"UHKM304": lint.SeverityWarning,
	}

	// One file that trips every pragma rule: a duplicate and an unknown key in
	// the preamble, an empty required value, a stray pragma in the body, and no
	// @uhkm-version anywhere in the preamble.
	content := "// @uhkm-name:\n" +
		"// @uhkm-name: my-macro\n" +
		"// @nope: x\n" +
		"\n" +
		"tapKey enter\n" +
		"// @uhkm-os: linux\n"

	got := make(map[string]lint.Severity)
	for _, iss := range lint.Check("my-macro.uhkm", []byte(content), config.Default()) {
		got[iss.Code] = iss.Severity
	}

	for code, severity := range want {
		actual, ok := got[code]
		if !ok {
			t.Errorf("%s not reported", code)
			continue
		}
		if actual != severity {
			t.Errorf("%s severity = %v, want %v", code, actual, severity)
		}
	}
}

func TestSeverityString(t *testing.T) {
	if got := lint.SeverityError.String(); got != "error" {
		t.Errorf("SeverityError = %q, want %q", got, "error")
	}
	if got := lint.SeverityWarning.String(); got != "warning" {
		t.Errorf("SeverityWarning = %q, want %q", got, "warning")
	}
	var zero lint.Severity
	if zero != lint.SeverityError {
		t.Error("zero value of Severity should be SeverityError")
	}
}

// The *-valid.uhkm fixtures are the reference for a conformant file, so they
// must produce no issues at all, not merely no indentation issues.
func TestCheckValidFixturesAreClean(t *testing.T) {
	tests := []struct {
		file string
		cfg  config.Config
	}{
		{file: "spaces-valid.uhkm", cfg: config.Default()},
		{
			file: "tabs-valid.uhkm",
			cfg: config.Config{
				Lint: config.Lint{
					Indentation: config.Indentation{Style: "tabs"},
					Naming:      config.Naming{Convention: "kebab"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			path := filepath.Join("testdata", tt.file)
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if issues := lint.Check(path, content, tt.cfg); len(issues) != 0 {
				t.Errorf("got %d issues, want none: %v", len(issues), issues)
			}
		})
	}
}

// --- UHKM305 ---

func TestCheckPragmasVersionSemver(t *testing.T) {
	tests := []struct {
		version string
		want    bool // whether UHKM305 is expected
	}{
		{version: "1.0.0"},
		{version: "0.1.0"},
		{version: "10.20.30"},
		{version: "1.0.0-rc.1"},
		{version: "1.0.0+build.5"},
		{version: "1.0.0-alpha+001"},
		{version: "1", want: true},
		{version: "1.0", want: true},
		{version: "v1.0.0", want: true},
		{version: "1.0.0.0", want: true},
		{version: "01.0.0", want: true},
		{version: "not-semver", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			content := "// my-macro.uhkm\n// @uhkm-name: my-macro\n// @uhkm-version: " + tt.version + "\n\ntapKey enter\n"
			got := hasCode(lint.Check("my-macro.uhkm", []byte(content), config.Default()), "UHKM305")
			if got != tt.want {
				t.Errorf("UHKM305 for %q = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

// An empty version is UHKM301's business, so it must be reported once.
func TestCheckPragmasEmptyVersionIsNotASemverIssue(t *testing.T) {
	content := []byte("// my-macro.uhkm\n// @uhkm-name: my-macro\n// @uhkm-version:\n\ntapKey enter\n")
	issues := lint.Check("my-macro.uhkm", content, config.Default())

	if !hasCode(issues, "UHKM301") {
		t.Error("expected UHKM301 for an empty required value")
	}
	if hasCode(issues, "UHKM305") {
		t.Error("an empty version must be reported once, by UHKM301")
	}
}
