This file provides guidance for AI agents (and human contributors) when working
with code in this repository.

## Commands

```bash
# Run all tests (canonical flags)
go test -count=1 -race ./...

# Run a single test
go test -count=1 -race -run Test_Parse ./pkg/dotenv/

# Run a specific subtest
go test -count=1 -race -run 'Test_parseLine/key_value' ./pkg/dotenv/

# Vet
go vet ./...
```

## Module

`github.com/ctx42/dotenv` — a single-package library with one public-facing
package at `pkg/dotenv`. No `cmd/`, no sub-packages.

## Architecture

The library exposes two public symbols:

- `Parse(m map[string]string, r io.Reader) error` — reads dotenv content
  from a reader into the caller-supplied map. Keys from the reader overwrite
  existing map entries, which is the designed mechanism for pre-seeding
  environment variables (see `Split`).
- `Split(env []string) map[string]string` — converts `os.Environ()` output
  into a map suitable for passing to `Parse`.

Internal parsing pipeline:
```
Parse → isIgnoredLine (skip blank/comment lines)
      → parseLine
          ├─ hash-stripping (respects open quotes)
          ├─ key extraction via exportRX (strips leading `export`)
          └─ parseValue
               ├─ single quotes  → raw, no expansion
               ├─ double quotes  → escape sequences (\n \r) + variable expansion
               └─ unquoted       → variable expansion only
                    └─ expandVariables (expandVarRX: $VAR and ${VAR})
```

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
