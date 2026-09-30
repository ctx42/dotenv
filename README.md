# dotenv: Parse .env files for Go

[![Go Report Card](https://goreportcard.com/badge/github.com/ctx42/dotenv)](https://goreportcard.com/report/github.com/ctx42/dotenv)
[![GoDoc](https://img.shields.io/badge/api-Godoc-blue.svg)](https://pkg.go.dev/github.com/ctx42/dotenv)
![Tests](https://github.com/ctx42/dotenv/actions/workflows/go.yml/badge.svg?branch=master)

<!-- TOC -->
* [dotenv: Parse .env files for Go](#dotenv-parse-env-files-for-go)
  * [Installation](#installation)
  * [Usage](#usage)
    * [Basic](#basic)
    * [Comments](#comments)
    * [Variable Expansion](#variable-expansion)
    * [Split](#split)
    * [Preventing Expansion](#preventing-expansion)
    * [Escape Characters](#escape-characters)
  * [Disclaimer](#disclaimer)
<!-- TOC -->

Dotenv files, typically named .env, store configuration settings as key-value
pairs. This format originates from shell scripts used to set environment
variables.

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

The package has two functions: `Parse` reads dotenv-formatted content from a
reader and stores it in a key-value map, and `Split` converts `os.Environ()`
output into a map for pre-seeding `Parse`.

```go
func Parse(m map[string]string, r io.Reader) error
func Split(env []string) map[string]string
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

```
export KEY_ONE=value_one
export KEY_TWO=value_two
```

For cross-system compatibility, key names should use letters, numbers, and
underscores only and must not start with a number. This follows the regular
expression:

```regexp
[a-zA-Z_]+[a-zA-Z0-9_]*
```

```
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

```
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

Certain backslash sequences are replaced with special characters during parsing,
similar to shell script behavior. Files are read as UTF-8, converting specific
byte pairs into single characters:

- `\n` - becomes a newline (line feed)
- `\r` - becomes a carriage return
- `\t` - becomes a tab
- `\f` - becomes a form feed
- `\b` - becomes a backspace
- `\"` - becomes a double quote
- `\'` - becomes a single quote
- `\\` - becomes a backslash
- `\uABCD` inserts a Unicode character (four hex digits); a UTF-16 surrogate
  pair such as `\uD83D\uDE00` inserts one character

If a backslash precedes any other character, the backslash is removed, and the
character is retained as is.

## Disclaimer

The parsing internals — comment stripping, quote handling, and variable
expansion — are derived from [godotenv](https://github.com/joho/godotenv).
The public API is an independent design that differs in several key ways:

- **Single entry point.** godotenv exposes many functions (`Load`, `Overload`,
  `Read`, `Write`, `Marshal`, `Unmarshal`, `Env`, `Exec`). This library
  provides one: `Parse`.
- **`io.Reader` interface.** godotenv accepts file paths. `Parse` accepts any
  `io.Reader`, keeping I/O concerns with the caller and making the parser
  straightforward to test.
- **Caller-owned map.** The caller passes in the destination map. This makes
  the output container explicit, avoids hidden allocations, and enables the
  pre-seeding pattern below.
- **Explicit environment variable expansion.** godotenv expands OS environment
  variables automatically. Here, expansion is opt-in: pre-seed the map with
  `Split(os.Environ())` before calling `Parse`.
- **No OS side effects.** godotenv's `Load` calls `os.Setenv`. `Parse` never
  touches the OS environment.
