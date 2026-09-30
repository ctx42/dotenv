// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dotenv

import (
	"errors"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_Parse(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		// --- Given ---
		content := "" +
			"ONE=1\n" +
			"TWO='2'\n" +
			"THREE = \"3\""
		rdr := strings.NewReader(content)

		// --- When ---
		have := map[string]string{}
		err := Parse(have, rdr)

		// --- Then ---
		assert.NoError(t, err)
		want := map[string]string{
			"ONE":   "1",
			"TWO":   "2",
			"THREE": "3",
		}
		assert.Equal(t, want, have)
	})

	t.Run("commented lines are skipped", func(t *testing.T) {
		// --- Given ---
		content := "" +
			"\t#ONE=1\n" +
			"TWO='2'\n" +
			"#THREE = \"3\""
		rdr := strings.NewReader(content)

		// --- When ---
		have := map[string]string{}
		err := Parse(have, rdr)

		// --- Then ---
		assert.NoError(t, err)
		want := map[string]string{
			"TWO": "2",
		}
		assert.Equal(t, want, have)
	})

	t.Run("whitespace around equal sign", func(t *testing.T) {
		// --- Given ---
		content := "" +
			"OPTION_A=1\n" +
			"OPTION_B=2\n" +
			"OPTION_C= 3\n" +
			"OPTION_D =4\n" +
			"OPTION_E = 5\n" +
			"OPTION_F =\n" +
			"OPTION_G=\n"

		// --- When ---
		have := map[string]string{}
		err := Parse(have, strings.NewReader(content))

		// --- Then ---
		assert.NoError(t, err)
		want := map[string]string{
			"OPTION_A": "1",
			"OPTION_B": "2",
			"OPTION_C": "3",
			"OPTION_D": "4",
			"OPTION_E": "5",
			"OPTION_F": "",
			"OPTION_G": "",
		}
		assert.Equal(t, want, have)
	})

	t.Run("lines starting with export are ok", func(t *testing.T) {
		// --- Given ---
		content := "" +
			"export OPTION_A=2\n" +
			"export OPTION_B='\\n'"

		// --- When ---
		have := map[string]string{}
		err := Parse(have, strings.NewReader(content))

		// --- Then ---
		assert.NoError(t, err)
		want := map[string]string{
			"OPTION_A": "2",
			"OPTION_B": "\\n",
		}
		assert.Equal(t, want, have)
	})

	t.Run("values may contain equal sign", func(t *testing.T) {
		// --- Given ---
		content := "export OPTION_A='postgres://localhost:5432/database?sslmode=disable'\n"

		// --- When ---
		have := map[string]string{}
		err := Parse(have, strings.NewReader(content))

		// --- Then ---
		assert.NoError(t, err)
		want := map[string]string{
			"OPTION_A": "postgres://localhost:5432/database?sslmode=disable",
		}
		assert.Equal(t, want, have)
	})

	t.Run("quoted values", func(t *testing.T) {
		// --- Given ---
		content := "" +
			"OPTION_A='1'\n" +
			"OPTION_B='2'\n" +
			"OPTION_C=''\n" +
			"OPTION_D='\\n'\n" +
			"OPTION_E=\"1\"\n" +
			"OPTION_F=\"2\"\n" +
			"OPTION_G=\"\"\n" +
			"OPTION_H=\"\\n\"\n" +
			"OPTION_I = \"echo 'asd'\"\n" +
			"OPTION_J = \"echo \"\\r\"\""

		// --- When ---
		have := map[string]string{}
		err := Parse(have, strings.NewReader(content))

		// --- Then ---
		assert.NoError(t, err)
		want := map[string]string{
			"OPTION_A": "1",
			"OPTION_B": "2",
			"OPTION_C": "",
			"OPTION_D": "\\n",
			"OPTION_E": "1",
			"OPTION_F": "2",
			"OPTION_G": "",
			"OPTION_H": "\n",
			"OPTION_I": "echo 'asd'",
			"OPTION_J": "echo \"\r\"",
		}
		assert.Equal(t, want, have)
	})

	t.Run("error - invalid content on line 1", func(t *testing.T) {
		// --- Given ---
		content := "" +
			"INVALID LINE\n" +
			"foo=bar"

		// --- When ---
		have := map[string]string{}
		err := Parse(have, strings.NewReader(content))

		// --- Then ---
		assert.Error(t, err)
		assert.Equal(t, "invalid config line: 1", err.Error())
		assert.Equal(t, 0, len(have))
	})

	t.Run("error - invalid content on line 2", func(t *testing.T) {
		// --- Given ---
		content := "" +
			"VALID=ok\n" +
			"INVALID LINE"

		// --- When ---
		have := map[string]string{}
		err := Parse(have, strings.NewReader(content))

		// --- Then ---
		assert.Error(t, err)
		assert.Equal(t, "invalid config line: 2", err.Error())
		assert.Equal(t, 1, len(have))
	})

	t.Run("error - reading", func(t *testing.T) {
		// --- When ---
		have := map[string]string{}
		err := Parse(have, iotest.ErrReader(errors.New("test error")))

		// --- Then ---
		assert.Error(t, err)
		assert.Equal(t, "dotenv: test error", err.Error())
		assert.Equal(t, 0, len(have))
	})
}

