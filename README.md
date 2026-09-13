# go-decimal

Fixed-point decimal arithmetic with a configurable Context (scale + rounding) and a compile-once expression engine.

## Features

- Fixed-point decimal core: store scaled integers to avoid binary float drift
- Context controls output scale and rounding mode
- Math functions: `Sqrt`, `Exp`, `Log`, `Pow`, `Cmp`
- Expression compiler: tokenize + shunting-yard + RPN VM (compile once, eval fast)
- `^` operator in expressions for exponentiation

## Comparison with other Go decimal libraries

Choose `go-decimal` when the number of decimal places is part of the rule you are
implementing. For example, an invoice total may always use two fractional digits
and a rate may always use six. Set that rule once in a `Context`, then use the
same scale and rounding mode for parsing, arithmetic, division, and math
functions.

This gives `go-decimal` three practical advantages:

- **Predictable output:** context-aware operations normalize results to the same
  `Context.Scale`, so a calculation does not silently change its displayed
  number of fractional digits.
- **Explicit rounding:** choose from nine rounding modes, then apply the same
  rule to every context-aware operation in the calculation.
- **Reusable formulas:** `expr.Compile` parses a formula once. Reuse the
  resulting program with new variables without reparsing the expression.

Libraries described as decimal floating-point usually control significant
precision and exponent instead of a fixed number of fractional digits. This
table focuses on where each design fits and where `go-decimal` is the better
choice. It is not a benchmark. Performance depends on operand size, scale, and
workload.

