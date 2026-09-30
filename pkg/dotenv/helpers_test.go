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
