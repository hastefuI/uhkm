# uhkm [![Build](https://github.com/hastefuI/uhkm/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/hastefuI/uhkm/actions/workflows/ci.yml) [![Go](https://img.shields.io/badge/Go-1.27-00ADD8?logo=go&logoColor=white)](https://go.dev) [![Release](https://img.shields.io/github/v/release/hastefuI/uhkm)](https://github.com/hastefuI/uhkm/releases) [![License](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/hastefuI/uhkm/blob/main/LICENSE)

A CLI for linting and formatting [Ultimate Hacking Keyboard](https://ultimatehackingkeyboard.com/) Macro (`.uhkm`) files based on the [UHKM specification](https://github.com/hastefuI/uhkm-spec).

## Usage

```
uhkm check [paths...]         Run lint checks (UHKM100, UHKM200)
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

| Code     | Description                         | Auto-fix |
|----------|-------------------------------------|----------|
| UHKM100  | Indentation style and width         | Yes      |
| UHKM200  | File naming convention              | No       |

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