func Test_parseLine_interpolation(t *testing.T) {
	tt := []struct {
		testN string
		line  string
		with  map[string]string
		wKey  string
		wVal  string
	}{
		{
			"expands variables found in values",
			"BAR=$FOO",
			map[string]string{"FOO": "test"},
			"BAR",
			"test",
		},
		{
			"parses variables wrapped in brackets",
			"BAR=${FOO} bar",
			map[string]string{"FOO": "test"},
			"BAR",
			"test bar",
		},
		{
			"expands undefined variables to an empty string",
			"BAR=$FOO",
			map[string]string{},
			"BAR",
			"",
		},
		{
			"expands variables in double-quoted strings",
			`BAR="quote $FOO"`,
			map[string]string{"FOO": "test"},
			"BAR",
			"quote test",
		},
		{
			"does not expand variables in single quoted strings",
			"BAR='quote $FOO'",
			map[string]string{},
			"BAR",
			"quote $FOO",
		},
		{
			"does not expand escaped variables",
			`FOO="foo\$BAR"`,
			map[string]string{},
			"FOO",
			"foo$BAR",
		},
		{
			"does not expand escaped variables",
			`FOO="foo\${BAR}"`,
			map[string]string{},
			"FOO",
			"foo${BAR}",
		},
		{
			"does not expand escaped variables",
			`BAR="foo\\${FOO} ${FOO}"`,
			map[string]string{"FOO": "test"},
			"BAR",
			"foo${FOO} test",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			hKey, hVal, err := parseLine(123, tc.line, tc.with)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.wKey, hKey)
			assert.Equal(t, tc.wVal, hVal)
		})
	}
}

