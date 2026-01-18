# go-decimal

Fixed-point decimal arithmetic with a configurable Context (scale + rounding) and a compile-once expression engine.

## Features

- Fixed-point decimal core: store scaled integers to avoid binary float drift
- Context controls output scale and rounding mode
- Expression compiler: tokenize + shunting-yard + RPN VM (compile once, eval fast)

## Install

```
go get github.com/tingzhen/go-decimal
```

## Decimal usage

```go
ctx := decimal.Context{Scale: 2, Mode: decimal.RoundHalfUp}

price := decimal.MustParse(ctx, "12.345") // 12.35
qty := decimal.MustParse(ctx, "2")

subtotal := decimal.Mul(ctx, price, qty) // 24.70

discount := decimal.MustParse(ctx, "1.01")

final := decimal.Sub(ctx, subtotal, discount)
fmt.Println(final.String()) // 23.69
```

### Rounding modes

- `RoundDown`: toward zero
- `RoundUp`: away from zero
- `RoundHalfUp`: halves away from zero

All operations normalize to `Context.Scale`.

## Expression usage

```go
ctx := decimal.Context{Scale: 2, Mode: decimal.RoundHalfUp}

prog, err := expr.Compile("1.2 + x/3")
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

fmt.Println(res.String()) // 4.53
```

## Notes

- Fixed-point decimals: `value = int / 10^scale`
- Division is integer division with rounding to `Context.Scale`
- No exponent notation or math functions in v1

## Benchmarks

```
go test ./... -bench=. -benchmem
```