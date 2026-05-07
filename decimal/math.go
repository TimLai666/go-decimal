package decimal

import "math/big"

const guardDigits int32 = 12

var bigFive = big.NewInt(5)

// Cmp compares a and b and returns -1, 0, or 1 for a < b, a == b, or
// a > b respectively.
//
// The values are aligned to a common scale first, so 1.5, 1.50, and
// 1.500 all compare as equal. Cmp does not depend on a Context and
// never mutates its inputs.
func Cmp(a, b Decimal) int {
	if a.scale == b.scale {
		return a.i.Cmp(&b.i)
	}
	if a.scale < b.scale {
		var x big.Int
		x.Mul(&a.i, pow10(b.scale-a.scale))
		return x.Cmp(&b.i)
	}
	var y big.Int
	y.Mul(&b.i, pow10(a.scale-b.scale))
	return a.i.Cmp(&y)
}

// Sqrt returns √x at ctx.Scale digits, rounded with ctx.Mode.
//
// A negative x yields ErrNegativeSqrt; Sqrt(0) is 0.
//
// Internally x is scaled up to an integer N = x · 10^(2·ws) where
// ws = ctx.Scale + 12 guard digits, and big.Int.Sqrt produces the
// floor of √N. The result is then normalized back to ctx.Scale, so
// the final-digit error stays under 0.5 ULP in HalfUp mode.
func Sqrt(ctx Context, x Decimal) (Decimal, error) {
	sign := x.i.Sign()
	if sign < 0 {
		return Decimal{}, ErrNegativeSqrt
	}
	if sign == 0 {
		return ctx.Normalize(Decimal{}), nil
	}

	ws := max(ctx.Scale+guardDigits, 0)
	exp := 2*ws - x.scale
	if exp < 0 {
		ws = (x.scale+1)/2 + guardDigits
		exp = 2*ws - x.scale
	}

	var n big.Int
	n.Mul(&x.i, pow10(exp))

	var y big.Int
	y.Sqrt(&n)

	return ctx.Normalize(Decimal{i: y, scale: ws}), nil
}

// halveExact returns d/2 without precision loss by multiplying the
// underlying integer by 5 and increasing scale by 1.
func halveExact(d Decimal) Decimal {
	var z big.Int
	z.Mul(&d.i, bigFive)
	return Decimal{i: z, scale: d.scale + 1}
}

// Exp returns e^x at ctx.Scale digits, rounded with ctx.Mode.
//
// Defined on the entire real line. Exp(0) short-circuits to 1, and
// negative inputs are handled as 1 / Exp(|x|).
//
// Algorithm:
//  1. Take |x| and apply argument reduction by repeatedly halving until
//     |x| < 0.5 (k iterations). Halving is done by "multiply by 5,
//     scale + 1", which is exact.
//  2. Apply the Taylor series 1 + x + x²/2! + … to the reduced argument,
//     stopping once the next term rounds to zero at the working scale.
//  3. Square the accumulator k times to recover e^x for the original
//     argument.
//  4. If x was negative, take the reciprocal of the final result.
//
// Beware that for very large |x| (say 1000) the result is itself a huge
// number; both runtime and output size grow with |x|, so the caller
// should keep that in mind.
func Exp(ctx Context, x Decimal) Decimal {
	if x.i.Sign() == 0 {
		return ctx.Normalize(NewFromInt64(ctx, 1))
	}

	work := Context{Scale: ctx.Scale + guardDigits, Mode: RoundingModeHalfUp}

	negative := x.i.Sign() < 0
	xa := x
	if negative {
		xa = Neg(x)
	}

	half := Decimal{scale: 1}
	half.i.SetInt64(5)

	k := 0
	for Cmp(xa, half) >= 0 {
		xa = halveExact(xa)
		k++
		if k > 2000 {
			break
		}
	}

	one := NewFromInt64(work, 1)
	sum := one
	term := one
	for n := 1; n < 100000; n++ {
		term = Mul(work, term, xa)
		nDec := NewFromInt64(work, int64(n))
		quo, err := Div(work, term, nDec)
		if err != nil {
			break
		}
		term = quo
		if term.i.Sign() == 0 {
			break
		}
		sum = Add(work, sum, term)
	}

	for range k {
		sum = Mul(work, sum, sum)
	}

	if negative {
		inv, err := Div(work, NewFromInt64(work, 1), sum)
		if err == nil {
			sum = inv
		}
	}

	return ctx.Normalize(sum)
}

