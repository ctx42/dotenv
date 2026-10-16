// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dotenv

import (
	"strings"
	"testing"

	"github.com/ctx42/testing/pkg/assert"
	"github.com/ctx42/testing/pkg/must"
)

func Test_Layers_Read(t *testing.T) {
	t.Run("base layer", func(t *testing.T) {
		// --- Given ---
		lay := &Layers{}
		src := "" +
			"ROOT=/ctx42\n" +
			"# Comment.\n" +
			"GOPATH=$ROOT/go # Comment.\n"

		// --- When ---
		err := lay.Read(strings.NewReader(src))

		// --- Then ---
		assert.NoError(t, err)
		want := []entry{
			{key: "ROOT", raw: "/ctx42"},
			{key: "GOPATH", raw: "$ROOT/go"},
		}
		assert.Equal(t, want, lay.entries)
		assert.Equal(t, map[string][]int{"ROOT": {0}, "GOPATH": {1}}, lay.index)
	})

	t.Run("override replaces earlier layer key in place", func(t *testing.T) {
		// --- Given ---
		lay := newLayers(t, "A=1\nB=2\nA=3\n")

		// --- When ---
		err := lay.Read(strings.NewReader("A=x\nC=y\n"))

		// --- Then ---
		assert.NoError(t, err)
		want := []entry{
			{key: "A", raw: "x"},
			{key: "B", raw: "2"},
			{key: "A", raw: "x"},
			{key: "C", raw: "y"},
		}
		assert.Equal(t, want, lay.entries)
	})

	t.Run("new key repeated in one layer", func(t *testing.T) {
		// --- Given ---
		lay := newLayers(t, "A=1\n")

		// --- When ---
		err := lay.Read(strings.NewReader("C=1\nC=2\n"))

		// --- Then ---
		assert.NoError(t, err)
		want := []entry{
			{key: "A", raw: "1"},
			{key: "C", raw: "1"},
			{key: "C", raw: "2"},
		}
		assert.Equal(t, want, lay.entries)
		assert.Equal(t, []int{1, 2}, lay.index["C"])
	})

	t.Run("error - invalid line", func(t *testing.T) {
		// --- Given ---
		lay := &Layers{}

		// --- When ---
		err := lay.Read(strings.NewReader("A=1\ninvalid\nB=2\n"))

		// --- Then ---
		assert.ErrorIs(t, ErrInvLine, err)
		assert.ErrorEqual(t, "dotenv: invalid config line: 2", err)
		assert.Equal(t, []entry{{key: "A", raw: "1"}}, lay.entries)
	})
}

