[![Go Report Card](https://goreportcard.com/badge/github.com/ctx42/dotenv)](https://goreportcard.com/report/github.com/ctx42/dotenv)
[![GoDoc](https://img.shields.io/badge/api-Godoc-blue.svg)](https://pkg.go.dev/github.com/ctx42/dotenv)
[![Tests](https://github.com/ctx42/dotenv/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/ctx42/dotenv/actions/workflows/go.yml)

# dotenv: Parse .env files for Go

<!-- TOC -->
* [dotenv: Parse .env files for Go](#dotenv-parse-env-files-for-go)
  * [Features](#features)
  * [Prerequisites](#prerequisites)
  * [Installation](#installation)
  * [Usage](#usage)
    * [Basic](#basic)
    * [Comments](#comments)
    * [Variable Expansion](#variable-expansion)
    * [Split](#split)
    * [Override](#override)
    * [Layers](#layers)
    * [Preventing Expansion](#preventing-expansion)
    * [Escape Characters](#escape-characters)
    * [Format Notes](#format-notes)
    * [Errors](#errors)
  * [Disclaimer](#disclaimer)
<!-- TOC -->

`dotenv` is a Go package that parses `.env` files into a map. Dotenv files
store configuration settings as key-value pairs, a format that originates from
shell scripts used to set environment variables. The package suits
applications and tools that read such files without changing the process
environment.

## Features

- Reads from any `io.Reader`: a file, a string, or any other stream.
- Writes into a map the caller owns, so entries can be pre-seeded.
- Pre-seeds from the process environment with `Split(os.Environ())`.
- Lets the environment override parsed keys with `Override`.
- Layers configurations with `Layers`: an override moves the values derived
  from the key it replaces.
- Expands `$NAME` and `${NAME}` references; single quotes keep values literal.
- Never modifies the OS environment.
- Depends on the Go standard library only.

## Prerequisites

Go 1.26 or later.

## Installation

To use `dotenv` in your Go project, install it with:

```bash
go get github.com/ctx42/dotenv
```

Then import the package:

```go
import "github.com/ctx42/dotenv/pkg/dotenv"
```

## Usage

The package has three functions and one type: `Parse` reads
dotenv-formatted content from a reader and stores it in a key-value map,
`Split` converts `os.Environ()` output into a map for pre-seeding `Parse`,
`Override` replaces parsed values with the environment ones, and `Layers`
builds a configuration from a base file, override files, and the environment.

```go
func Parse(m map[string]string, r io.Reader) error
func Split(env []string) map[string]string
func Override(m map[string]string, env []string) []string

func (lay *Layers) Read(r io.Reader) error
func (lay *Layers) Expand(env []string) (map[string]string, []string)
```

### Basic

<!-- gmmce:pkg/dotenv/ExampleParse_basic -->
```go
file := `
HELLO="hello"
WORLD=world
`

m := map[string]string{}
r := strings.NewReader(file)

if err := dotenv.Parse(m, r); err != nil {
	panic(err)
}

fmt.Println(dump.Any(m))
// Output:
// map[string]string{
//   "HELLO": "hello",
//   "WORLD": "world",
// }
```

You can prefix lines with "export" to enable sourcing the file in a shell
environment:

```dotenv
export KEY_ONE=value_one
export KEY_TWO=value_two
```

For cross-system compatibility, key names should use letters, numbers, and
underscores only and must not start with a number. This follows the regular
expression:

```regexp
[a-zA-Z_]+[a-zA-Z0-9_]*
```

```dotenv
FOOBAR   # ok
FOO_BAR  # ok
foobar   # ok
foo_bar  # ok
foo-bar  # avoid
∑KEY     # avoid
123VAR   # avoid
```

`Parse` rejects only an empty key or a key containing whitespace, returning an
error that wraps `ErrInvLine`; other keys are stored as written. Variable
expansion resolves only names made of uppercase letters, digits, and
underscores, so `${foobar}` is never expanded.

Values follow the equals sign and can be enclosed in quotes if needed. Single
quotes prevent variable expansion within the value.

```dotenv
SIMPLE=hello
EXPAND="multiple\nlines text with variable expansion: ${SIMPLE}"
DO_NOT_EXPAND='raw text without variable interpolation'
```

### Comments

Lines starting with `#` are treated as comments. Comments can also appear after
a value if separated by a space but only outside quotes. Within quoted strings,
`#` is treated as a regular character.

<!-- gmmce:pkg/dotenv/ExampleParse_comments -->
```go
file := `
# Comment.
QUOTED="a # b # c #"
AFTER_SPACE=world # Comment.
`

m := map[string]string{}
r := strings.NewReader(file)

if err := dotenv.Parse(m, r); err != nil {
	panic(err)
}

fmt.Println(dump.Any(m))
// Output:
// map[string]string{
//   "AFTER_SPACE": "world",
//   "QUOTED": "a # b # c #",
// }
```

### Variable Expansion

Unquoted or double-quoted values support placeholders like `${VAR_NAME}`, which
are replaced with the values of previously defined variables in the file.
**Environment variables are not expanded.**

- Both `$VAR_NAME` and `${VAR_NAME}` are supported.
- Names consist of uppercase letters, digits, and underscores; other text
  after `$` is kept as written.
- An undefined variable expands to an empty string.
- `\$` produces a literal `$`, and `$(...)` is kept as written because
  command substitution is not supported.

For instance:

<!-- gmmce:pkg/dotenv/ExampleParse_variableExpansion -->
```go
file := `
HELLO=hello
WORLD=world
MESSAGE="${HELLO} ${WORLD}!"
`

m := map[string]string{}
r := strings.NewReader(file)

if err := dotenv.Parse(m, r); err != nil {
	panic(err)
}

fmt.Println(m["MESSAGE"])
// Output:
// hello world!
```

To expand environment variables, provide an initialized map with the keys to
expand.

<!-- gmmce:pkg/dotenv/ExampleParse_variableExpansionFromEnvironment -->
```go
file := `
HELLO=hello
WORLD=world
MESSAGE="${HELLO} ${WORLD}${EXCLAIM}"
`

env := map[string]string{
	"HELLO":   "env-hello",
	"EXCLAIM": "!",
}
r := strings.NewReader(file)

if err := dotenv.Parse(env, r); err != nil {
	panic(err)
}

fmt.Println(dump.Any(env))
// Output:
// map[string]string{
//   "EXCLAIM": "!",
//   "HELLO": "hello",
//   "MESSAGE": "hello world!",
//   "WORLD": "world",
// }
```

Note that keys from the reader overwrite keys in the initialized map.

### Split

`Split` converts the output of `os.Environ()` into a map suitable for
pre-seeding `Parse` with existing environment variables.

<!-- gmmce:pkg/dotenv/ExampleSplit -->
```go
// Convert os.Environ() output into a map suitable for
// pre-seeding Parse with existing environment variables.
env := dotenv.Split([]string{
	"HOME=/home/user",
	"PATH=/usr/local/bin:/usr/bin",
	"USER=alice",
})

fmt.Println(dump.Any(env))
// Output:
// map[string]string{
//   "HOME": "/home/user",
//   "PATH": "/usr/local/bin:/usr/bin",
//   "USER": "alice",
// }
```

### Override

`Override` replaces the value of every key in the map that is also set in the
environment, given in `os.Environ()` form, and returns the overridden keys in
sorted order. A variable set to an empty value overrides too, and a variable
whose key is not in the map is ignored, so the parsed file defines what the
environment may override. Overrides apply after expansion: a value expanded
from an overridden key keeps the file's value. Use [Layers](#layers) to have
it follow the override.

<!-- gmmce:pkg/dotenv/ExampleOverride -->
```go
file := `
HOST=example.com
REPO=$HOST/team
`
cfg := make(map[string]string)
if err := dotenv.Parse(cfg, strings.NewReader(file)); err != nil {
	panic(err)
}

// Let the environment override the keys the file declares. Variables the
// file does not declare, such as HOME, are ignored.
keys := dotenv.Override(cfg, []string{
	"HOME=/home/user",
	"REPO=registry.acme.io/platform",
})

fmt.Println(keys)
fmt.Println(dump.Any(cfg))
// Output:
// [REPO]
// map[string]string{
//   "HOST": "example.com",
//   "REPO": "registry.acme.io/platform",
// }
```

### Layers

`Layers` builds one configuration from several sources. `Read` adds a layer:
the first one is the base, and each later one overrides it. `Expand` takes the
environment, given in `os.Environ()` form, as the last layer and returns the
configuration with its variable references expanded, along with the keys the
environment overrode in sorted order.

A layer replaces the value of a key an earlier layer declares before anything
is expanded, so a value referring to that key follows the override. A later
file may add keys of its own; an environment variable only overrides a key
some file declares, and its value is used as is, never expanded. A key a later
layer sets keeps that value even when the base derives it. Values are expanded
in declaration order, as `Parse` expands them, so a single layer with no
environment gives the map `Parse` gives.

<!-- gmmce:pkg/dotenv/ExampleLayers -->
```go
base := `
ROOT=/ctx42
GOPATH=$ROOT/go
HOST=ghcr.io
REPO=$HOST/ctx42
`
override := `
ROOT=/acme
`
var lay dotenv.Layers
for _, src := range []string{base, override} {
	if err := lay.Read(strings.NewReader(src)); err != nil {
		panic(err)
	}
}

// The overrides replace values before expansion, so GOPATH follows ROOT
// and REPO follows HOST. Variables no layer declares, such as HOME, are
// ignored.
cfg, keys := lay.Expand([]string{
	"HOME=/home/user",
	"HOST=registry.acme.io",
})

fmt.Println(keys)
fmt.Println(dump.Any(cfg))
// Output:
// [HOST]
// map[string]string{
//   "GOPATH": "/acme/go",
//   "HOST": "registry.acme.io",
//   "REPO": "registry.acme.io/ctx42",
//   "ROOT": "/acme",
// }
```

### Preventing Expansion

To preserve placeholders like `${}` in a value, enclose it in single quotes:

<!-- gmmce:pkg/dotenv/ExampleParse_preventVariableExpansion -->
```go
file := `
HELLO=hello
WORLD=world
MESSAGE='${HELLO} ${WORLD}!'
`

m := map[string]string{}
r := strings.NewReader(file)

if err := dotenv.Parse(m, r); err != nil {
	panic(err)
}

fmt.Println(m["MESSAGE"])
// Output:
// ${HELLO} ${WORLD}!
```

### Escape Characters

In double-quoted values, certain backslash sequences are replaced with special
characters, similar to shell script behavior:

- `\n` - becomes a newline (line feed)
- `\r` - becomes a carriage return
- `\t` - becomes a tab
- `\f` - becomes a form feed
- `\b` - becomes a backspace
- `\"` - becomes a double quote
- `\'` - becomes a single quote
- `\\` - becomes a backslash
- `\$` - becomes a literal `$` instead of starting a variable
- `\uABCD` inserts a Unicode character (four hex digits); a UTF-16 surrogate
  pair such as `\uD83D\uDE00` inserts one character

If a backslash precedes any other character, the backslash is removed, and the
character is retained as is. Single-quoted values are taken literally, and
unquoted values keep their backslashes, except that `\$` is a literal `$`.

### Format Notes

- A key and its value may be separated by `=` or `:` (`KEY: value`); the
  first of the two on the line is the separator.
- A UTF-8 byte order mark at the start of the input is ignored.
- Lines are limited to 1 MiB; a longer line fails with an error wrapping
  `bufio.ErrTooLong`.

### Errors

A line that is not a key-value pair makes `Parse` return an error wrapping
`ErrInvLine` that names the line number, for example
`dotenv: invalid config line: 3`. The map keeps the pairs read before that
line. A nil map returns `ErrNilMap`. `Layers.Read` returns the same error for
such a line and keeps the assignments read before it.

## Disclaimer

The parsing internals — comment stripping, quote handling, and variable
expansion — are derived from [godotenv](https://github.com/joho/godotenv).
The public API is an independent design that differs in several key ways:

- **Single entry point.** godotenv exposes many functions (`Load`, `Overload`,
  `Read`, `Write`, `Marshal`, `Unmarshal`, `Env`, `Exec`). This library
  provides one parsing function, `Parse`, plus the `Split` and `Override`
  helpers.
- **`io.Reader` interface.** godotenv accepts file paths. `Parse` accepts any
  `io.Reader`, keeping I/O concerns with the caller and making the parser
  straightforward to test.
- **Caller-owned map.** The caller passes in the destination map. This makes
  the output container explicit and enables pre-seeding it (see
  [Split](#split)).
- **Explicit environment variable expansion.** godotenv expands OS environment
  variables automatically. Here, expansion is opt-in: pre-seed the map with
  `Split(os.Environ())` before calling `Parse`.
- **No OS side effects.** godotenv's `Load` calls `os.Setenv`. `Parse` never
  touches the OS environment.
