// SPDX-FileCopyrightText: (c) 2013 John Barton
// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

// Package dotenv parses .env format configuration files.
package dotenv

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// maxLineSize is the longest line, in bytes, [Parse] accepts.
const maxLineSize = 1 << 20

// Sentinel errors.
var (
	// ErrInvLine is returned when a line cannot be parsed as a key=value
	// pair.
	ErrInvLine = errors.New("dotenv: invalid config line")

	// ErrNilMap is returned by [Parse] when given a nil map.
	ErrNilMap = errors.New("dotenv: nil map")
)

// Parse reads dotenv format configuration from r into m. Keys present in m
// before the call are overwritten by keys read from r, which is the intended
// mechanism for pre-seeding the map with environment variables via [Split].
// Returns [ErrNilMap] when m is nil. A line that is not a key-value pair
// returns an error wrapping [ErrInvLine] with the line number; m then keeps
// the pairs read before that line.
func Parse(m map[string]string, r io.Reader) error {
	if m == nil {
		return ErrNilMap
	}
	scn := bufio.NewScanner(r)
	scn.Buffer(nil, maxLineSize)
	n := 0
	for scn.Scan() {
		n++
		lin := scn.Text()
		if n == 1 {
			// Drop the UTF-8 byte order mark some editors write.
			lin = strings.TrimPrefix(lin, "\ufeff")
		}
		if isIgnoredLine(lin) {
			continue
		}
		key, value, err := parseLine(n, lin, m)
		if err != nil {
			return err
		}
		m[key] = value
	}
	if err := scn.Err(); err != nil {
		return fmt.Errorf("dotenv: %w", err)
	}
	return nil
}

// parseLine parses a configuration line and returns its key and value. Errors
// report n as the line number.
func parseLine(n int, lin string, m map[string]string) (string, string, error) {
	if lin == "" {
		return "", "", fmt.Errorf("%w: %d", ErrInvLine, n)
	}

	lin = stripComment(lin)

	firstEquals := strings.Index(lin, "=")
	firstColon := strings.Index(lin, ":")

	var sepIdx int
	if firstColon != -1 && (firstColon < firstEquals || firstEquals == -1) {
		sepIdx = firstColon
	} else if firstEquals != -1 {
		sepIdx = firstEquals
	} else {
		return "", "", fmt.Errorf("%w: %d", ErrInvLine, n)
	}

	key := extractKey(lin[:sepIdx])
	if key == "" || strings.ContainsAny(key, " \t") {
		return "", "", fmt.Errorf("%w: %d", ErrInvLine, n)
	}
	val := parseValue(lin[sepIdx+1:], m)
	return key, val, nil
}

// parseValue parses the value part of a key-value configuration line.
func parseValue(val string, m map[string]string) string {
	val = strings.TrimSpace(val)

	// Check if we've got quoted values or possible escapes.
	if len(val) <= 1 {
		return val
	}

	first, last := val[0], val[len(val)-1]
	isSingle := first == '\'' && last == '\''
	isDouble := first == '"' && last == '"'
	if isSingle || isDouble {
		val = val[1 : len(val)-1]
	}

	switch {
	case isSingle:
		return val
	case isDouble:
		return decodeDoubleQuoted(val, m)
	case strings.Contains(val, "$"):
		return expandVariables(val, m)
	}
	return val
}

// decodeDoubleQuoted decodes a double-quoted value in one pass: it expands
// variable references with values from m and decodes backslash escapes.
// "\n", "\r", "\t", "\f", "\b", and "\uABCD" become the characters they
// name, "\$" is a literal "$", and any other escaped character is kept
// without the backslash, so "\\$NAME" is a backslash followed by the value
// of NAME.
func decodeDoubleQuoted(val string, m map[string]string) string {
	if !strings.ContainsAny(val, `\$`) {
		return val
	}
	var buf strings.Builder
	buf.Grow(len(val))
	for i := 0; i < len(val); i++ {
		if val[i] == '$' {
			name, n := varRef(val[i:])
			if n == 0 {
				buf.WriteByte('$')
				continue
			}
			buf.WriteString(m[name])
			i += n - 1
			continue
		}
		if val[i] != '\\' || i+1 == len(val) {
			buf.WriteByte(val[i])
			continue
		}
		i++
		switch c := val[i]; c {
		case 'n':
			buf.WriteByte('\n')
		case 'r':
			buf.WriteByte('\r')
		case 't':
			buf.WriteByte('\t')
		case 'f':
			buf.WriteByte('\f')
		case 'b':
			buf.WriteByte('\b')
		case 'u':
			if i+5 <= len(val) {
				hex := val[i+1 : i+5]
				code, err := strconv.ParseUint(hex, 16, 16)
				if err == nil {
					buf.WriteRune(rune(code))
					i += 4
					continue
				}
			}
			buf.WriteByte(c)
		default:
			buf.WriteByte(c)
		}
	}
	return buf.String()
}

