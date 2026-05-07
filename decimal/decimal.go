// Package decimal provides arbitrary-precision fixed-point decimal arithmetic.
//
// A value is stored as the pair (i, scale) representing the rational
// number i / 10^scale. All operations route through a Context that
// pins the output Scale and rounding Mode, so results stay free of
// the cumulative drift that plagues binary floating point.
package decimal

import (
	"math/big"
	"strings"
)

// Decimal is a fixed-point decimal number whose value is i / 10^scale.
//
// scale is the number of fractional digits. With scale = 2, an i of 1234
// represents 12.34. Both fields are unexported; build a Decimal through
// NewFromScaledInt, NewFromInt64, Parse, or its variants.
//
// Decimal is a value type. Copying it is cheap and produces a fully
// independent value — there is no shared mutable state to worry about.
type Decimal struct {
	i     big.Int
	scale int32
}

// NewFromScaledInt builds a Decimal from a pre-scaled integer i and the
// given scale, so the resulting value is i / 10^scale.
//
// A nil i is treated as 0. The returned Decimal does not share storage
// with the supplied *big.Int, so the caller may keep mutating its own
// copy without affecting the result.
//
// This constructor performs no rounding or normalization; if you need
// the value to obey a Context's Scale, call Context.Normalize on it.
func NewFromScaledInt(i *big.Int, scale int32) Decimal {
	var z big.Int
	if i != nil {
		z.Set(i)
	}
	return Decimal{i: z, scale: scale}
}

// NewFromInt64 wraps an int64 as a Decimal and normalizes it to ctx.Scale.
//
// For example, with ctx.Scale = 2, NewFromInt64(ctx, 5) returns 5.00
// (i = 500, scale = 2).
func NewFromInt64(ctx Context, n int64) Decimal {
	d := NewFromScaledInt(big.NewInt(n), 0)
	return ctx.Normalize(d)
}

// ParseExact parses a decimal string and preserves the exact scale found
// in the input — no rounding is performed.
//
// The input accepts an optional leading sign and digits with or without
// a decimal point. An empty string, illegal characters, or more than one
// decimal point all yield ErrInvalidDecimal. Examples:
//
//	"123.45" → 12345 / 100
//	"-0.10"  → -10  / 100  (trailing zero is preserved)
//	".5"     → 5    / 10
//	"1."     → 1    / 1
//
// Scientific notation (e.g. "1e5") is not supported. Use Parse instead
// when you want the result clamped to a Context's Scale.
func ParseExact(s string) (Decimal, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Decimal{}, ErrInvalidDecimal
	}

	sign := int64(1)
	if s[0] == '+' || s[0] == '-' {
		if s[0] == '-' {
			sign = -1
		}
		s = s[1:]
	}
	if s == "" {
		return Decimal{}, ErrInvalidDecimal
	}

	digits := make([]byte, 0, len(s))
	var scale int32
	seenDot := false
	digitCount := 0
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch {
		case ch >= '0' && ch <= '9':
			digits = append(digits, ch)
			if seenDot {
				scale++
			}
			digitCount++
		case ch == '.':
			if seenDot {
				return Decimal{}, ErrInvalidDecimal
			}
			seenDot = true
		default:
			return Decimal{}, ErrInvalidDecimal
		}
	}

	if digitCount == 0 {
		return Decimal{}, ErrInvalidDecimal
	}

	var bi big.Int
	if _, ok := bi.SetString(string(digits), 10); !ok {
		return Decimal{}, ErrInvalidDecimal
	}

	if sign < 0 {
		bi.Neg(&bi)
	}

	return Decimal{i: bi, scale: scale}, nil
}

// Parse is ParseExact followed by normalization to ctx.Scale using ctx.Mode.
//
// Inputs with more fractional digits than ctx.Scale are rounded according
// to the rounding mode.
func Parse(ctx Context, s string) (Decimal, error) {
	d, err := ParseExact(s)
	if err != nil {
		return Decimal{}, err
	}
	return ctx.Normalize(d), nil
}

// MustParse is like Parse but panics on parse failure. Convenient for
// initializing package-level constants from string literals.
func MustParse(ctx Context, s string) Decimal {
	d, err := Parse(ctx, s)
	if err != nil {
		panic(err)
	}
	return d
}

// Neg returns -d. The scale is preserved and the result does not depend
// on any Context.
func Neg(d Decimal) Decimal {
	var z big.Int
	z.Neg(&d.i)
	return Decimal{i: z, scale: d.scale}
}

// Scale returns the number of fractional digits stored on the Decimal.
//
// 12.345 (i = 12345, scale = 3) returns 3. The scale is not collapsed
// when trailing digits are zero, so 12.30 (i = 1230, scale = 2) still
// returns 2.
func (d Decimal) Scale() int32 {
	return d.scale
}

// String renders the value in conventional decimal notation.
//
// Trailing zeros implied by scale are preserved exactly: Decimal{i:1230,
// scale:2} prints as "12.30". When scale is 0 the output has no decimal
// point. Negative values are prefixed with "-" before the absolute
// magnitude. The output never contains thousands separators or
// scientific notation.
func (d Decimal) String() string {
	if d.scale == 0 {
		return d.i.String()
	}

	sign := ""
	if d.i.Sign() < 0 {
		sign = "-"
	}

	var abs big.Int
	abs.Abs(&d.i)
	digits := abs.String()

	scale := int(d.scale)
	if len(digits) <= scale {
		zeros := strings.Repeat("0", scale-len(digits)+1)
		digits = zeros + digits
	}

	cut := len(digits) - scale
	return sign + digits[:cut] + "." + digits[cut:]
}
