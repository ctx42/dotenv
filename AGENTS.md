This file provides guidance for AI agents (and human contributors) when working
with code in this repository.

## Commands

```bash
# Run all tests (canonical flags)
go test -count=1 -race ./...

# Run a single test
go test -count=1 -race -run Test_Parse ./pkg/dotenv/

# Run a specific subtest
go test -count=1 -race -run 'Test_parseLine_tabular/key_value' ./pkg/dotenv/

# Vet
go vet ./...
```

## Module

`github.com/ctx42/dotenv` — a single-package library with one public-facing
package at `pkg/dotenv`. No `cmd/`, no sub-packages.

## Architecture

The library exposes three public symbols:

- `Parse(m map[string]string, r io.Reader) error` — reads dotenv content
  from a reader into the caller-supplied map. Keys from the reader overwrite
  existing map entries, which is the designed mechanism for pre-seeding
  environment variables (see `Split`).
- `Split(env []string) map[string]string` — converts `os.Environ()` output
  into a map suitable for passing to `Parse`.
- `Override(m map[string]string, env []string) []string` — replaces the value
  of every key in `m` also set in `env` and returns those keys sorted; keys
  absent from `m` are ignored, so the parsed file defines what may be
  overridden.

Internal parsing pipeline:
```
Parse → strip UTF-8 BOM (first line only)
      → isIgnoredLine (skip blank/comment lines)
      → parseLine
          ├─ stripComment (`#` after whitespace, outside a leading quote)
          ├─ separator: first `=` or `:`
          ├─ extractKey (strips leading `export`); empty or whitespace key
          │  → ErrInvLine
          └─ parseValue
               ├─ single quotes  → raw, no escapes, no expansion
               ├─ double quotes  → decodeDoubleQuoted: escapes
               │                   (\n \r \t \f \b \uXXXX) + expansion
               │                   in one pass
               └─ unquoted       → expandVariables (only `\$` escape)
```

Both expansion paths resolve `$NAME` and `${NAME}` with `varRef`; names are
uppercase letters, digits, and underscores (`nameLen`), and `$(...)` is kept
verbatim.

Variable expansion operates against the map `m` passed to `Parse`. Variables
defined earlier in the same file are already in `m` by the time later lines
are processed — this is how intra-file expansion works.

## Test Conventions

- Table-driven tests use a `testN string` field as the subtest name.
- Test structure follows `// --- Given ---` / `// --- When ---` / `// --- Then ---`.
- Assertions use `github.com/ctx42/testing/pkg/assert`.
- Internal (unexported) functions are tested in `package dotenv` (white-box).
- Example functions live in `package dotenv_test` in `examples_test.go`.
- Subtest names use only filename-safe characters (no `/`, `:`, `?`, etc.).
