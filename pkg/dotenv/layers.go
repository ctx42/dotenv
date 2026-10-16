// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dotenv

import (
	"io"
	"slices"
)

// Layers builds a configuration from layered dotenv sources: a base
// configuration read first, override configurations read after it, and the
// environment variables given to [Layers.Expand]. A later layer replaces the
// value of a key an earlier layer declares before any variable reference is
// expanded, so a value referring to that key follows the override. The zero
// value is an empty configuration ready to use.
type Layers struct {
	entries []entry          // Assignments in declaration order.
	index   map[string][]int // Indexes in entries of each key's assignments.
}

// entry is a key assignment with its value as written.
type entry struct {
	key string // Key name.
	raw string // Value as written, not yet expanded.
}

// Read reads a layer of dotenv format configuration from r. An assignment to
// a key an earlier layer declares replaces that key's value where the earlier
// layer declared it; an assignment to a key no earlier layer declares is added
// after the keys already declared. A line that is not a key-value pair returns
// an error wrapping [ErrInvLine] with the line number; the assignments read
// before that line are kept.
func (lay *Layers) Read(r io.Reader) error {
	if lay.index == nil {
		lay.index = make(map[string][]int)
	}
	// Entries before prev come from earlier layers.
	prev := len(lay.entries)
	return scanLines(r, func(n int, lin string) error {
		key, raw, err := splitLine(n, lin)
		if err != nil {
			return err
		}
		idx := lay.index[key]
		if len(idx) > 0 && idx[0] < prev {
			for _, i := range idx {
				lay.entries[i].raw = raw
			}
			return nil
		}
		lay.index[key] = append(idx, len(lay.entries))
		lay.entries = append(lay.entries, entry{key: key, raw: raw})
		return nil
	})
}

// Expand returns the configuration with its variable references expanded.
// Values are expanded in declaration order, a reference resolving to the value
// of the key declared before it, as [Parse] does. An environment variable in
// env, given in [os.Environ] form, replaces the value of a declared key before
// expansion, an empty value included; the variable's value is used as is,
// never expanded. A variable whose key no layer declares is ignored. Expand
// also returns the keys env overrode in sorted order, or nil when none.
func (lay *Layers) Expand(env []string) (map[string]string, []string) {
	vars := Split(env)
	m := make(map[string]string, len(lay.index))
	var keys []string
	for _, ent := range lay.entries {
		val, ok := vars[ent.key]
		if !ok {
			m[ent.key] = parseValue(ent.raw, m)
			continue
		}
		m[ent.key] = val
		if !slices.Contains(keys, ent.key) {
			keys = append(keys, ent.key)
		}
	}
	slices.Sort(keys)
	return m, keys
}