func Test_parseLine(t *testing.T) {
	tt := []struct {
		testN string

		lin  string
		wKey string
		wVal string
	}{
		{"key value", "FOO=bar", "FOO", "bar"},
		{"space after the key", "FOO =bar", "FOO", "bar"},
		{"space after equal sign", "FOO= bar", "FOO", "bar"},
		{"quoted value", `FOO="bar"`, "FOO", "bar"},
		{"single quote value", "FOO='bar'", "FOO", "bar"},
		{
			"value with escaped character",
			`FOO="escaped\"bar"`,
			"FOO",
			`escaped"bar`,
		},
		{"double quoted value", `FOO="'d'"`, "FOO", `'d'`},
		{"key colon trimmed value", "OPTION_A: 1", "OPTION_A", "1"},
		{
			"key colon trimmed value with equal sign",
			"OPTION_A: Foo=bar",
			"OPTION_A",
			"Foo=bar",
		},
		{"value with colon", "OPTION_A=1:B", "OPTION_A", "1:B"},
		{"exported key", "export OPTION_A=2", "OPTION_A", "2"},
		{
			"exported and single quoted",
			`export OPTION_B='\n'`,
			"OPTION_B",
			"\\n",
		},
		{"camel case key", "export exportFoo=2", "exportFoo", "2"},
		{"key prefixed with export", "exportFOO=2", "exportFOO", "2"},
		{
			"key prefixed with export joined with an underscore",
			"export_FOO =2",
			"export_FOO",
			"2",
		},
		{
			"key prefixed with export joined with a period",
			"export.FOO= 2",
			"export.FOO",
			"2",
		},
		{"tab between export and key", "export\tOPTION_A=2", "OPTION_A", "2"},
		{"spaces before export", "  export OPTION_A=2", "OPTION_A", "2"},
		{"tab before export", "\texport OPTION_A=2", "OPTION_A", "2"},
		{"multi line value", `FOO="bar\nbaz"`, "FOO", "bar\nbaz"},
		{"key separated with a period", "FOO.BAR=foobar", "FOO.BAR", "foobar"},
		{"multiple equal signs", "FOO=foobar=", "FOO", "foobar="},
		{"value ending with space", "FOO=bar ", "FOO", "bar"},
		{"line with comment", "FOO=bar # this is foo", "FOO", "bar"},
		{
			"line with comment containing hash",
			"FOO=bar # this # is foo",
			"FOO",
			"bar",
		},
		{
			"double-quoted value with hash and a line comment",
			`FOO="bar#baz" # comment`,
			"FOO",
			"bar#baz",
		},
		{"single-quoted value", "FOO='bar#baz' # comment", "FOO", "bar#baz"},
		{
			"double-quoted value with hashes and line comment",
			`FOO="bar#baz#bang" # comment`,
			"FOO",
			"bar#baz#bang",
		},
		{
			"double-quoted value with escaped characters",
			`FOO="bar\n\ b\az"`,
			"FOO",
			"bar\n baz",
		},
		{
			"double-quoted value with escaped characters 1",
			`FOO="bar\\\n\ b\az"`,
			"FOO",
			"bar\\\n baz",
		},
		{
			"double-quoted value with escaped characters 2",
			`FOO="bar\\r\ b\az"`,
			"FOO",
			"bar\\r baz",
		},
		{
			"escaped double quote and a line comment",
			`FOO="a\"b" # comment`,
			"FOO",
			`a"b`,
		},
		{"apostrophe and a line comment", "FOO=it's # note", "FOO", "it's"},
		{"line without a key name and with a value", `="value"`, "", "value"},
		{"value is a single double-quote character", `KEY="`, "KEY", "\""},
		{"value with not closed double-quote", `KEY="value`, "KEY", "\"value"},
		{"space between key name and equal sign", " KEY =value", "KEY", "value"},
		{"spaces before key name", "   KEY=value", "KEY", "value"},
		{"tab before key name", "\tKEY=value", "KEY", "value"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			m := make(map[string]string)

			// --- When ---
			hKey, hVal, err := parseLine(0, tc.lin, m)

			// --- Then ---
			assert.NoError(t, err)
			assert.Equal(t, tc.wKey, hKey)
			assert.Equal(t, tc.wVal, hVal)
			assert.Len(t, 0, m)
		})
	}
}

func Test_parseLine_error(t *testing.T) {
	tt := []struct {
		testN string

		lin  string
		want string
	}{
		{"line with invalid format", "lol$wut", "invalid config line: 0"},
		{"empty line", "", "invalid config line: 1"},
	}

	for n, tc := range tt {
		n := n
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			m := make(map[string]string)

			// --- When ---
			key, val, err := parseLine(n, tc.lin, m)

			// --- Then ---
			assert.Error(t, err)
			if !errors.Is(err, ErrInvLine) {
				t.Error("expected err to have ErrInvDotEnvLine in the chain")
				return
			}
			assert.Equal(t, tc.want, err.Error())
			assert.Equal(t, "", key)
			assert.Equal(t, "", val)
		})
	}
}

