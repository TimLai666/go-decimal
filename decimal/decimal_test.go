package decimal

import "testing"

func mustParseExact(t *testing.T, s string) Decimal {
	t.Helper()
	d, err := ParseExact(s)
	if err != nil {
		t.Fatalf("ParseExact(%q) error: %v", s, err)
	}
	return d
}

func TestParseExactString(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"123.45", "123.45"},
		{".5", "0.5"},
		{"-0.10", "-0.10"},
		{"1.", "1"},
	}

	for _, tc := range cases {
		got := mustParseExact(t, tc.in).String()
		if got != tc.want {
			t.Fatalf("ParseExact(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestParseContextRounding(t *testing.T) {
	ctx := Context{Scale: 2, Mode: RoundingModeHalfUp}
	got := MustParse(ctx, "1.235").String()
	if got != "1.24" {
		t.Fatalf("rounded = %q, want 1.24", got)
	}

	ctx = Context{Scale: 2, Mode: RoundingModeUp}
	got = MustParse(ctx, "-1.231").String()
	if got != "-1.24" {
		t.Fatalf("round up = %q, want -1.24", got)
	}
}

func TestArith(t *testing.T) {
	ctx := Context{Scale: 2, Mode: RoundingModeHalfUp}

	a := mustParseExact(t, "1.20")
	b := mustParseExact(t, "1.234")
	sum := Add(ctx, a, b).String()
	if sum != "2.43" {
		t.Fatalf("Add = %q, want 2.43", sum)
	}

	prod := Mul(ctx, mustParseExact(t, "1.20"), mustParseExact(t, "2.00")).String()
	if prod != "2.40" {
		t.Fatalf("Mul = %q, want 2.40", prod)
	}

	q, err := Div(ctx, mustParseExact(t, "1.234"), mustParseExact(t, "2"))
	if err != nil {
		t.Fatalf("Div error: %v", err)
	}
	if q.String() != "0.62" {
		t.Fatalf("Div = %q, want 0.62", q.String())
	}

	if _, err := Div(ctx, mustParseExact(t, "1"), mustParseExact(t, "0")); err == nil {
		t.Fatalf("Div by zero expected error")
	}
}
