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
	"regexp"
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

var (
	// Matches an escaped "\$", a "$(", "${NAME}", or "$NAME". Names are
	// uppercase-only by convention: .env files use SCREAMING_SNAKE_CASE
	// keys.
	expandVarRX = regexp.MustCompile(
		`\\\$|\$\(|\$\{[A-Z0-9_]+\}|\$[A-Z0-9_]+`,
	)
)

// Parse reads dotenv format configuration from r into m. Keys present in m
// before the call are overwritten by keys read from r, which is the intended
// mechanism for pre-seeding the map with environment variables via [Split].
// Returns [ErrNilMap] when m is nil.
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

	if isDouble {
		val = unescape(val)
	}

	if !isSingle && strings.Contains(val, "$") {
		val = expandVariables(val, m)
	}

	return val
}

// unescape decodes backslash escape sequences in a double-quoted value:
// "\n", "\r", "\t", "\f", "\b", and "\uABCD" become the characters they name;
// any other escaped character is kept without the backslash. An escaped "$"
// is kept escaped for expandVariables.
func unescape(val string) string {
	if !strings.Contains(val, `\`) {
		return val
	}
	var buf strings.Builder
	buf.Grow(len(val))
	for i := 0; i < len(val); i++ {
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
		case '$':
			buf.WriteString(`\$`)
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
// from m. In the example below, "${OPTION_A}" expands to "1".
//
//	OPTION_A=1
//	OPTION_B=${OPTION_A}
func expandVariables(src string, m map[string]string) string {
	repl := func(s string) string {
		switch s {
		case `\$`:
			return "$"
		case "$(":
			// Keep unsupported command substitution verbatim.
			return s
		}
		return m[strings.Trim(s[1:], "{}")]
	}
	return expandVarRX.ReplaceAllStringFunc(src, repl)
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