| Package | Main strength | Why `go-decimal` may be the better fit |
| --- | --- | --- |
| **[go-decimal](https://github.com/TimLai666/go-decimal)** | Fixed-point values backed by `big.Int`, a shared `Context` for scale and rounding, decimal math, and a compile-once expression engine | Choose this package when fixed fractional digits, explicit rounding, arithmetic functions that do not modify their inputs, and reusable formulas matter more than built-in database or serialization adapters |
| **[shopspring/decimal](https://github.com/shopspring/decimal)** | Arbitrary-precision fixed-point values with `database/sql`, JSON, and XML serialization | Choose `go-decimal` when one `Context` should define the scale and rounding used across a calculation, and when `Sqrt`, `Exp`, `Log`, `Pow`, or expression evaluation belong in the same package |
| **[cockroachdb/apd/v3](https://github.com/cockroachdb/apd)** | Arbitrary-precision decimal arithmetic based on much of the General Decimal Arithmetic specification, with precision, range, condition flags, and traps | Choose `go-decimal` when a fixed number of fractional digits and a smaller API are easier to reason about than precision and condition management |
| **[ericlagergren/decimal](https://github.com/ericlagergren/decimal)** | Arbitrary-precision decimal floating-point with General Decimal Arithmetic and Go modes, plus a broad math package | Choose `go-decimal` when your application needs fixed-point output and a formula evaluator instead of floating-point semantics and a separate math package |
| **[govalues/decimal](https://github.com/govalues/decimal)** | Decimal floating-point with 19 digits of precision, half-to-even rounding, SQL and serialization interfaces, and a no-heap-allocation design | Choose `go-decimal` when the scale and rounding mode must be configurable instead of fixed to 19-digit half-to-even arithmetic |

For API details and current support status, check each project's documentation
before choosing a dependency.

## Install

```
go get github.com/TimLai666/go-decimal
```

## Decimal usage

```go
ctx := decimal.Context{Scale: 2, Mode: decimal.RoundingModeHalfUp}

price := decimal.MustParse(ctx, "12.345") // 12.35
qty := decimal.MustParse(ctx, "2")

subtotal := decimal.Mul(ctx, price, qty) // 24.70

discount := decimal.MustParse(ctx, "1.01")

final := decimal.Sub(ctx, subtotal, discount)
fmt.Println(final.String()) // 23.69
```

### Rounding modes

Listed in roughly decreasing order of typical use. `RoundingModeHalfUp` is the zero value, so `Context{}` rounds half away from zero by default.

| Mode                      | Direction                                     | `1.25` → 1dp | `-1.25` → 1dp |
| ------------------------- | --------------------------------------------- | ------------ | ------------- |
| `RoundingModeHalfUp`      | halves away from zero (default)               | `1.3`        | `-1.3`        |
| `RoundingModeHalfEven`    | halves to even (banker's rounding)            | `1.2`        | `-1.2`        |
| `RoundingModeDown`        | toward zero                                   | `1.2`        | `-1.2`        |
| `RoundingModeUp`          | away from zero                                | `1.3`        | `-1.3`        |
| `RoundingModeCeiling`     | toward +∞                                     | `1.3`        | `-1.2`        |
| `RoundingModeFloor`       | toward −∞                                     | `1.2`        | `-1.3`        |
| `RoundingModeHalfDown`    | halves toward zero                            | `1.2`        | `-1.2`        |
| `RoundingMode05Up`        | step iff last kept digit is 0 or 5            | `1.2`        | `-1.2`        |
| `RoundingModeUnnecessary` | assert no rounding; panic if any is required  | panic        | panic         |

Compatibility:

- `HalfUp` / `HalfDown` / `HalfEven` match Java `BigDecimal.ROUND_HALF_*` and Python `decimal.ROUND_HALF_*`.
- `Ceiling` / `Floor` / `Down` correspond to IEEE 754 `roundToward{Positive,Negative,Zero}`.
- `HalfEven` is the IEEE 754 default and Python `decimal`'s default — prefer it for long sums to avoid the upward bias that `HalfUp` accumulates.
- `05Up` matches Python's `decimal.ROUND_05UP`, an accounting-oriented rule that avoids producing 5-ending digits unless they are exact.
- `Unnecessary` mirrors Java's `RoundingMode.UNNECESSARY`. When rounding would actually be required, the operation panics with `ErrRoundingNecessary`; recover with `errors.Is(r.(error), decimal.ErrRoundingNecessary)`.

All non-panicking operations normalize to `Context.Scale`.

## Math functions

```go
ctx := decimal.Context{Scale: 10, Mode: decimal.RoundingModeHalfUp}

decimal.Sqrt(ctx, decimal.MustParse(ctx, "2"))   // 1.4142135624
decimal.Exp(ctx, decimal.MustParse(ctx, "1"))    // 2.7182818285
decimal.Log(ctx, decimal.MustParse(ctx, "10"))   // 2.3025850930
decimal.Pow(ctx, decimal.MustParse(ctx, "2"),
                 decimal.MustParse(ctx, "10"))   // 1024.0000000000
```

`Pow` switches strategy automatically: integer exponents use square-and-multiply
(exact, supports negative bases), non-integer exponents go through
`Exp(exp · Log(base))` and require a positive base. Errors are returned for
`Sqrt(<0)`, `Log(≤0)`, `Pow(<0, fractional)`, and `Pow(0, <0)`.

## Expression usage

```go
ctx := decimal.Context{Scale: 2, Mode: decimal.RoundingModeHalfUp}

prog, err := expr.Compile("1.2 + x/3 + x^2")
if err != nil {
    log.Fatal(err)
}

vars := expr.MapVars{
    "x": decimal.MustParse(ctx, "10"),
}

res, err := prog.Eval(ctx, vars)
if err != nil {
    log.Fatal(err)
}

fmt.Println(res.String()) // 104.53
```

Operators recognized by `expr.Compile`: `+`, `-`, `*`, `/`, `^`, plus unary `+/-`
and parentheses. `^` is right-associative and binds tighter than `*` and `/`,
matching Python: `2^3^2 == 512`, `-2^2 == -4`, `(-2)^2 == 4`.

## Notes

- Fixed-point decimals: `value = int / 10^scale`
- Division is integer division with rounding to `Context.Scale`
- No scientific notation in literals (`1e5` is not accepted)

## Benchmarks

```
go test ./... -bench=. -benchmem
```
