// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dotenv

import (
	"strings"
	"testing"
)

// Realistic .env file for full-parse benchmarks — covers the common
// value types: bare, single-quoted, double-quoted, commented,
// exported, and variable-expanded.
const benchEnvFile = `# Application settings
APP_NAME=myapp
APP_ENV=production
APP_DEBUG=false

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=mydb
DB_USER=admin
DB_PASS='s3cr3t!pass'

# Redis
REDIS_URL="redis://localhost:6379/0"
CACHE_TTL=3600

# API (variable expansion)
BASE_URL=https://api.example.com
API_URL="${BASE_URL}/v1"

# Logging
LOG_LEVEL=info # can be debug, info, warn, error
LOG_FILE=/var/log/myapp.log

# Export-style
export EXPORTED_VAR=exported_value
export ANOTHER_VAR='another value'
`

func BenchmarkParse(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = Parse(map[string]string{}, strings.NewReader(benchEnvFile))
	}
}

func BenchmarkParseLine(b *testing.B) {
	tt := []struct {
		testN string
		line  string
	}{
		{"bare", "KEY=value"},
		{"single_quoted", "KEY='single quoted value'"},
		{"double_quoted", `KEY="double quoted value"`},
		{"with_comment", "KEY=value # inline comment"},
		{"with_expansion", "KEY=${OTHER}"},
		{"with_export", "export KEY=value"},
	}

	m := map[string]string{"OTHER": "expanded"}
	for _, tc := range tt {
		b.Run(tc.testN, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_, _, _ = parseLine(1, tc.line, m)
			}
		})
	}
}

func BenchmarkExpandVariables(b *testing.B) {
	tt := []struct {
		testN string
		input string
		env   map[string]string
	}{
		{
			"no_vars",
			"plain value with no placeholders",
			map[string]string{},
		},
		{
			"single_var",
			"${GREETING} world",
			map[string]string{"GREETING": "hello"},
		},
		{
			"multi_var",
			"${A} and ${B} and ${C}",
			map[string]string{"A": "one", "B": "two", "C": "three"},
		},
	}

	for _, tc := range tt {
		b.Run(tc.testN, func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_ = expandVariables(tc.input, tc.env)
			}
		})
	}
}
