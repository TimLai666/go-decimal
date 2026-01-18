package expr

import (
	"testing"

	"github.com/TimLai666/go-decimal/decimal"
)

func TestCompileEval(t *testing.T) {
	ctx := decimal.Context{Scale: 2, Mode: decimal.RoundHalfUp}

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
	ctx := decimal.Context{Scale: 2, Mode: decimal.RoundHalfUp}

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
	ctx := decimal.Context{Scale: 2, Mode: decimal.RoundHalfUp}

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
