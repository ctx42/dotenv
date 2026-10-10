// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dotenv

import (
	"testing"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_Split_tabular(t *testing.T) {
	tt := []struct {
		testN string

		env  []string
		want map[string]string
	}{
		{"empty slice", []string{}, map[string]string{}},
		{"slice with empty string", []string{""}, map[string]string{}},
		{"without equal sign", []string{"A"}, map[string]string{}},
		{"empty key", []string{`=C:=C:\`}, map[string]string{}},
		{
			"simple key value",
			[]string{"A=B"},
			map[string]string{"A": "B"},
		},
		{
			"multiple equal",
			[]string{"A=B=C"},
			map[string]string{"A": "B=C"},
		},
		{
			"key without value",
			[]string{"A="},
			map[string]string{"A": ""},
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := Split(tc.env)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_Override(t *testing.T) {
	t.Run("overrides keys set in env", func(t *testing.T) {
		// --- Given ---
		m := map[string]string{"A": "a", "B": "b", "C": "c"}
		env := []string{"C=env-c", "A=env-a"}

		// --- When ---
		have := Override(m, env)

		// --- Then ---
		assert.Equal(t, []string{"A", "C"}, have)
		want := map[string]string{"A": "env-a", "B": "b", "C": "env-c"}
		assert.Equal(t, want, m)
	})

	t.Run("empty env value overrides", func(t *testing.T) {
		// --- Given ---
		m := map[string]string{"A": "a"}
		env := []string{"A="}

		// --- When ---
		have := Override(m, env)

		// --- Then ---
		assert.Equal(t, []string{"A"}, have)
		assert.Equal(t, map[string]string{"A": ""}, m)
	})

	t.Run("env key not in map is ignored", func(t *testing.T) {
		// --- Given ---
		m := map[string]string{"A": "a"}
		env := []string{"B=env-b"}

		// --- When ---
		have := Override(m, env)

		// --- Then ---
		assert.Nil(t, have)
		assert.Equal(t, map[string]string{"A": "a"}, m)
	})

	t.Run("nil map", func(t *testing.T) {
		// --- Given ---
		env := []string{"A=env-a"}

		// --- When ---
		have := Override(nil, env)

		// --- Then ---
		assert.Nil(t, have)
	})
}
