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

// ErrInvLine is returned when a line cannot be parsed as a key=value pair.
var ErrInvLine = errors.New("invalid config line")

var (
	escapeRX        = regexp.MustCompile(`\\.`)
	unescapeCharsRX = regexp.MustCompile(`\\([^$])`)
	// Uppercase-only by convention: .env files use SCREAMING_SNAKE_CASE keys.
	expandVarRX = regexp.MustCompile(`(\\)?(\$)(\()?{?([A-Z0-9_]+)?}?`)
)

// Parse reads dotenv format configuration from r into m. Keys present in m
// before the call are overwritten by keys read from r, which is the intended
// mechanism for pre-seeding the map with environment variables via [Split].
func Parse(m map[string]string, r io.Reader) error {
	scn := bufio.NewScanner(r)
	n := 0
	for scn.Scan() {
		n++
		lin := scn.Text()
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

// parseLine parses configuration line, returns key, value pair.
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
	val := parseValue(lin[sepIdx+1:], m)
	return key, val, nil
}

// parseValue parses value part of the key-value configuration line.
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
		val = escapeRX.ReplaceAllStringFunc(val, func(match string) string {
			c := strings.TrimPrefix(match, `\`)
			switch c {
			case "n":
				return "\n"
			case "r":
				return "\r"
			default:
				return match
			}
		})
		// Unescape characters.
		val = unescapeCharsRX.ReplaceAllString(val, "$1")
	}

	if !isSingle && strings.Contains(val, "$") {
		val = expandVariables(val, m)
	}

	return val
}

// expandVariables expands variables. In the below example, "${OPTION_A}" will
// be expanded to "1".
//
//	OPTION_A=1
//	OPTION_B=${OPTION_A}
func expandVariables(src string, m map[string]string) string {
	repl := func(s string) string {
		sub := expandVarRX.FindStringSubmatch(s)
		if sub == nil {
			return s
		}
		if sub[1] == "\\" || sub[2] == "(" {
			return sub[0][1:]
		} else if sub[4] != "" {
			return m[sub[4]]
		}
		return s
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

// stripComment removes an inline comment from a line, preserving
// '#' characters inside quoted strings.
func stripComment(lin string) string {
	inSingle := false
	inDouble := false
	for i := 0; i < len(lin); i++ {
		switch lin[i] {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case '#':
			if !inSingle && !inDouble {
				return strings.TrimRight(lin[:i], " \t")
			}
		}
	}
	return lin
}