func Test_Layers_Expand(t *testing.T) {
	t.Run("override config moves derived key", func(t *testing.T) {
		// --- Given ---
		lay := newLayers(t, "ROOT=/ctx42\nGOPATH=$ROOT/go\n", "ROOT=/acme\n")

		// --- When ---
		have, keys := lay.Expand(nil)

		// --- Then ---
		want := map[string]string{"ROOT": "/acme", "GOPATH": "/acme/go"}
		assert.Equal(t, want, have)
		assert.Nil(t, keys)
	})

	t.Run("environment moves derived key", func(t *testing.T) {
		// --- Given ---
		lay := newLayers(t, "HOST=ghcr.io\nREPO=$HOST/ctx42\n")
		env := []string{"HOST=registry.acme.io"}

		// --- When ---
		have, keys := lay.Expand(env)

		// --- Then ---
		want := map[string]string{
			"HOST": "registry.acme.io",
			"REPO": "registry.acme.io/ctx42",
		}
		assert.Equal(t, want, have)
		assert.Equal(t, []string{"HOST"}, keys)
	})

	t.Run("explicit derived key kept", func(t *testing.T) {
		// --- Given ---
		lay := newLayers(
			t,
			"ROOT=/ctx42\nGOPATH=$ROOT/go\nGOBIN=$ROOT/bin\n",
			"ROOT=/acme\nGOPATH=/x\n",
		)
		env := []string{"GOBIN=/y"}

		// --- When ---
		have, keys := lay.Expand(env)

		// --- Then ---
		want := map[string]string{
			"ROOT":   "/acme",
			"GOPATH": "/x",
			"GOBIN":  "/y",
		}
		assert.Equal(t, want, have)
		assert.Equal(t, []string{"GOBIN"}, keys)
	})

	t.Run("override config adds key", func(t *testing.T) {
		// --- Given ---
		lay := newLayers(t, "ROOT=/ctx42\n", "ROOT=/acme\nNEW=$ROOT/new\n")

		// --- When ---
		have, _ := lay.Expand(nil)

		// --- Then ---
		want := map[string]string{"ROOT": "/acme", "NEW": "/acme/new"}
		assert.Equal(t, want, have)
	})

	t.Run("undeclared environment variable ignored", func(t *testing.T) {
		// --- Given ---
		lay := newLayers(t, "A=1\n")
		env := []string{"B=2", "malformed", "=3"}

		// --- When ---
		have, keys := lay.Expand(env)

		// --- Then ---
		assert.Equal(t, map[string]string{"A": "1"}, have)
		assert.Nil(t, keys)
	})

	t.Run("environment value used as is", func(t *testing.T) {
		// --- Given ---
		lay := newLayers(t, "ROOT=/ctx42\nGOPATH=$ROOT/go\n")
		env := []string{"ROOT=$HOME"}

		// --- When ---
		have, _ := lay.Expand(env)

		// --- Then ---
		want := map[string]string{"ROOT": "$HOME", "GOPATH": "$HOME/go"}
		assert.Equal(t, want, have)
	})

	t.Run("empty environment value overrides", func(t *testing.T) {
		// --- Given ---
		lay := newLayers(t, "A=1\nB=${A}b\n")

		// --- When ---
		have, keys := lay.Expand([]string{"A="})

		// --- Then ---
		assert.Equal(t, map[string]string{"A": "", "B": "b"}, have)
		assert.Equal(t, []string{"A"}, keys)
	})

	t.Run("later layer wins", func(t *testing.T) {
		// --- Given ---
		lay := newLayers(t, "A=1\n", "A=2\n", "A=3\n")

		// --- When ---
		have, _ := lay.Expand(nil)

		// --- Then ---
		assert.Equal(t, map[string]string{"A": "3"}, have)
	})

	t.Run("environment wins over every layer", func(t *testing.T) {
		// --- Given ---
		lay := newLayers(t, "A=1\nA=2\n", "A=3\n")

		// --- When ---
		have, keys := lay.Expand([]string{"A=x"})

		// --- Then ---
		assert.Equal(t, map[string]string{"A": "x"}, have)
		assert.Equal(t, []string{"A"}, keys)
	})

	t.Run("overridden keys sorted", func(t *testing.T) {
		// --- Given ---
		lay := newLayers(t, "B=1\nA=2\nC=3\n")

		// --- When ---
		_, keys := lay.Expand([]string{"C=x", "B=y"})

		// --- Then ---
		assert.Equal(t, []string{"B", "C"}, keys)
	})

	t.Run("quoting kept", func(t *testing.T) {
		// --- Given ---
		lay := newLayers(
			t,
			"ROOT=/ctx42\nRAW='$ROOT'\nDQ=\"$ROOT\\tx\"\n",
			"ROOT=/acme\n",
		)

		// --- When ---
		have, _ := lay.Expand(nil)

		// --- Then ---
		want := map[string]string{
			"ROOT": "/acme",
			"RAW":  "$ROOT",
			"DQ":   "/acme\tx",
		}
		assert.Equal(t, want, have)
	})

	t.Run("single layer matches Parse", func(t *testing.T) {
		// --- Given ---
		src := "A=1\nB=$A\nA=2\nC=${A}\nD='$A'\n"
		lay := newLayers(t, src)

		want := map[string]string{}
		must.Nil(Parse(want, strings.NewReader(src)))

		// --- When ---
		have, _ := lay.Expand(nil)

		// --- Then ---
		assert.Equal(t, want, have)
	})

	t.Run("zero value", func(t *testing.T) {
		// --- Given ---
		lay := &Layers{}

		// --- When ---
		have, keys := lay.Expand([]string{"A=1"})

		// --- Then ---
		assert.Equal(t, map[string]string{}, have)
		assert.Nil(t, keys)
	})
}