// expandVariables replaces variable references in src with their values
// from m; "\$" is a literal "$". In the example below, "${OPTION_A}" expands
// to "1".
//
//	OPTION_A=1
//	OPTION_B=${OPTION_A}
func expandVariables(src string, m map[string]string) string {
	var buf strings.Builder
	buf.Grow(len(src))
	for i := 0; i < len(src); i++ {
		switch {
		case src[i] == '\\' && i+1 < len(src) && src[i+1] == '$':
			buf.WriteByte('$')
			i++
		case src[i] == '$':
			name, n := varRef(src[i:])
			if n == 0 {
				buf.WriteByte('$')
				continue
			}
			buf.WriteString(m[name])
			i += n - 1
		default:
			buf.WriteByte(src[i])
		}
	}
	return buf.String()
}

// varRef returns the name of the "$NAME" or "${NAME}" reference at the start
// of s and the reference length in bytes; the length is 0 when s does not
// start with a reference. A "$(" is never a reference, so command
// substitution stays verbatim.
func varRef(s string) (string, int) {
	if len(s) > 1 && s[1] == '{' {
		n := nameLen(s[2:])
		if n > 0 && 2+n < len(s) && s[2+n] == '}' {
			return s[2 : 2+n], n + 3
		}
		return "", 0
	}
	n := nameLen(s[1:])
	if n == 0 {
		return "", 0
	}
	return s[1 : 1+n], n + 1
}

// nameLen returns the length of the variable name at the start of s. Names
// are uppercase-only by convention: .env files use SCREAMING_SNAKE_CASE keys.
func nameLen(s string) int {
	n := 0
	for n < len(s) {
		c := s[n]
		if (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '_' {
			break
		}
		n++
	}
	return n
}

// isIgnoredLine returns true for empty or comment lines.
func isIgnoredLine(lin string) bool {
	lin = strings.TrimSpace(lin)
	return lin == "" || strings.HasPrefix(lin, "#")
}

// extractKey strips leading/trailing whitespace and the optional
// "export" keyword prefix from a key string.
func extractKey(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 6 && s[:6] == "export" {
		c := s[6]
		if c == ' ' || c == '\t' {
			return strings.TrimLeft(s[7:], " \t")
		}
	}
	return s
}

// stripComment removes an inline comment from a line. In a value, a '#'
// starts a comment only at the value start or after a space or tab, so
// "p#ss" is kept. A '#' inside a quoted value is kept too: a value is quoted
// only when a quote opens it right after the separator, and the quoted span
// ends at the matching quote, skipping backslash-escaped characters in double
// quotes. A line with an unclosed quoted value is returned unchanged.
func stripComment(lin string) string {
	sep := strings.IndexAny(lin, "=:#")
	if sep == -1 {
		return lin
	}
	if lin[sep] == '#' {
		return strings.TrimRight(lin[:sep], " \t")
	}
	from := sep + 1
	for from < len(lin) && (lin[from] == ' ' || lin[from] == '\t') {
		from++
	}
	if from < len(lin) && (lin[from] == '"' || lin[from] == '\'') {
		quote := lin[from]
		from++
		for from < len(lin) && lin[from] != quote {
			if quote == '"' && lin[from] == '\\' {
				from++
			}
			from++
		}
		if from >= len(lin) {
			return lin
		}
		from++ // Past the closing quote.
	}
	for i := from; i < len(lin); i++ {
		if lin[i] != '#' {
			continue
		}
		if i == from || lin[i-1] == ' ' || lin[i-1] == '\t' {
			return strings.TrimRight(lin[:i], " \t")
		}
	}
	return lin
}
