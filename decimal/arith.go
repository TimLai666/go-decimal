package decimal

import "math/big"

var (
	bigOne = big.NewInt(1)
	bigTwo = big.NewInt(2)
)

func Add(ctx Context, a, b Decimal) Decimal {
	a2, b2, scale := alignScales(a, b)
	var sum big.Int
	sum.Add(&a2.i, &b2.i)
	return normalize(Decimal{i: sum, scale: scale}, ctx.Scale, ctx.Mode)
}

func Sub(ctx Context, a, b Decimal) Decimal {
	a2, b2, scale := alignScales(a, b)
	var diff big.Int
	diff.Sub(&a2.i, &b2.i)
	return normalize(Decimal{i: diff, scale: scale}, ctx.Scale, ctx.Mode)
}

func Mul(ctx Context, a, b Decimal) Decimal {
	var prod big.Int
	prod.Mul(&a.i, &b.i)
	scale := a.scale + b.scale
	return normalize(Decimal{i: prod, scale: scale}, ctx.Scale, ctx.Mode)
}

func Div(ctx Context, a, b Decimal) (Decimal, error) {
	if b.i.Sign() == 0 {
		return Decimal{}, ErrDivisionByZero
	}

	exp := ctx.Scale + b.scale - a.scale

	var numer big.Int
	var denom big.Int
	if exp >= 0 {
		numer.Mul(&a.i, pow10(exp))
		denom.Set(&b.i)
	} else {
		numer.Set(&a.i)
		denom.Mul(&b.i, pow10(-exp))
	}

	var quo big.Int
	var rem big.Int
	quo.QuoRem(&numer, &denom, &rem)

	if rem.Sign() != 0 {
		resultNeg := (numer.Sign() < 0) != (denom.Sign() < 0)
		switch ctx.Mode {
		case RoundingModeUp:
			if resultNeg {
				quo.Sub(&quo, bigOne)
			} else {
				quo.Add(&quo, bigOne)
			}
		case RoundingModeHalfUp:
			var absRem big.Int
			absRem.Abs(&rem)
			absRem.Mul(&absRem, bigTwo)

			var absDenom big.Int
			absDenom.Abs(&denom)

			if absRem.Cmp(&absDenom) >= 0 {
				if resultNeg {
					quo.Sub(&quo, bigOne)
				} else {
					quo.Add(&quo, bigOne)
				}
			}
		case RoundingModeDown:
		default:
		}
	}

	return Decimal{i: quo, scale: ctx.Scale}, nil
}

func normalize(d Decimal, scale int32, mode RoundingMode) Decimal {
	if d.scale == scale {
		return d
	}

	if d.scale < scale {
		diff := scale - d.scale
		var z big.Int
		z.Mul(&d.i, pow10(diff))
		return Decimal{i: z, scale: scale}
	}

	diff := d.scale - scale
	divisor := pow10(diff)

	var quo big.Int
	var rem big.Int
	quo.QuoRem(&d.i, divisor, &rem)

	if rem.Sign() != 0 {
		switch mode {
		case RoundingModeUp:
			if d.i.Sign() < 0 {
				quo.Sub(&quo, bigOne)
			} else {
				quo.Add(&quo, bigOne)
			}
		case RoundingModeHalfUp:
			var absRem big.Int
			absRem.Abs(&rem)
			absRem.Mul(&absRem, bigTwo)
			if absRem.Cmp(divisor) >= 0 {
				if d.i.Sign() < 0 {
					quo.Sub(&quo, bigOne)
				} else {
					quo.Add(&quo, bigOne)
				}
			}
		case RoundingModeDown:
		default:
		}
	}

	return Decimal{i: quo, scale: scale}
}

func alignScales(a, b Decimal) (Decimal, Decimal, int32) {
	if a.scale == b.scale {
		return a, b, a.scale
	}

	if a.scale > b.scale {
		return a, rescaleUp(b, a.scale), a.scale
	}

	return rescaleUp(a, b.scale), b, b.scale
}

func rescaleUp(d Decimal, scale int32) Decimal {
	if d.scale == scale {
		return d
	}

	if d.scale > scale {
		return normalize(d, scale, RoundingModeDown)
	}

	diff := scale - d.scale
	var z big.Int
	z.Mul(&d.i, pow10(diff))
	return Decimal{i: z, scale: scale}
}
