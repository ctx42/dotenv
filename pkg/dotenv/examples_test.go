// SPDX-FileCopyrightText: (c) 2026 Rafal Zajac
// SPDX-License-Identifier: MIT

package dotenv_test

import (
	"fmt"
	"strings"

	"github.com/ctx42/testing/pkg/dump"

	"github.com/ctx42/dotenv/pkg/dotenv"
)

func ExampleParse_basic() {
	file := `
HELLO="hello"
WORLD=world
`

	m := map[string]string{}
	r := strings.NewReader(file)

	if err := dotenv.Parse(m, r); err != nil {
		panic(err)
	}

	fmt.Println(dump.Any(m))
	// Output:
	// map[string]string{
	//   "HELLO": "hello",
	//   "WORLD": "world",
	// }
}

func ExampleParse_comments() {
	file := `
# Comment.
QUOTED="a # b # c #" 
AFTER_SPACE=world # Comment.
`

	m := map[string]string{}
	r := strings.NewReader(file)

	if err := dotenv.Parse(m, r); err != nil {
		panic(err)
	}

	fmt.Println(dump.Any(m))
	// Output:
	// map[string]string{
	//   "AFTER_SPACE": "world",
	//   "QUOTED": "a # b # c #",
	// }
}

func ExampleParse_variableExpansion() {
	file := `
HELLO=hello 
WORLD=world
MESSAGE="${HELLO} ${WORLD}!"
`

	m := map[string]string{}
	r := strings.NewReader(file)

	if err := dotenv.Parse(m, r); err != nil {
		panic(err)
	}

	fmt.Println(m["MESSAGE"])
	// Output:
	// hello world!
}

func ExampleParse_variableExpansionFromEnvironment() {
	file := `
HELLO=hello 
WORLD=world
MESSAGE="${HELLO} ${WORLD}${EXCLAIM}"
`

	env := map[string]string{
		"HELLO":   "env-hello",
		"EXCLAIM": "!",
	}
	r := strings.NewReader(file)

	if err := dotenv.Parse(env, r); err != nil {
		panic(err)
	}

	fmt.Println(dump.Any(env))
	// Output:
	// map[string]string{
	//   "EXCLAIM": "!",
	//   "HELLO": "hello",
	//   "MESSAGE": "hello world!",
	//   "WORLD": "world",
	// }
}

func ExampleParse_preventVariableExpansion() {
	file := `
HELLO=hello
WORLD=world
MESSAGE='${HELLO} ${WORLD}!'
`

	m := map[string]string{}
	r := strings.NewReader(file)

	if err := dotenv.Parse(m, r); err != nil {
		panic(err)
	}

	fmt.Println(m["MESSAGE"])
	// Output:
	// ${HELLO} ${WORLD}!
}

func ExampleSplit() {
	// Convert os.Environ() output into a map suitable for
	// pre-seeding Parse with existing environment variables.
	env := dotenv.Split([]string{
		"HOME=/home/user",
		"PATH=/usr/local/bin:/usr/bin",
		"USER=alice",
	})

	fmt.Println(dump.Any(env))
	// Output:
	// map[string]string{
	//   "HOME": "/home/user",
	//   "PATH": "/usr/local/bin:/usr/bin",
	//   "USER": "alice",
	// }
}
