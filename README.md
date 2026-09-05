# uhkm [![Build](https://github.com/hastefuI/uhkm/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/hastefuI/uhkm/actions/workflows/ci.yml) [![Go](https://img.shields.io/badge/Go-1.27-00ADD8?logo=go&logoColor=white)](https://go.dev) [![Release](https://img.shields.io/github/v/release/hastefuI/uhkm)](https://github.com/hastefuI/uhkm/releases) [![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/hastefuI/uhkm/blob/main/LICENSE) [![Go Reference](https://pkg.go.dev/badge/go.hasteful.org/uhkm.svg)](https://pkg.go.dev/go.hasteful.org/uhkm)

A CLI for linting and formatting [Ultimate Hacking Keyboard](https://ultimatehackingkeyboard.com/) Macro (`.uhkm`) files based on the [UHKM specification](https://github.com/hastefuI/uhkm-spec).

## Usage

```
uhkm check [paths...]         Run lint checks (UHKM100 to UHKM400)
uhkm check --fix [paths...]   Run lint checks and auto-fix where possible
uhkm format [paths...]        Format files in place
uhkm config [paths...]        Print resolved configuration
uhkm version                  Print version information
uhkm help [command]           Help for any command
```

## Configuration

Place a `.uhkm.toml` file at any directory level. `uhkm` searches upward from each file, stopping at a `.git/` boundary.

```toml
[lint.indentation]
style = "spaces"   # "spaces" or "tabs"
width = 4          # indent width when style = "spaces"

[lint.naming]
convention = "kebab"   # "kebab", "snake", or "pascal"
```

CLI flags override file/default config for `check`, `format`, and `config`:

```sh
--indent-style spaces|tabs
--indent-width <n>
--naming-convention kebab|snake|pascal
```

## Lint Rules

| Code     | Level   | Description                                  | Auto-fix |
|----------|---------|----------------------------------------------|----------|
| UHKM100  | error   | Indentation style and width                  | Yes      |
| UHKM200  | error   | File naming convention                       | No       |
| UHKM201  | error   | Missing first-line filename comment          | No       |
| UHKM300  | error   | Missing required pragma                      | No       |
| UHKM301  | error   | Required pragma has an empty value           | No       |
| UHKM302  | error   | Duplicate pragma key                         | No       |
| UHKM303  | warning | Unknown pragma key                           | No       |
| UHKM304  | warning | Pragma below the preamble (ignored by tools) | No       |
| UHKM305  | error   | @uhkm-version is not a valid semver          | No       |
| UHKM400  | error   | UTF-8 byte order mark                        | Yes      |

Every reported issue, warning or error, makes `check` exit `1`.

UHKM201 requires the first line of a file to be a comment naming the file, for
example `// on-init.uhkm`, so the name survives when a macro is read outside its
original file.

UHKM300 to UHKM305 validate the preamble, the leading block of `// @key: value`
pragmas at the top of a file. `@uhkm-name` and `@uhkm-version` are required, and
`@uhkm-version` must be a semver `MAJOR.MINOR.PATCH` value. The
[UHKM specification](https://github.com/hastefuI/uhkm-spec) documents the full
pragma set.

UHKM400 reports a UTF-8 byte order mark, which the specification forbids. The
mark is ignored by every other rule, so a stray BOM is reported once instead of
cascading into spurious missing-pragma errors, and `check --fix` and `format`
both remove it.

## Exit Codes

| Code | Meaning       |
|------|---------------|
| 0    | Success       |
| 1    | Lint issues   |
| 2    | Tool error    |

## Development

**Prerequisites:** Go 1.27+

```sh
# Run tests
go test ./...

# Build the binary
go build -o uhkm .

# Run directly without building
go run . check [paths...]
go run . format [paths...]
```

## License

Licensed under the [MIT License](https://opensource.org/licenses/MIT). See [`LICENSE`](https://github.com/hastefuI/uhkm/blob/main/LICENSE) for details.

Copyright (c) 2026-present hasteful
