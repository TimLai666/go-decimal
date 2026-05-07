# go-decimal

Fixed-point decimal arithmetic with a configurable Context (scale + rounding) and a compile-once expression engine.

## Features

- Fixed-point decimal core: store scaled integers to avoid binary float drift
- Context controls output scale and rounding mode
- Math functions: `Sqrt`, `Exp`, `Log`, `Pow`, `Cmp`
- Expression compiler: tokenize + shunting-yard + RPN VM (compile once, eval fast)
- `^` operator in expressions for exponentiation

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

| Mode                      | Direction                                     | `1.25` → 1dp | `-1.25` → 1dp |
| ------------------------- | --------------------------------------------- | ------------ | ------------- |
| `RoundingModeDown`        | toward zero                                   | `1.2`        | `-1.2`        |
| `RoundingModeUp`          | away from zero                                | `1.3`        | `-1.3`        |
| `RoundingModeCeiling`     | toward +∞                                     | `1.3`        | `-1.2`        |
| `RoundingModeFloor`       | toward −∞                                     | `1.2`        | `-1.3`        |
| `RoundingModeHalfUp`      | halves away from zero                         | `1.3`        | `-1.3`        |
| `RoundingModeHalfDown`    | halves toward zero                            | `1.2`        | `-1.2`        |
| `RoundingModeHalfEven`    | halves to even (banker's rounding)            | `1.2`        | `-1.2`        |
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