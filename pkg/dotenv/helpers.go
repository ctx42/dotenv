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
		if s == "" {
			continue
		}
		parts := strings.SplitN(s, "=", 2)
		if len(parts) == 2 {
			if parts[0] == "" {
				continue
			}
			m[parts[0]] = parts[1]
		}
	}
	return m
}
