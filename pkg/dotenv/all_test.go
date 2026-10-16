// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dotenv

import (
	"strings"

	"github.com/ctx42/testing/pkg/tester"
)

// newLayers returns [Layers] holding the given layers read in order.
func newLayers(t tester.T, layers ...string) *Layers {
	t.Helper()
	lay := &Layers{}
	for _, src := range layers {
		if err := lay.Read(strings.NewReader(src)); err != nil {
			t.Fatal(err)
		}
	}
	return lay
}
