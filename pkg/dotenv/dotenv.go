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
	escapeRX        = regexp.MustCompile(`\\.`)
	unescapeCharsRX = regexp.MustCompile(`\\([^$])`)
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
		// Expand newlines.
		expand := func(match string) string {
			c := strings.TrimPrefix(match, `\`)
			switch c {
			case "n":
				return "\n"
			case "r":
				return "\r"
			default:
				return match
			}
		}
		val = escapeRX.ReplaceAllStringFunc(val, expand)
		// Unescape characters.
		val = unescapeCharsRX.ReplaceAllString(val, "$1")
	}

	if !isSingle && strings.Contains(val, "$") {
		val = expandVariables(val, m)
	}

	return val
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

// stripComment removes an inline comment from a line, preserving '#'
// characters inside a quoted value. A value is quoted only when a quote opens
// it right after the separator; the quoted span ends at the matching quote,
// skipping backslash-escaped characters in double quotes. A line with an
// unclosed quoted value is returned unchanged.
func stripComment(lin string) string {
	from := 0
	sep := strings.IndexAny(lin, "=:#")
	if sep != -1 && lin[sep] != '#' {
		from = sep + 1
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
		}
	}
	if idx := strings.IndexByte(lin[from:], '#'); idx != -1 {
		return strings.TrimRight(lin[:from+idx], " \t")
	}
	return lin
}
