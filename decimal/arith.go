package decimal

import "math/big"

var (
	bigOne = big.NewInt(1)
	bigTwo = big.NewInt(2)
)

// Add returns a + b normalized to ctx.
//
// The two operands are aligned to the larger of their scales before being
// summed, so addition itself is exact; the only place rounding can sneak
// in is the final clamp to ctx.Scale (using ctx.Mode) when the aligned
// scale exceeds it.
func Add(ctx Context, a, b Decimal) Decimal {
	a2, b2, scale := alignScales(a, b)
	var sum big.Int
	sum.Add(&a2.i, &b2.i)
	return normalize(Decimal{i: sum, scale: scale}, ctx.Scale, ctx.Mode)
}

// Sub returns a - b normalized to ctx. Same semantics as Add but in the
// other direction.
func Sub(ctx Context, a, b Decimal) Decimal {
	a2, b2, scale := alignScales(a, b)
	var diff big.Int
	diff.Sub(&a2.i, &b2.i)
	return normalize(Decimal{i: diff, scale: scale}, ctx.Scale, ctx.Mode)
}

// Mul returns a * b normalized to ctx.
//
// The intermediate result is held at a.scale + b.scale and is exact;
// rounding (per ctx.Mode) only happens when collapsing the product down
// to ctx.Scale.
func Mul(ctx Context, a, b Decimal) Decimal {
	var prod big.Int
	prod.Mul(&a.i, &b.i)
	scale := a.scale + b.scale
	return normalize(Decimal{i: prod, scale: scale}, ctx.Scale, ctx.Mode)
}

// Div returns a / b at ctx.Scale digits of precision, rounded with ctx.Mode.
//
// Division by zero yields ErrDivisionByZero. Because division is an
// inherently approximate finite-precision operation, the result has a
// last-digit error of at most 0.5 ULP under HalfUp and at most 1 ULP
// under Up or Down.
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
		var absDenom big.Int
		absDenom.Abs(&denom)
		resultNeg := (numer.Sign() < 0) != (denom.Sign() < 0)
		if shouldStepAway(&quo, &rem, &absDenom, ctx.Mode, resultNeg) {
			if resultNeg {
				quo.Sub(&quo, bigOne)
			} else {
				quo.Add(&quo, bigOne)
			}
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
		resultNeg := d.i.Sign() < 0
		if shouldStepAway(&quo, &rem, divisor, mode, resultNeg) {
			if resultNeg {
				quo.Sub(&quo, bigOne)
			} else {
				quo.Add(&quo, bigOne)
			}
		}
	}

	return Decimal{i: quo, scale: scale}
}

// shouldStepAway decides whether a quotient that has been truncated toward
// zero should be moved one step further away from zero to satisfy mode.
// rem is the (non-zero) remainder of the truncating division and absDivisor
// is the positive divisor used to detect the halfway point. resultNeg gives
// the sign of the true (un-truncated) value, since quo can be 0 even when
// the true value is negative.
func shouldStepAway(quo, rem, absDivisor *big.Int, mode RoundingMode, resultNeg bool) bool {
	switch mode {
	case RoundingModeDown:
		return false
	case RoundingModeUp:
		return true
	case RoundingModeCeiling:
		return !resultNeg
	case RoundingModeFloor:
		return resultNeg
	case RoundingModeHalfUp, RoundingModeHalfEven:
		var twice big.Int
		twice.Abs(rem)
		twice.Mul(&twice, bigTwo)
		cmp := twice.Cmp(absDivisor)
		if cmp > 0 {
			return true
		}
		if cmp < 0 {
			return false
		}
		if mode == RoundingModeHalfUp {
			return true
		}
		return quo.Bit(0) == 1
	default:
		return false
	}
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
