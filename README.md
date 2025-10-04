# dotenv: Parse .env files for Go
i
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

## Usage

The library offers a single function, Parse, which reads dotenv-formatted
content from a reader and stores it in a key-value map.

```go
func Parse(m map[string]string, r io.Reader) error
```

### Basic

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
foo-bar  # invalid
∑KEY     # invalid
123VAR   # invalid
```

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

For instance:

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

```go
file := `
HELLO=hello 
WORLD=world
MESSAGE="${HELLO} ${WORLD}${EXCLAIM}"
`

// env := dotenv.Split(os.Environ())
// or
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

### Preventing Expansion

To preserve placeholders like `${}` in a value, enclose it in single quotes:

```
SECURE_PASS='complex$%{pass}word'
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
- `\uABCD` inserts a Unicode character (four hex digits)

If a backslash precedes any other character, the backslash is removed, and the
character is retained as is.

## Disclaimer

This project is a heavily customized version derived from the original godotenv
library found at [godotenv](https://github.com/joho/godotenv).