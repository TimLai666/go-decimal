package decimal

import (
	"math/big"
	"strings"
)

type Decimal struct {
	i     big.Int
	scale int32
}

func NewFromScaledInt(i *big.Int, scale int32) Decimal {
	var z big.Int
	if i != nil {
		z.Set(i)
	}
	return Decimal{i: z, scale: scale}
}

func NewFromInt64(ctx Context, n int64) Decimal {
	d := NewFromScaledInt(big.NewInt(n), 0)
	return ctx.Normalize(d)
}

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

func Parse(ctx Context, s string) (Decimal, error) {
	d, err := ParseExact(s)
	if err != nil {
		return Decimal{}, err
	}
	return ctx.Normalize(d), nil
}

func MustParse(ctx Context, s string) Decimal {
	d, err := Parse(ctx, s)
	if err != nil {
		panic(err)
	}
	return d
}

func Neg(d Decimal) Decimal {
	var z big.Int
	z.Neg(&d.i)
	return Decimal{i: z, scale: d.scale}
}

func (d Decimal) Scale() int32 {
	return d.scale
}

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
