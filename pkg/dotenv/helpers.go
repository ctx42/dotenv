// SPDX-FileCopyrightText: (c) 2025 Rafal Zajac <rzajac@gmail.com>
// SPDX-License-Identifier: MIT

package dotenv

import (
	"strings"
)

// Split parses [os.Environ] results and returns it as a key value map.
func Split(env []string) map[string]string {
	m := make(map[string]string, 10)
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