func Test_expandVariables_tabular(t *testing.T) {
	tt := []struct {
		testN string

		input string
		with  map[string]string
		want  string
	}{
		{
			"simple parameter expansion",
			"BAR=$FOO",
			map[string]string{"FOO": "test"},
			"BAR=test",
		},
		{
			"simple parameter expansion",
			"BAR=$FOO$FOO",
			map[string]string{"FOO": "ABC"},
			"BAR=ABCABC",
		},
		{
			"parameter expansion",
			"BAR=${FOO} bar",
			map[string]string{"FOO": "test"},
			"BAR=test bar",
		},
		{
			"expands undefined variables to an empty string",
			"BAR=$FOO",
			map[string]string{},
			"BAR=",
		},
		{
			"expands variables in double-quoted strings",
			"BAR=\"quote $FOO\"",
			map[string]string{"FOO": "test"},
			`BAR="quote test"`,
		},
		{
			"does not expand escaped variables",
			`FOO="foo\$BAR"`,
			map[string]string{},
			`FOO="foo$BAR"`,
		},
		{
			"does not expand escaped variables in braces",
			`FOO="foo\${BAR}"`,
			map[string]string{},
			`FOO="foo${BAR}"`,
		},
		{
			"expand only not escaped variables",
			"BAR=\"foo\\${FOO} ${FOO}\"",
			map[string]string{"FOO": "test"},
			`BAR="foo${FOO} test"`,
		},
		{
			"line without variables",
			"$%^&*",
			map[string]string{"BAR": "aaa"},
			"$%^&*",
		},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := expandVariables(tc.input, tc.with)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_extractKey_tabular(t *testing.T) {
	tt := []struct {
		testN string

		inp  string
		want string
	}{
		{"bare key", "FOO", "FOO"},
		{"leading space", "  FOO", "FOO"},
		{"trailing space", "FOO  ", "FOO"},
		{"leading and trailing space", "  FOO  ", "FOO"},
		{"leading tab", "\tFOO", "FOO"},
		{"trailing tab", "FOO\t", "FOO"},
		{"key with space before equals", "FOO ", "FOO"},
		{"empty string", "", ""},
		{"whitespace only", "   ", ""},
		{"export with space", "export FOO", "FOO"},
		{"export with tab", "export\tFOO", "FOO"},
		{"export with multiple spaces", "export  FOO", "FOO"},
		{"export with multiple tabs", "export\t\tFOO", "FOO"},
		{"export with leading space", "  export FOO", "FOO"},
		{"export with leading tab", "\texport FOO", "FOO"},
		{"export camel case key", "export exportFoo", "exportFoo"},
		{"export alone", "export", "export"},
		{"export no separator", "exportFOO", "exportFOO"},
		{"export underscore prefix", "export_FOO", "export_FOO"},
		{"export period prefix", "export.FOO", "export.FOO"},
		{"export digit prefix", "export1FOO", "export1FOO"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := extractKey(tc.inp)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_stripComment_tabular(t *testing.T) {
	tt := []struct {
		testN string

		lin  string
		want string
	}{
		{"empty string", "", ""},
		{"no hash", "FOO=bar", "FOO=bar"},
		{"inline comment", "FOO=bar # comment", "FOO=bar"},
		{"inline comment no space before hash", "FOO=bar# comment", "FOO=bar"},
		{"inline comment no space or text after hash", "FOO=bar#", "FOO=bar"},
		{"multiple spaces before hash", "FOO=bar   # comment", "FOO=bar"},
		{"tab before hash", "FOO=bar\t# comment", "FOO=bar"},
		{"multiple hashes unquoted", "FOO=bar # this # is foo", "FOO=bar"},
		{"comment only", "# comment", ""},
		{"hash only", "#", ""},
		{"hash at start no space", "#comment", ""},
		{"hash in double-quoted value", `FOO="bar#baz"`, `FOO="bar#baz"`},
		{"hash in single-quoted value", "FOO='bar#baz'", "FOO='bar#baz'"},
		{
			"hash after double-quoted with comment",
			`FOO="bar" # comment`,
			`FOO="bar"`,
		},
		{
			"hash after single-quoted with comment",
			"FOO='bar' # comment",
			"FOO='bar'",
		},
		{
			"multiple hashes in double-quoted",
			`FOO="bar#baz#bang" # comment`,
			`FOO="bar#baz#bang"`,
		},
		{
			"multiple hashes in single-quoted",
			"FOO='bar#baz#bang' # comment",
			"FOO='bar#baz#bang'",
		},
		{
			"single quote inside double-quoted with hash",
			`FOO="bar'#'baz" # comment`,
			`FOO="bar'#'baz"`,
		},
		{
			"double quote inside single-quoted with hash",
			"FOO='bar\"#\"baz' # comment",
			"FOO='bar\"#\"baz'",
		},
		{
			"empty double-quoted value with comment",
			`FOO="" # comment`,
			`FOO=""`,
		},
		{
			"empty single-quoted value with comment",
			"FOO='' # comment",
			"FOO=''",
		},
		{"apostrophe in unquoted value", "FOO=it's # note", "FOO=it's"},
		{"double quote in unquoted value", `FOO=5" # inches`, `FOO=5"`},
		{
			"escaped double quote in double-quoted",
			`FOO="a\"b" # comment`,
			`FOO="a\"b"`,
		},
		{"quoted value after colon", `FOO: "a#b" # comment`, `FOO: "a#b"`},
		{"hash in not closed double-quote", `FOO="a # b`, `FOO="a # b`},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := stripComment(tc.lin)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}

func Test_isIgnoredLine(t *testing.T) {
	tt := []struct {
		testN string

		lin  string
		want bool
	}{
		{"newline linux", "\n", true},
		{"newline windows", "\r\n", true},
		{"tabs", "\t\t", true},
		{"comment line", "# comment", true},
		{"tab and comment", "\t# comment", true},
		{"export line", `export OPTION_B='\n'`, false},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := isIgnoredLine(tc.lin)

			// --- Then ---
			assert.Equal(t, tc.want, have)
		})
	}
}
