// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dotenv

import (
	"slices"
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

// Override replaces the value of every key in m that is also set in env, given
// in [os.Environ] form. A variable set to an empty value overrides too, and a
// variable whose key is not in m is ignored, so the keys of m define what may
// be overridden. Returns the overridden keys in sorted order, or nil when none.
func Override(m map[string]string, env []string) []string {
	var keys []string
	vars := Split(env)
	for key := range m {
		if val, ok := vars[key]; ok {
			m[key] = val
			keys = append(keys, key)
		}
	}
	slices.Sort(keys)
	return keys
}
