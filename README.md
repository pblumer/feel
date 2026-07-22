# feel

A standalone [FEEL](https://www.omg.org/spec/DMN/) (Friendly Enough Expression
Language) engine for Go — the lexer, parser, type checker and compiler that
lower FEEL expressions into reusable, allocation-light Go closures.

FEEL is the expression language defined by the OMG **DMN** (Decision Model and
Notation) standard. This package is the FEEL front-end extracted from
[temis](https://github.com/pblumer/temis) so it can be reused on its own, e.g.
to evaluate business rules, decision-table cells, or user-authored expressions
in any Go program.

## Highlights

- **Compile once, evaluate many.** Parsing, type checking and lowering happen
  up front; the result is a `CompiledExpr` closure that evaluates on the hot
  path with minimal allocation.
- **Decimal numbers, never `float64`.** Arithmetic uses
  [`apd`](https://github.com/cockroachdb/apd) decimals, so results match FEEL /
  DMN semantics (no binary floating-point surprises).
- **Three-valued logic and pervasive null propagation**, per the spec.
- **A type checker** that reports positioned findings before you evaluate.
- **A catalog of built-in functions** — conversion, boolean, string, list,
  numeric, date/time, range, temporal, sort and context functions.
- **Unary tests** for decision-table input cells (`> 10`, `[1..5]`,
  `"Winter", "Spring"`, …).
- **Execution limits** (recursion depth, iteration and list-size caps) to keep
  evaluation of untrusted expressions bounded.
- Pure Go, one dependency (`apd/v3`); no cgo.

## Install

```sh
go get github.com/pblumer/feel
```

Requires Go 1.24 or newer.

## Quick start

```go
package main

import (
	"fmt"

	"github.com/pblumer/feel"
	"github.com/pblumer/feel/value"
)

func main() {
	// Declare the variables the expression may reference.
	env := feel.NewEnv("Season", "Guest Count")

	// Parse + type-check + compile into a reusable closure.
	expr, err := feel.CompileString(
		`if Season = "Winter" and Guest Count > 8 then "Spareribs" else "Salad"`,
		env,
	)
	if err != nil {
		panic(err)
	}

	// Evaluate: bind values by name and run the closure.
	out, err := expr(env.NewScope(map[string]value.Value{
		"Season":      value.Str("Winter"),
		"Guest Count": value.NumberFromInt64(10),
	}))
	if err != nil {
		panic(err)
	}
	fmt.Println(out) // Spareribs
}
```

A runnable version lives in [`example_test.go`](example_test.go).

## Packages

| Import path | Purpose |
|---|---|
| `github.com/pblumer/feel` | The engine: lexer, parser, AST, type system, type checker and compiler. Entry points: `CompileString`, `CompileStringWith`, `Parse`, `Compile`, `NewEnv`, `Typecheck`. |
| `github.com/pblumer/feel/value` | The runtime value model: the `Value` interface and its kinds (null, bool, number, string, temporal types, list, context, range, function) plus FEEL-conformant equality, ordering and arithmetic. |
| `github.com/pblumer/feel/builtins` | The built-in function catalog, bound at compile time. |

## Building values

Bind inputs with the constructors in the `value` package:

```go
value.Str("Winter")          // string
value.NumberFromInt64(10)    // number (decimal)
value.MustNumber("3.14")     // number from a decimal string
value.BoolOf(true)           // boolean
value.NewList(a, b, c)       // list
value.NewContext().Put(...)  // context (map)
```

## Unary tests

Decision-table input cells are FEEL *unary tests* — implicit predicates over the
cell's input value, referenced as `?`:

```go
env := feel.NewEnv(feel.InputVar) // declare "?" (feel.InputVar)
test, err := feel.CompileUnaryTest(`> 10`, env)
if err != nil {
	panic(err)
}
ok, err := feel.Matches(test, env.NewScope(map[string]value.Value{
	feel.InputVar: value.NumberFromInt64(15), // feel.InputVar == "?"
}))
// ok == true
```

## Execution limits

`feel.DefaultLimits()` returns sensible caps (recursion depth, iteration count,
list size). Build a scope with `env.NewScopeWithLimits(values, limits)` — or
share one `*feel.EvalState` across several evaluations with
`env.NewScopeShared` — to bound evaluation of untrusted input.

## Provenance & license

Extracted from [temis](https://github.com/pblumer/temis). Licensed under the
[Apache License 2.0](LICENSE).