// Log returns ln(x) (the natural logarithm) at ctx.Scale digits, rounded
// with ctx.Mode.
//
// Inputs ≤ 0 yield ErrNonPositiveLog. Log(1) is 0.
//
// Algorithm:
//  1. Repeatedly take sqrt(x) until x falls into [0.9, 1.1] (count
//     iterations). This shrinks the convergence radius of the series in
//     step 2 to something tiny.
//  2. With y close to 1, compute z = (y - 1) / (y + 1) and apply the
//     atanh series atanh(z) = z + z³/3 + z⁵/5 + ….
//  3. ln(y) = 2·atanh(z); then ln(x) = 2^count · ln(y).
//
// Even for |log10(x)| ≈ 100 the sqrt loop only needs ~20 passes, so the
// total cost scales roughly linearly with ctx.Scale.
func Log(ctx Context, x Decimal) (Decimal, error) {
	if x.i.Sign() <= 0 {
		return Decimal{}, ErrNonPositiveLog
	}

	work := Context{Scale: ctx.Scale + guardDigits, Mode: RoundingModeHalfUp}

	lower := Decimal{scale: 1}
	lower.i.SetInt64(9)
	upper := Decimal{scale: 1}
	upper.i.SetInt64(11)

	y := work.Normalize(x)
	count := 0
	for Cmp(y, upper) > 0 || Cmp(y, lower) < 0 {
		var err error
		y, err = Sqrt(work, y)
		if err != nil {
			return Decimal{}, err
		}
		count++
		if count > 200 {
			break
		}
	}

	one := NewFromInt64(work, 1)
	num := Sub(work, y, one)
	den := Add(work, y, one)
	z, err := Div(work, num, den)
	if err != nil {
		return Decimal{}, err
	}

	z2 := Mul(work, z, z)
	sum := z
	term := z
	for n := 3; n < 1000000; n += 2 {
		term = Mul(work, term, z2)
		if term.i.Sign() == 0 {
			break
		}
		nDec := NewFromInt64(work, int64(n))
		quo, err := Div(work, term, nDec)
		if err != nil {
			break
		}
		if quo.i.Sign() == 0 {
			break
		}
		sum = Add(work, sum, quo)
	}

	// ln(x) = 2^(count+1) * sum (sum = atanh((y-1)/(y+1)), ln(y) = 2*sum)
	var multiplier big.Int
	multiplier.Lsh(bigOne, uint(count+1))
	mDec := work.Normalize(Decimal{i: multiplier, scale: 0})
	result := Mul(work, sum, mDec)

	return ctx.Normalize(result), nil
}

// Pow returns base^exp at ctx.Scale digits, rounded with ctx.Mode.
//
// Behavior splits on whether exp is an integer:
//
//   - Integer exp (including negative and zero): handled by binary
//     exponentiation (square-and-multiply) using only multiplication
//     and division, so there is no Log/Exp approximation in the loop.
//     base may be negative, e.g. Pow(-2, 3) = -8 and Pow(-2, 4) = 16.
//     Pow(0, exp) for exp < 0 returns ErrDivisionByZero; Pow(0, 0)
//     returns 1 by convention.
//
//   - Non-integer exp: computed as exp · ln(base) followed by Exp, so
//     base must be > 0. A non-positive base with a fractional exponent
//     returns ErrInvalidPow (the result would be complex-valued).
//
// "Integer" here means the scaled integer leaves no remainder modulo
// 10^scale: 2.00 is integer 2, but 2.001 is not.
func Pow(ctx Context, base, exp Decimal) (Decimal, error) {
	work := Context{Scale: ctx.Scale + guardDigits, Mode: RoundingModeHalfUp}

	if intExp, ok := tryAsBigInt(exp); ok {
		return powInt(ctx, work, base, intExp)
	}

	if base.i.Sign() <= 0 {
		return Decimal{}, ErrInvalidPow
	}

	logBase, err := Log(work, base)
	if err != nil {
		return Decimal{}, err
	}
	product := Mul(work, exp, logBase)
	return ctx.Normalize(Exp(work, product)), nil
}

func tryAsBigInt(d Decimal) (*big.Int, bool) {
	if d.scale == 0 {
		var z big.Int
		z.Set(&d.i)
		return &z, true
	}
	if d.scale < 0 {
		var z big.Int
		z.Mul(&d.i, pow10(-d.scale))
		return &z, true
	}
	divisor := pow10(d.scale)
	var quo, rem big.Int
	quo.QuoRem(&d.i, divisor, &rem)
	if rem.Sign() != 0 {
		return nil, false
	}
	return &quo, true
}

func powInt(ctx, work Context, base Decimal, exp *big.Int) (Decimal, error) {
	if exp.Sign() == 0 {
		return ctx.Normalize(NewFromInt64(ctx, 1)), nil
	}
	if base.i.Sign() == 0 {
		if exp.Sign() < 0 {
			return Decimal{}, ErrDivisionByZero
		}
		return ctx.Normalize(Decimal{}), nil
	}

	negExp := exp.Sign() < 0
	var absExp big.Int
	absExp.Abs(exp)

	result := NewFromInt64(work, 1)
	cur := work.Normalize(base)
	bits := absExp.BitLen()
	for i := range bits {
		if absExp.Bit(i) == 1 {
			result = Mul(work, result, cur)
		}
		if i < bits-1 {
			cur = Mul(work, cur, cur)
		}
	}

	if negExp {
		inv, err := Div(work, NewFromInt64(work, 1), result)
		if err != nil {
			return Decimal{}, err
		}
		result = inv
	}
	return ctx.Normalize(result), nil
}
