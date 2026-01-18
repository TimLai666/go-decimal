package bench

import (
	"testing"

	"github.com/TimLai666/go-decimal/decimal"
)

func BenchmarkAdd(b *testing.B) {
	ctx := decimal.Context{Scale: 2, Mode: decimal.RoundingModeHalfUp}
	a := decimal.MustParse(ctx, "12345.67")
	c := decimal.MustParse(ctx, "890.12")

	var res decimal.Decimal
	for i := 0; i < b.N; i++ {
		res = decimal.Add(ctx, a, c)
	}
	_ = res
}

func BenchmarkMul(b *testing.B) {
	ctx := decimal.Context{Scale: 2, Mode: decimal.RoundingModeHalfUp}
	a := decimal.MustParse(ctx, "123.45")
	c := decimal.MustParse(ctx, "6.78")

	var res decimal.Decimal
	for i := 0; i < b.N; i++ {
		res = decimal.Mul(ctx, a, c)
	}
	_ = res
}
