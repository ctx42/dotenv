## v0.7.1 (Sat, 17 Oct 2026 22:00:30 UTC)
- ci: set C42_CC_RUN_ID in the workflow environment.
- chore(deps): update dependencies.

## v0.7.0 (Fri, 16 Oct 2026 13:33:50 UTC)
- refactor(dotenv): split line scanning and splitting out of Parse.
- feat(dotenv): add Layers to build a configuration from layered sources.

## v0.6.1 (Sat, 10 Oct 2026 19:09:58 UTC)
- build(deps): update github.com/ctx42/testing to v0.57.0.

## v0.6.0 (Sat, 10 Oct 2026 08:52:02 UTC)
- feat(dotenv): add Override to apply environment values.

## v0.5.0 (Wed, 30 Sep 2026 13:38:10 UTC)
- fix(dotenv): only a leading quote opens a quoted value.
- fix(dotenv): keep command substitution verbatim.
- fix(dotenv): require balanced braces in variable references.
- fix(dotenv)!: return ErrNilMap when Parse is given a nil map.
- fix(dotenv)!: reject empty keys and keys containing whitespace.
- test(dotenv): order tests by declaration order in dotenv.go.
- test(dotenv): add _tabular suffix to table-driven test names.
- chore: update dependency ctx42/testing to v0.56.0.
- test(dotenv): shorten subtest names to the condition under test.
- test(dotenv): prefix parseLine error subtest names with "error - ".
- test(dotenv): make parseLine error table branch-free.
- test(dotenv): name parseLine error results hKey and hVal.
- test(dotenv): use ErrorEqual and Len assertions.
- test(dotenv): prepare Parse arguments in the Given step.
- test(dotenv): cover Split entries without a separator or key.
- fix(dotenv): accept lines up to 1 MiB.
- fix(dotenv): ignore a UTF-8 byte order mark on the first line.
- refactor(dotenv): prefix all errors with "dotenv:".
- style(dotenv): keep lines within 80 columns.
- docs(dotenv): fix godoc grammar.
- style(dotenv): break long table rows one element per line.
- fix(dotenv): require whitespace before an inline comment.
- fix(dotenv): decode all documented escapes in double-quoted values.
- fix(dotenv): expand variables after an escaped backslash.
- docs(dotenv): document Parse errors on invalid lines.
- refactor(dotenv): split environment entries with strings.Cut.
- test(dotenv): use b.Loop in benchmarks.
- test(dotenv): remove trailing spaces from example inputs.
- style(dotenv): separate the map assertion in parseLine tests.
- docs: update the parsing pipeline in AGENTS.md.
- fix(dotenv): do not close a double-quoted value on an escaped quote.
- fix(dotenv): decode UTF-16 surrogate pairs in unicode escapes.
- docs: drop regex patterns from the godotenv attribution.
- docs(dotenv): fix the short-value comment in parseValue.
- docs: inject README examples with gomake :doc:mce markers.
- docs: show the package import path in the README.
- docs: list both public functions in the README usage.
- docs: describe which keys Parse accepts and expands.
- docs: spell out the variable expansion rules.
- docs: limit README escape sequences to double-quoted values.
- docs: declare the dotenv language on README code fences.
- docs: add a README overview and feature list.
- docs: put README badges above the title and link the tests badge.
- docs: state the required Go version in the README.
- docs: correct the godotenv comparison in the README.
- docs: document the separator, input limits, and errors in the README.
- chore: bump version to v0.5.0.

## v0.5.0 (Wed, 30 Sep 2026 13:35:44 UTC)
- fix(dotenv): only a leading quote opens a quoted value.
- fix(dotenv): keep command substitution verbatim.
- fix(dotenv): require balanced braces in variable references.
- fix(dotenv)!: return ErrNilMap when Parse is given a nil map.
- fix(dotenv)!: reject empty keys and keys containing whitespace.
- test(dotenv): order tests by declaration order in dotenv.go.
- test(dotenv): add _tabular suffix to table-driven test names.
- chore: update dependency ctx42/testing to v0.56.0.
- test(dotenv): shorten subtest names to the condition under test.
- test(dotenv): prefix parseLine error subtest names with "error - ".
- test(dotenv): make parseLine error table branch-free.
- test(dotenv): name parseLine error results hKey and hVal.
- test(dotenv): use ErrorEqual and Len assertions.
- test(dotenv): prepare Parse arguments in the Given step.
- test(dotenv): cover Split entries without a separator or key.
- fix(dotenv): accept lines up to 1 MiB.
- fix(dotenv): ignore a UTF-8 byte order mark on the first line.
- refactor(dotenv): prefix all errors with "dotenv:".
- style(dotenv): keep lines within 80 columns.
- docs(dotenv): fix godoc grammar.
- style(dotenv): break long table rows one element per line.
- fix(dotenv): require whitespace before an inline comment.
- fix(dotenv): decode all documented escapes in double-quoted values.
- fix(dotenv): expand variables after an escaped backslash.
- docs(dotenv): document Parse errors on invalid lines.
- refactor(dotenv): split environment entries with strings.Cut.
- test(dotenv): use b.Loop in benchmarks.
- test(dotenv): remove trailing spaces from example inputs.
- style(dotenv): separate the map assertion in parseLine tests.
- docs: update the parsing pipeline in AGENTS.md.
- fix(dotenv): do not close a double-quoted value on an escaped quote.
- fix(dotenv): decode UTF-16 surrogate pairs in unicode escapes.
- docs: drop regex patterns from the godotenv attribution.
- docs(dotenv): fix the short-value comment in parseValue.
- docs: inject README examples with gomake :doc:mce markers.
- docs: show the package import path in the README.
- docs: list both public functions in the README usage.
- docs: describe which keys Parse accepts and expands.
- docs: spell out the variable expansion rules.
- docs: limit README escape sequences to double-quoted values.
- docs: declare the dotenv language on README code fences.
- docs: add a README overview and feature list.
- docs: put README badges above the title and link the tests badge.
- docs: state the required Go version in the README.
- docs: correct the godotenv comparison in the README.
- docs: document the separator, input limits, and errors in the README.

## v0.4.0 (Sun, 07 Jun 2026 20:28:47 UTC)
- chore: update dependency ctx42/testing to v0.51.0.

## v0.3.0 (Sun, 07 Jun 2026 20:28:06 UTC)
- docs: use gmdoceg task for README.md examples.
- docs: clarify godotenv attribution and add missing copyright.
- chore: clean up .editorconfig and expand .gitignore.
- perf(dotenv): replace regex hot-path with string operations.
- Add AGENTS.md file.

## v0.2.0 (Fri, 01 May 2026 19:37:55 UTC)
- doc: Edit README.md style.
- chore: update to Go 1.26.

## v0.1.0 (Sat, 04 Oct 2025 08:57:18 UTC)
- style: Add lint ignore in examples.

## v0.0.0 (Sat, 04 Oct 2025 08:55:41 UTC)
- Initial commit.

