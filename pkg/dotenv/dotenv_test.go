// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dotenv

import (
	"bufio"
	"errors"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/ctx42/testing/pkg/assert"
)

func Test_Parse(t *testing.T) {
	t.Run("simple", func(t *testing.T) {
		// --- Given ---
		have := map[string]string{}

		content := "" +
			"ONE=1\n" +
			"TWO='2'\n" +
			"THREE = \"3\""
		rdr := strings.NewReader(content)

		// --- When ---
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

	t.Run("commented lines", func(t *testing.T) {
		// --- Given ---
		have := map[string]string{}

		content := "" +
			"\t#ONE=1\n" +
			"TWO='2'\n" +
			"#THREE = \"3\""
		rdr := strings.NewReader(content)

		// --- When ---
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
		have := map[string]string{}

		content := "" +
			"OPTION_A=1\n" +
			"OPTION_B=2\n" +
			"OPTION_C= 3\n" +
			"OPTION_D =4\n" +
			"OPTION_E = 5\n" +
			"OPTION_F =\n" +
			"OPTION_G=\n"
		rdr := strings.NewReader(content)

		// --- When ---
		err := Parse(have, rdr)

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

	t.Run("export prefix", func(t *testing.T) {
		// --- Given ---
		have := map[string]string{}

		content := "" +
			"export OPTION_A=2\n" +
			"export OPTION_B='\\n'"
		rdr := strings.NewReader(content)

		// --- When ---
		err := Parse(have, rdr)

		// --- Then ---
		assert.NoError(t, err)
		want := map[string]string{
			"OPTION_A": "2",
			"OPTION_B": "\\n",
		}
		assert.Equal(t, want, have)
	})

	t.Run("equal sign in value", func(t *testing.T) {
		// --- Given ---
		have := map[string]string{}

		url := "postgres://localhost:5432/database?sslmode=disable"
		rdr := strings.NewReader("export OPTION_A='" + url + "'\n")

		// --- When ---
		err := Parse(have, rdr)

		// --- Then ---
		assert.NoError(t, err)
		want := map[string]string{
			"OPTION_A": url,
		}
		assert.Equal(t, want, have)
	})

	t.Run("quoted values", func(t *testing.T) {
		// --- Given ---
		have := map[string]string{}

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
		rdr := strings.NewReader(content)

		// --- When ---
		err := Parse(have, rdr)

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

	t.Run("line longer than 64 KiB", func(t *testing.T) {
		// --- Given ---
		have := map[string]string{}

		val := strings.Repeat("a", 100<<10)
		rdr := strings.NewReader("KEY=" + val)

		// --- When ---
		err := Parse(have, rdr)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{"KEY": val}, have)
	})

	t.Run("byte order mark", func(t *testing.T) {
		// --- Given ---
		have := map[string]string{}

		rdr := strings.NewReader("\ufeffKEY=value")

		// --- When ---
		err := Parse(have, rdr)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{"KEY": "value"}, have)
	})

	t.Run("byte order mark before comment", func(t *testing.T) {
		// --- Given ---
		have := map[string]string{}

		rdr := strings.NewReader("\ufeff# comment\nKEY=value")

		// --- When ---
		err := Parse(have, rdr)

		// --- Then ---
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{"KEY": "value"}, have)
	})

	t.Run("error - invalid content on line 1", func(t *testing.T) {
		// --- Given ---
		have := map[string]string{}

		content := "" +
			"INVALID LINE\n" +
			"foo=bar"
		rdr := strings.NewReader(content)

		// --- When ---
		err := Parse(have, rdr)

		// --- Then ---
		assert.ErrorEqual(t, "dotenv: invalid config line: 1", err)
		assert.Len(t, 0, have)
	})

	t.Run("error - invalid content on line 2", func(t *testing.T) {
		// --- Given ---
		have := map[string]string{}

		content := "" +
			"VALID=ok\n" +
			"INVALID LINE"
		rdr := strings.NewReader(content)

		// --- When ---
		err := Parse(have, rdr)

		// --- Then ---
		assert.ErrorEqual(t, "dotenv: invalid config line: 2", err)
		assert.Len(t, 1, have)
	})

	t.Run("error - reading", func(t *testing.T) {
		// --- Given ---
		have := map[string]string{}

		rdr := iotest.ErrReader(errors.New("test error"))

		// --- When ---
		err := Parse(have, rdr)

		// --- Then ---
		assert.ErrorEqual(t, "dotenv: test error", err)
		assert.Len(t, 0, have)
	})

	t.Run("error - line too long", func(t *testing.T) {
		// --- Given ---
		have := map[string]string{}

		val := strings.Repeat("a", maxLineSize)
		rdr := strings.NewReader("KEY=" + val)

		// --- When ---
		err := Parse(have, rdr)

		// --- Then ---
		assert.ErrorIs(t, bufio.ErrTooLong, err)
		assert.Len(t, 0, have)
	})

	t.Run("error - nil map with empty input", func(t *testing.T) {
		// --- Given ---
		rdr := strings.NewReader("")

		// --- When ---
		err := Parse(nil, rdr)

		// --- Then ---
		assert.ErrorIs(t, ErrNilMap, err)
	})

	t.Run("error - nil map", func(t *testing.T) {
		// --- Given ---
		rdr := strings.NewReader("FOO=bar")

		// --- When ---
		err := Parse(nil, rdr)

		// --- Then ---
		assert.ErrorIs(t, ErrNilMap, err)
	})
}

func Test_parseLine_tabular(t *testing.T) {
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
		{
			"tab between export and key",
			"export\tOPTION_A=2",
			"OPTION_A",
			"2",
		},
		{
			"spaces before export",
			"  export OPTION_A=2",
			"OPTION_A",
			"2",
		},
		{"tab before export", "\texport OPTION_A=2", "OPTION_A", "2"},
		{"multi line value", `FOO="bar\nbaz"`, "FOO", "bar\nbaz"},
		{
			"key separated with a period",
			"FOO.BAR=foobar",
			"FOO.BAR",
			"foobar",
		},
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
		{
			"single-quoted value",
			"FOO='bar#baz' # comment",
			"FOO",
			"bar#baz",
		},
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
		{
			"apostrophe and a line comment",
			"FOO=it's # note",
			"FOO",
			"it's",
		},
		{"tab escape", `FOO="a\tb"`, "FOO", "a\tb"},
		{
			"form feed and backspace escapes",
			`FOO="a\fb\bc"`,
			"FOO",
			"a\fb\bc",
		},
		{"unicode escape", `FOO="caf\u00e9"`, "FOO", "caf\u00e9"},
		{"invalid unicode escape", `FOO="\u12"`, "FOO", "u12"},
		{"escaped single quote", `FOO="it\'s"`, "FOO", "it's"},
		{"lone double quote", `KEY="`, "KEY", "\""},
		{
			"value with not closed double-quote",
			`KEY="value`,
			"KEY",
			"\"value",
		},
		{
			"space between key name and equal sign",
			" KEY =value",
			"KEY",
			"value",
		},
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

func Test_parseLine_interpolation_tabular(t *testing.T) {
	tt := []struct {
		testN string
		line  string
		with  map[string]string
		wKey  string
		wVal  string
	}{
		{
			"variable",
			"BAR=$FOO",
			map[string]string{"FOO": "test"},
			"BAR",
			"test",
		},
		{
			"variable in braces",
			"BAR=${FOO} bar",
			map[string]string{"FOO": "test"},
			"BAR",
			"test bar",
		},
		{
			"undefined variable",
			"BAR=$FOO",
			map[string]string{},
			"BAR",
			"",
		},
		{
			"variable in double quotes",
			`BAR="quote $FOO"`,
			map[string]string{"FOO": "test"},
			"BAR",
			"quote test",
		},
		{
			"variable in single quotes",
			"BAR='quote $FOO'",
			map[string]string{},
			"BAR",
			"quote $FOO",
		},
		{
			"escaped variable",
			`FOO="foo\$BAR"`,
			map[string]string{},
			"FOO",
			"foo$BAR",
		},
		{
			"escaped variable in braces",
			`FOO="foo\${BAR}"`,
			map[string]string{},
			"FOO",
			"foo${BAR}",
		},
		{
			"escaped and not escaped variables",
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

func Test_parseLine_error_tabular(t *testing.T) {
	tt := []struct {
		testN string

		lin string
	}{
		{"error - invalid format", "lol$wut"},
		{"error - empty line", ""},
		{"error - empty key", `="value"`},
		{"error - space in key", "MY KEY=value"},
		{"error - tab in key", "MY\tKEY=value"},
		{"error - space in exported key", "export MY KEY=v"},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- Given ---
			m := make(map[string]string)

			// --- When ---
			hKey, hVal, err := parseLine(7, tc.lin, m)

			// --- Then ---
			assert.ErrorIs(t, ErrInvLine, err)
			want := "dotenv: invalid config line: 7"
			assert.ErrorEqual(t, want, err)
			assert.Equal(t, "", hKey)
			assert.Equal(t, "", hVal)
		})
	}
}

func Test_unescape_tabular(t *testing.T) {
	tt := []struct {
		testN string

		val  string
		want string
	}{
		{"no escapes", "abc", "abc"},
		{"newline", `a\nb`, "a\nb"},
		{"carriage return", `a\rb`, "a\rb"},
		{"tab", `a\tb`, "a\tb"},
		{"form feed", `a\fb`, "a\fb"},
		{"backspace", `a\bb`, "a\bb"},
		{"double quote", `a\"b`, `a"b`},
		{"single quote", `a\'b`, "a'b"},
		{"backslash", `a\\b`, `a\b`},
		{"escaped backslash before n", `a\\nb`, `a\nb`},
		{"dollar is kept escaped", `a\$B`, `a\$B`},
		{"unicode", `caf\u00e9`, "caf\u00e9"},
		{"unicode uppercase hex", `\u00C9`, "\u00c9"},
		{"unicode too short", `\u12`, "u12"},
		{"unicode not hex", `\uZZZZ`, "uZZZZ"},
		{"other character", `\a`, "a"},
		{"multi-byte character", `\` + "\u00e9", "\u00e9"},
		{"trailing backslash", `a\`, `a\`},
	}

	for _, tc := range tt {
		t.Run(tc.testN, func(t *testing.T) {
			// --- When ---
			have := unescape(tc.val)

			// --- Then ---
			assert.Equal(t, tc.want, have)
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
			"adjacent parameter expansions",
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
			"undefined variable",
			"BAR=$FOO",
			map[string]string{},
			"BAR=",
		},
		{
			"variable in double quotes",
			"BAR=\"quote $FOO\"",
			map[string]string{"FOO": "test"},
			`BAR="quote test"`,
		},
		{
			"escaped variable",
			`FOO="foo\$BAR"`,
			map[string]string{},
			`FOO="foo$BAR"`,
		},
		{
			"escaped variable in braces",
			`FOO="foo\${BAR}"`,
			map[string]string{},
			`FOO="foo${BAR}"`,
		},
		{
			"escaped and not escaped variables",
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
		{
			"command substitution",
			"BAR=$(FOO)",
			map[string]string{"FOO": "test"},
			"BAR=$(FOO)",
		},
		{
			"lowercase command substitution",
			"BAR=$(whoami)",
			map[string]string{},
			"BAR=$(whoami)",
		},
		{
			"not closed brace",
			"BAR=${FOO",
			map[string]string{"FOO": "test"},
			"BAR=${FOO",
		},
		{
			"closing brace without opening",
			"BAR=$FOO}",
			map[string]string{"FOO": "test"},
			"BAR=test}",
		},
		{
			"escaped dollar not followed by a name",
			`BAR=\$ 5`,
			map[string]string{},
			"BAR=$ 5",
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

func Test_isIgnoredLine_tabular(t *testing.T) {
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
		{
			"hash without space before",
			"FOO=bar# comment",
			"FOO=bar# comment",
		},
		{"trailing hash", "FOO=bar#", "FOO=bar#"},
		{"hash in password", "PASS=p#ss # comment", "PASS=p#ss"},
		{"hash in url", "URL=http://h/#frag", "URL=http://h/#frag"},
		{"hash starting value", "FOO=#comment", "FOO="},
		{
			"hash right after quoted value",
			`FOO="a"# comment`,
			`FOO="a"`,
		},
		{
			"multiple spaces before hash",
			"FOO=bar   # comment",
			"FOO=bar",
		},
		{"tab before hash", "FOO=bar\t# comment", "FOO=bar"},
		{
			"multiple hashes unquoted",
			"FOO=bar # this # is foo",
			"FOO=bar",
		},
		{"comment only", "# comment", ""},
		{"hash only", "#", ""},
		{"hash at start no space", "#comment", ""},
		{
			"hash in double-quoted value",
			`FOO="bar#baz"`,
			`FOO="bar#baz"`,
		},
		{
			"hash in single-quoted value",
			"FOO='bar#baz'",
			"FOO='bar#baz'",
		},
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
		{
			"quoted value after colon",
			`FOO: "a#b" # comment`,
			`FOO: "a#b"`,
		},
		{"hash in not closed double-quote", `FOO="a # b`, `FOO="a # b`},
		{"key only comment", "FOO # comment", "FOO"},
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
