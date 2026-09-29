# gosmell

A golangci-lint plugin (`gosmell`) that detects Go code smells (CS101–CS112:
long methods/classes, duplication, dead code, swallowed errors and more), plus a
standalone CLI that runs it through golangci-lint and prints a pretty, scored report.

Inspired by [Sloppy](https://github.com/Heyosseus/sloppy), a PHP static analyzer with the
same rule-code style (`SL101` God Method, `SL102` God Class, ...) and scored-report format
- this project ports that idea to Go, on top of `go/analysis` and golangci-lint instead of
PHP's AST tooling.

## Requirements

- Go 1.26+ (the toolchain is downloaded automatically via `GOTOOLCHAIN=auto` if you
  have an older `go` on PATH — no manual install needed)
- [golangci-lint](https://golangci-lint.run/) v2.x installed and on PATH

## Project layout

```
rules/            the 12 analyzers (CS101-CS112), one file each
report/           shared Finding/severity/confidence model
plugin.go         golangci-lint module-plugin entry point ("gosmell" linter)
cmd/gosmell/      standalone CLI: runs golangci-lint and renders the report
.custom-gcl.yml   recipe used to build a custom golangci-lint binary with this plugin
```

## Build

From this directory:

```bash
# 1. Build a golangci-lint binary with the gosmell plugin baked in
golangci-lint custom
# -> produces ./custom-gcl

# 2. Build the CLI that drives it and renders the report
go build -o gosmell ./cmd/gosmell
```

Re-run `golangci-lint custom` whenever you change anything under `rules/`, `report/`
or `plugin.go` — `gosmell` (the CLI) does not need rebuilding unless you change
`cmd/gosmell/main.go` itself.

## Testing

```bash
go test ./...
```

Every rule has its own `_test.go` (one per file in `rules/`), each compiling a small
throwaway module through `golang.org/x/tools/go/packages` and feeding it to the rule's
`CollectXFindings` function directly — the same `*analysis.Pass` shape golangci-lint
builds in production. `report/`, the plugin's config decoding, and the CLI's pure
helpers are covered the same way. There's no comment anywhere in the `.go` source
(including test files) — that's a deliberate project convention, not an oversight;
identifiers are named to make the "what", and test names plus failure messages carry
the "why".

## Usage

Run it against any Go project — no `.golangci.yml` setup needed in that project,
`gosmell` generates its own temporary config for the run and cleans it up
afterwards:

```bash
cd /path/to/your/project
/path/to/gosmell/gosmell ./...
```

Output looks like:

```
223/223 [▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓] 100%

  Gosmell

  Score  84/100  Clean
  Files  2 analysed  ·  43 lines  ·  4 findings

  pkg/a.go

     5  HIGH     CS101  Long Method  (78% confidence)
        BigOne() spans 14 lines with 13 statements, cyclomatic complexity 6, nesting
        depth 5 and 0 calls to 0 distinct collaborators.
        → Identify the distinct jobs this method performs and extract each into its
          own well-named method or class. ...

  ────────────────────────────────────────────────────────────

  4 findings: 2 high, 2 medium

  ✗ Failed — 2 finding(s) at high or above (fail_on: high)
```

Exit code is `1` if any HIGH-severity finding is present, `0` otherwise.

Output is colored (same palette as [Sloppy](https://github.com/Heyosseus/sloppy)'s own
console formatter: red for high/bad, yellow for medium, cyan for low/hints, gray for
secondary text, green for good) when stdout is a real terminal. Colors turn off
automatically when piped or redirected, and always with `NO_COLOR=1` set
(see [no-color.org](https://no-color.org/)).

### Which golangci-lint binary it uses

`gosmell` needs a golangci-lint binary that has this plugin built in. It resolves
one in this order:

1. `$GOSMELL_LINT_BIN`, if set
2. a `custom-gcl` binary sitting next to the `gosmell` executable (the default —
   just keep both binaries in the same directory, as built above)
3. `golangci-lint` on `PATH`, as a last resort (only works if that binary already
   has the plugin built in)

To point at a `custom-gcl` built somewhere else:

```bash
GOSMELL_LINT_BIN=/somewhere/else/custom-gcl ./gosmell ./...
```

### Passing a specific pattern

The first argument is the package pattern (default `./...`):

```bash
./gosmell ./internal/...
```

## Using it as a plain golangci-lint linter instead

If you'd rather wire `gosmell` into a project's own golangci-lint pipeline
(alongside other linters, its own config, CI, etc.) instead of using the CLI,
add this to that project's `.golangci.yml`:

```yaml
version: "2"

linters:
  enable:
    - gosmell
  settings:
    custom:
      gosmell:
        type: module
        description: "detects Go code smells (CS101-CS112)"
        original-url: github.com/bartollo/gosmell
```

Then run that project with the `custom-gcl` binary built above instead of the
regular `golangci-lint`:

```bash
/path/to/gosmell/custom-gcl run ./...
```

## Configuring thresholds

Five rules have size/shape thresholds that are genuinely project-specific (a
legacy codebase and a fresh microservice don't agree on "too many statements").
Their defaults live in `rules/config.go` (`DefaultConfig()`); override just the
ones you need under the linter's `settings:` block:

```yaml
linters:
  settings:
    custom:
      gosmell:
        type: module
        original-url: github.com/bartollo/gosmell
        settings:
          longMethod:
            maxStatements: 60      # default 40
            maxComplexity: 15      # default 10
            maxNesting: 4          # default 3
            maxCalls: 30           # default 25
            maxCollaborators: 20   # default 15
          largeClass:
            maxFields: 15          # default 10
            maxMethods: 20         # default 15
            maxCollaborators: 12   # default 8
            maxStatements: 400     # default 300
            minExceeded: 3         # dimensions that must exceed at once, default 3
          excessiveNesting:
            maxDepth: 5            # default 4
          duplicateCode:
            minStatements: 8       # default 5, ignores bodies shorter than this
          copyPasteDrift:
            minSiblingGroup: 4     # default 3
            maxEditDistance: 1     # default 2
```

Every field is optional — anything you don't set keeps its default. An unknown
field name is a build error (`json: unknown field ...`), not a silently
ignored typo. Everything else - confidence weights, the CLI's score bands,
display widths - stays hardcoded on purpose: those are internal scoring/UI
details, not per-project knobs.

Using the standalone `gosmell` CLI instead? Drop a `.gosmell.yml` in the
project you're scanning with the same override shape (no `settings:` wrapper
needed - just the rule sections directly), and the CLI folds it into the
temporary config it generates for that run:

```yaml
# .gosmell.yml, in the project gosmell is run against
longMethod:
  maxStatements: 60
```

## Rules

| Code | Name | Severity | What it detects |
|---|---|---|---|
| CS101 | Long Method | High | Methods that are excessively long, branchy or chatty across several independent measurements (statements, cyclomatic complexity, nesting depth, calls, collaborators). |
| CS102 | Large Class | High | Structs that are large across several dimensions at once: size, method count, injected dependencies and collaborators. |
| CS103 | Excessive Nesting | Medium | Methods whose conditionals, loops and blocks nest deeper than the configured limit. |
| CS104 | Duplicate Code | Medium | Method bodies that are structurally identical to another method in the project. |
| CS105 | Dead Code | Medium | Unexported functions and methods with no visible reference anywhere in the declaring package. |
| CS106 | Unused Constructor Dependency | Medium | Constructor-injected dependencies that are never read anywhere in the class. |
| CS107 | Swallowed Exception | High | An error checked but left empty (`if err != nil {}`), or explicitly discarded (`_ = err`, `_ = call()`, `v, _ := call()`) without logging, wrapping or returning it. |
| CS108 | Redundant Condition | Low | A condition re-tested inside itself, repeated within an if/elseif chain, duplicated across a boolean operator, or written as a literal true/false. |
| CS109 | Comments | Low | Short comments whose every meaningful word already appears in the statement directly below them. |
| CS110 | Defensive Programming Noise | Low | A method that guards the same subject the same way twice, with no reassignment in between. |
| CS111 | Copy-Paste Drift | High | (Heuristic) A method body nearly identical to its siblings, where the one difference looks like an unfinished copy. |
| CS112 | Placeholder Implementation | High | A body that signals it's unfinished: elided-code comments, a not-implemented panic/error, or a TODO over an empty/constant return. |

Names for CS101, CS102, CS104, CS105 and CS109 are taken directly from
[refactoring.guru's code-smell catalog](https://refactoring.guru/refactoring/smells)
(see the coverage section below) — the rest have no equivalent there and keep
a descriptive name of their own.

Confidence percentages and the overall score are heuristics computed locally
(based on how far metrics exceed their thresholds, and a severity-weighted
penalty), not a statistical model — treat them as a ranking signal, not ground
truth. CS104 and CS111 in particular are pattern-matching heuristics and can
produce false positives; investigate before "fixing" what they flag.

## Coverage vs. the classic code-smell catalog

[refactoring.guru's catalog](https://refactoring.guru/refactoring/smells) has
22 smells across 5 categories. Three don't apply here and are left out
entirely:

- **Refused Bequest** and **Parallel Inheritance Hierarchies** are defined in
  terms of class inheritance hierarchies, which Go doesn't have (no
  subclassing, no method overriding - only interfaces and struct embedding).
- **Data Class** (a class that's just fields with no behavior) isn't really a
  smell in Go: plain structs with no methods are normal and often correct
  (DTOs, config, API payloads). Flagging every data-only struct would be bad
  advice for this language.

The other 19 are listed below. A checked box means a rule in this plugin
targets that smell (possibly a narrower slice of it, noted in parentheses) -
unchecked means it's a real Go smell this plugin doesn't detect (yet).

**Bloaters**
- [x] Long Method — CS101
- [x] Large Class — CS102
- [ ] Primitive Obsession
- [ ] Long Parameter List
- [ ] Data Clumps

**Object-Orientation Abusers** (the category name is inherited from the source
catalog - Go isn't class-based OO, but the three smells below are about
type-based dispatch and design consistency, not classes or inheritance, so
they still apply)
- [ ] Alternative Classes with Different Interfaces
- [ ] Switch Statements (in Go: the same type-switch duplicated across the
      codebase instead of a method per type - not switches in general, which
      are often idiomatic)
- [ ] Temporary Field

**Change Preventers**
- [ ] Divergent Change
- [ ] Shotgun Surgery

**Dispensables**
- [x] Comments (partial — only comments that just narrate the line below
      them) — CS109
- [x] Duplicate Code — CS104, plus CS111 (Copy-Paste Drift) for
      near-duplicates with a suspicious one-line difference
- [x] Dead Code (partial — unexported functions/methods with no reference
      anywhere in the package) — CS105
- [ ] Lazy Class
- [ ] Speculative Generality

**Couplers**
- [ ] Feature Envy
- [ ] Inappropriate Intimacy
- [ ] Incomplete Library Class
- [ ] Message Chains
- [ ] Middle Man

`Divergent Change` and `Shotgun Surgery` need change-history/impact analysis
(why code changes, not just its current shape), which is out of scope for a
single-snapshot AST-based linter like this one — listed for completeness, not
as a near-term TODO.
