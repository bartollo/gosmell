# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.1.0] - 2026-09-29

### Added

- `gosmell` plugin for golangci-lint (`type: module`), with 12 Go code smell
  detection rules (CS101-CS112):
  - **CS101** Long Method — methods that are excessively long, branchy or
    chatty across several independent measurements.
  - **CS102** Large Class — structs that are large across several dimensions
    at once (size, method count, injected dependencies, collaborators).
  - **CS103** Excessive Nesting — conditionals/loops/blocks nesting deeper
    than the configured limit.
  - **CS104** Duplicate Code — method bodies that are structurally identical
    to another method in the project.
  - **CS105** Dead Code — unexported functions and methods with no visible
    reference anywhere in the declaring package.
  - **CS106** Unused Constructor Dependency — constructor-injected
    dependencies that are never read.
  - **CS107** Swallowed Exception — errors checked and discarded without
    logging, wrapping or returning them.
  - **CS108** Redundant Condition — conditions re-tested, repeated, or
    written as a literal `true`/`false`.
  - **CS109** Comments — short comments that just narrate the line below
    them.
  - **CS110** Defensive Programming Noise — the same defensive check
    repeated with no reassignment in between.
  - **CS111** Copy-Paste Drift (heuristic) — method bodies nearly identical
    to their siblings, with a suspicious one-line difference.
  - **CS112** Placeholder Implementation — bodies that signal they're
    unfinished (TODOs, not-implemented panics/errors, empty/constant
    returns).
- Standalone `gosmell` CLI (`cmd/gosmell`) that drives golangci-lint with the
  plugin baked in and renders a colored, scored report (0-100), with a
  progress bar, severity (`HIGH`/`MEDIUM`/`LOW`) and a confidence percentage
  per finding.
- Automatic generation of a temporary golangci-lint config by the CLI, with
  no `.golangci.yml` setup required in the scanned project.
- `.gosmell.yml` support in the scanned project to override thresholds
  without golangci-lint's `settings:` wrapper.
- Configurable resolution of the `golangci-lint` binary to use
  (`$GOSMELL_LINT_BIN` → a `custom-gcl` binary next to the executable →
  `golangci-lint` on `PATH`).
- Colored output (palette inspired by [Sloppy](https://github.com/Heyosseus/sloppy)),
  automatically disabled when stdout isn't a terminal or when `NO_COLOR=1`
  is set.
- Exit code `1` when at least one `HIGH`-severity finding is present, `0`
  otherwise.
- Per-rule configurable thresholds (`longMethod`, `largeClass`,
  `excessiveNesting`, `duplicateCode`, `copyPasteDrift`), with defaults in
  `rules/config.go` and strict validation of unknown fields.
- Full test suite: one `_test.go` per rule, plus tests for the `report/`
  model, plugin config decoding, and the CLI.
- Initial documentation (`README.md`) covering build instructions, usage,
  integration as a plain golangci-lint linter, and a coverage table against
  the classic code-smell catalog from
  [refactoring.guru](https://refactoring.guru/refactoring/smells).

[0.1.0]: https://github.com/bartollo/gosmell/releases/tag/v0.1.0
