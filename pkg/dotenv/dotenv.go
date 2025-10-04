package dotenv

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// ErrInvLine represents invalid dotenv configuration line.
var ErrInvLine = errors.New("invalid config line")

var (
	exportRX        = regexp.MustCompile(`^\s*(?:export\s+)?(.*?)\s*$`)
	singleQuotesRX  = regexp.MustCompile(`\A'(.*)'\z`)
	doubleQuotesRX  = regexp.MustCompile(`\A"(.*)"\z`)
	escapeRX        = regexp.MustCompile(`\\.`)
	unescapeCharsRX = regexp.MustCompile(`\\([^$])`)
	expandVarRX     = regexp.MustCompile(`(\\)?(\$)(\()?{?([A-Z0-9_]+)?}?`)
)

// Parse reads dotenv compatible configuration from the reader to the provided
// map.
func Parse(m map[string]string, r io.Reader) error {
	scn := bufio.NewScanner(r)
	n := 1
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

	// Ditch the comments (but keep quoted hashes).
	if strings.Contains(lin, "#") {
		segmentsBetweenHashes := strings.Split(lin, "#")
		quotesAreOpen := false
		var segmentsToKeep []string
		for _, seg := range segmentsBetweenHashes {
			if strings.Count(seg, "\"") == 1 || strings.Count(seg, "'") == 1 {
				if quotesAreOpen {
					quotesAreOpen = false
					segmentsToKeep = append(segmentsToKeep, seg)
				} else {
					quotesAreOpen = true
				}
			}

			if len(segmentsToKeep) == 0 || quotesAreOpen {
				segmentsToKeep = append(segmentsToKeep, seg)
			}
		}

		lin = strings.Join(segmentsToKeep, "#")
	}

	firstEquals := strings.Index(lin, "=")
	firstColon := strings.Index(lin, ":")
	splitString := strings.SplitN(lin, "=", 2)
	if firstColon != -1 && (firstColon < firstEquals || firstEquals == -1) {
		// This is a yaml-style line.
		splitString = strings.SplitN(lin, ":", 2)
	}

	if len(splitString) != 2 {
		return "", "", fmt.Errorf("%w: %d", ErrInvLine, n)
	}

	// Parse the key
	key := exportRX.ReplaceAllString(splitString[0], "$1")

	// Parse the value
	val := parseValue(splitString[1], m)
	return key, val, nil
}

// parseValue parses value part of the key-value configuration line.
func parseValue(val string, m map[string]string) string {
	val = strings.TrimSpace(val)

	// Check if we've got quoted values or possible escapes.
	if len(val) <= 1 {
		return val
	}

	singleQuotes := singleQuotesRX.FindStringSubmatch(val)
	doubleQuotes := doubleQuotesRX.FindStringSubmatch(val)
	if singleQuotes != nil || doubleQuotes != nil {
		val = val[1 : len(val)-1] // Pull the quotes off the edges.
	}

	if doubleQuotes != nil {
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

	if singleQuotes == nil {
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
