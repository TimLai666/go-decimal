package expr

import (
	"testing"

	"github.com/TimLai666/go-decimal/decimal"
)

func TestCompileEval(t *testing.T) {
	ctx := decimal.Context{Scale: 2, Mode: decimal.RoundingModeHalfUp}

	prog, err := Compile("1.2 + x/3")
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	vars := MapVars{
		"x": decimal.MustParse(ctx, "10"),
	}

	res, err := prog.Eval(ctx, vars)
	if err != nil {
		t.Fatalf("Eval error: %v", err)
	}
	if res.String() != "4.53" {
		t.Fatalf("Eval = %q, want 4.53", res.String())
	}
}

func TestUnaryAndParens(t *testing.T) {
	ctx := decimal.Context{Scale: 2, Mode: decimal.RoundingModeHalfUp}

	prog, err := Compile("-1 + 2")
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}
	res, err := prog.Eval(ctx, MapVars{})
	if err != nil {
		t.Fatalf("Eval error: %v", err)
	}
	if res.String() != "1.00" {
		t.Fatalf("Eval = %q, want 1.00", res.String())
	}

	prog, err = Compile("1 + 2*(3-1)")
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}
	res, err = prog.Eval(ctx, MapVars{})
	if err != nil {
		t.Fatalf("Eval error: %v", err)
	}
	if res.String() != "5.00" {
		t.Fatalf("Eval = %q, want 5.00", res.String())
	}
}

func TestUnknownVar(t *testing.T) {
	ctx := decimal.Context{Scale: 2, Mode: decimal.RoundingModeHalfUp}

	prog, err := Compile("x + 1")
	if err != nil {
		t.Fatalf("Compile error: %v", err)
	}

	if _, err := prog.Eval(ctx, MapVars{}); err == nil {
		t.Fatalf("expected unknown var error")
	}
}

func TestCompileInvalid(t *testing.T) {
	if _, err := Compile("1 +"); err == nil {
		t.Fatalf("expected compile error")
	}
}

func TestPowOperator(t *testing.T) {
	ctx := decimal.Context{Scale: 6, Mode: decimal.RoundingModeHalfUp}

	cases := []struct {
		expr string
		want string
	}{
		{"2^10", "1024.000000"},
		{"2^3^2", "512.000000"},   // right-associative: 2^(3^2)=2^9
		{"-2^2", "-4.000000"},     // unary minus binds looser than ^
		{"(-2)^2", "4.000000"},    // parenthesised negation
		{"2 + 3^2", "11.000000"},  // ^ binds tighter than +
		{"2 * 3^2", "18.000000"},  // ^ binds tighter than *
		{"4^0.5", "2.000000"},     // fractional exponent → sqrt
	}

	for _, tc := range cases {
		prog, err := Compile(tc.expr)
		if err != nil {
			t.Fatalf("Compile(%q) error: %v", tc.expr, err)
		}
		res, err := prog.Eval(ctx, MapVars{})
		if err != nil {
			t.Fatalf("Eval(%q) error: %v", tc.expr, err)
		}
		if res.String() != tc.want {
			t.Fatalf("Eval(%q) = %s, want %s", tc.expr, res.String(), tc.want)
		}
	}
}
