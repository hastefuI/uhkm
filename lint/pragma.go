package lint

import (
	"fmt"
	"regexp"
	"slices"

	"go.hasteful.org/uhkm/preamble"
)

// requiredPragmas are the pragma keys every published .uhkm file must define.
var requiredPragmas = []string{"uhkm-name", "uhkm-version"}

// knownPragmas are the pragma keys defined by the UHKM specification. Keys
// outside this set are reported as warnings rather than errors so that files
// written against a newer spec still lint.
var knownPragmas = []string{
	"uhkm-author",
	"uhkm-description",
	"uhkm-firmware",
	"uhkm-license",
	"uhkm-name",
	"uhkm-os",
	"uhkm-spec",
	"uhkm-version",
}

// reSemver matches a SemVer 2.0.0 version: MAJOR.MINOR.PATCH with optional
// prerelease and build metadata, as published at https://semver.org. This
// validates a pragma *value*, not pragma syntax, which stays in the preamble
// package.
var reSemver = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)` +
	`(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?` +
	`(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

// --- UHKM300-UHKM305: Preamble and pragmas ---

// checkPragmas validates the preamble of a .uhkm file:
//
//	UHKM300  required pragma is missing            (error)
//	UHKM301  required pragma has an empty value    (error)
//	UHKM302  pragma key is defined more than once  (error)
//	UHKM303  pragma key is not known to the spec   (warning)
//	UHKM304  pragma sits below the preamble        (warning)
//	UHKM305  @uhkm-version is not a semver value   (error)
//
// A pragma below the preamble is ignored by tooling, so it neither satisfies a
// required key nor counts towards duplicate detection.
func checkPragmas(path string, content []byte) []Issue {
	var issues []Issue
	firstLine := make(map[string]int) // pragma key -> line of its first definition

	for _, p := range preamble.Parse(content) {
		if !p.InPreamble {
			issues = append(issues, Issue{
				Code:     "UHKM304",
				Severity: SeverityWarning,
				File:     path,
				Line:     p.Line,
				Message:  fmt.Sprintf("pragma %q appears below the preamble and is ignored", pragmaName(p.Key)),
			})
			continue
		}

		if first, duplicate := firstLine[p.Key]; duplicate {
			issues = append(issues, Issue{
				Code:     "UHKM302",
				Severity: SeverityError,
				File:     path,
				Line:     p.Line,
				Message:  fmt.Sprintf("duplicate pragma %q, first defined on line %d", pragmaName(p.Key), first),
			})
			continue
		}
		firstLine[p.Key] = p.Line

		switch {
		case !slices.Contains(knownPragmas, p.Key):
			issues = append(issues, Issue{
				Code:     "UHKM303",
				Severity: SeverityWarning,
				File:     path,
				Line:     p.Line,
				Message:  fmt.Sprintf("unknown pragma %q", pragmaName(p.Key)),
			})
		case p.Value == "" && slices.Contains(requiredPragmas, p.Key):
			issues = append(issues, Issue{
				Code:     "UHKM301",
				Severity: SeverityError,
				File:     path,
				Line:     p.Line,
				Message:  fmt.Sprintf("required pragma %q has an empty value", pragmaName(p.Key)),
			})
		case p.Key == "uhkm-version" && !reSemver.MatchString(p.Value):
			issues = append(issues, Issue{
				Code:     "UHKM305",
				Severity: SeverityError,
				File:     path,
				Line:     p.Line,
				Message:  fmt.Sprintf("pragma %q value %q is not a semver version (MAJOR.MINOR.PATCH)", pragmaName(p.Key), p.Value),
			})
		}
	}

	for _, key := range requiredPragmas {
		if _, ok := firstLine[key]; !ok {
			issues = append(issues, Issue{
				Code:     "UHKM300",
				Severity: SeverityError,
				File:     path,
				Message:  fmt.Sprintf("missing required pragma %q", pragmaName(key)),
			})
		}
	}

	return issues
}

// pragmaName renders a pragma key the way it is written in a file.
func pragmaName(key string) string { return "@" + key }
