// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dotenv

import (
	"strings"
)

// Split converts [os.Environ] output into a key-value map.
func Split(env []string) map[string]string {
	m := make(map[string]string, len(env))
	for _, s := range env {
		key, val, ok := strings.Cut(s, "=")
		if !ok || key == "" {
			continue
		}
		m[key] = val
	}
	return m
}
