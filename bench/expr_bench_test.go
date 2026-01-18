package bench

import (
	"testing"

	"github.com/TimLai666/go-decimal/decimal"
	"github.com/TimLai666/go-decimal/expr"
)

func BenchmarkCompile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if _, err := expr.Compile("1.2 + x/3 - y*2"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEval(b *testing.B) {
	ctx := decimal.Context{Scale: 2, Mode: decimal.RoundingModeHalfUp}
	prog, err := expr.Compile("1.2 + x/3 - y*2")
	if err != nil {
		b.Fatal(err)
	}

	vars := expr.MapVars{
		"x": decimal.MustParse(ctx, "10"),
		"y": decimal.MustParse(ctx, "2"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := prog.Eval(ctx, vars); err != nil {
			b.Fatal(err)
		}
	}
}
