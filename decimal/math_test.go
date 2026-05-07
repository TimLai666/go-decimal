package decimal

import "testing"

func TestSqrt(t *testing.T) {
	ctx := Context{Scale: 10, Mode: RoundingModeHalfUp}

	cases := []struct {
		in   string
		want string
	}{
		{"4", "2.0000000000"},
		{"0", "0.0000000000"},
		{"2", "1.4142135624"},
		{"100", "10.0000000000"},
		{"0.25", "0.5000000000"},
	}

	for _, tc := range cases {
		x := mustParseExact(t, tc.in)
		got, err := Sqrt(ctx, x)
		if err != nil {
			t.Fatalf("Sqrt(%s) error: %v", tc.in, err)
		}
		if got.String() != tc.want {
			t.Fatalf("Sqrt(%s) = %s, want %s", tc.in, got.String(), tc.want)
		}
	}

	if _, err := Sqrt(ctx, mustParseExact(t, "-1")); err == nil {
		t.Fatalf("Sqrt(-1) expected error")
	}
}

func TestExp(t *testing.T) {
	ctx := Context{Scale: 10, Mode: RoundingModeHalfUp}

	cases := []struct {
		in   string
		want string
	}{
		{"0", "1.0000000000"},
		{"1", "2.7182818285"},
		{"2", "7.3890560989"},
		{"-1", "0.3678794412"},
		{"0.5", "1.6487212707"},
	}

	for _, tc := range cases {
		x := mustParseExact(t, tc.in)
		got := Exp(ctx, x)
		if got.String() != tc.want {
			t.Fatalf("Exp(%s) = %s, want %s", tc.in, got.String(), tc.want)
		}
	}
}

func TestLog(t *testing.T) {
	ctx := Context{Scale: 10, Mode: RoundingModeHalfUp}

	cases := []struct {
		in   string
		want string
	}{
		{"1", "0.0000000000"},
		{"2", "0.6931471806"},
		{"10", "2.3025850930"},
		{"0.5", "-0.6931471806"},
	}

	for _, tc := range cases {
		x := mustParseExact(t, tc.in)
		got, err := Log(ctx, x)
		if err != nil {
			t.Fatalf("Log(%s) error: %v", tc.in, err)
		}
		if got.String() != tc.want {
			t.Fatalf("Log(%s) = %s, want %s", tc.in, got.String(), tc.want)
		}
	}

	if _, err := Log(ctx, mustParseExact(t, "0")); err == nil {
		t.Fatalf("Log(0) expected error")
	}
	if _, err := Log(ctx, mustParseExact(t, "-1")); err == nil {
		t.Fatalf("Log(-1) expected error")
	}
}

func TestLogExpRoundTrip(t *testing.T) {
	ctx := Context{Scale: 20, Mode: RoundingModeHalfUp}
	for _, s := range []string{"1.5", "3.7", "0.123", "42.5"} {
		x := mustParseExact(t, s)
		l, err := Log(ctx, x)
		if err != nil {
			t.Fatalf("Log(%s) error: %v", s, err)
		}
		back := Exp(ctx, l)
		// Round to 10 places for comparison; the last few guard digits may drift.
		out := Context{Scale: 10, Mode: RoundingModeHalfUp}
		expected := out.Normalize(x).String()
		got := out.Normalize(back).String()
		if got != expected {
			t.Fatalf("Exp(Log(%s)) = %s, want %s", s, got, expected)
		}
	}
}

func TestPowInt(t *testing.T) {
	ctx := Context{Scale: 6, Mode: RoundingModeHalfUp}

	cases := []struct {
		base, exp, want string
	}{
		{"2", "10", "1024.000000"},
		{"3", "0", "1.000000"},
		{"2", "-3", "0.125000"},
		{"-2", "3", "-8.000000"},
		{"-2", "4", "16.000000"},
		{"0", "5", "0.000000"},
		{"0", "0", "1.000000"},
		{"1.5", "2", "2.250000"},
	}

	for _, tc := range cases {
		got, err := Pow(ctx, mustParseExact(t, tc.base), mustParseExact(t, tc.exp))
		if err != nil {
			t.Fatalf("Pow(%s, %s) error: %v", tc.base, tc.exp, err)
		}
		if got.String() != tc.want {
			t.Fatalf("Pow(%s, %s) = %s, want %s", tc.base, tc.exp, got.String(), tc.want)
		}
	}

	if _, err := Pow(ctx, mustParseExact(t, "0"), mustParseExact(t, "-1")); err == nil {
		t.Fatalf("Pow(0, -1) expected error")
	}
}

func TestPowFractional(t *testing.T) {
	ctx := Context{Scale: 10, Mode: RoundingModeHalfUp}

	got, err := Pow(ctx, mustParseExact(t, "2"), mustParseExact(t, "0.5"))
	if err != nil {
		t.Fatalf("Pow(2, 0.5) error: %v", err)
	}
	if got.String() != "1.4142135624" {
		t.Fatalf("Pow(2, 0.5) = %s, want 1.4142135624", got.String())
	}

	got, err = Pow(ctx, mustParseExact(t, "9"), mustParseExact(t, "0.5"))
	if err != nil {
		t.Fatalf("Pow(9, 0.5) error: %v", err)
	}
	if got.String() != "3.0000000000" {
		t.Fatalf("Pow(9, 0.5) = %s, want 3.0000000000", got.String())
	}

	if _, err := Pow(ctx, mustParseExact(t, "-2"), mustParseExact(t, "0.5")); err == nil {
		t.Fatalf("Pow(-2, 0.5) expected error")
	}
}

func TestCmp(t *testing.T) {
	a := mustParseExact(t, "1.5")
	b := mustParseExact(t, "1.50")
	if Cmp(a, b) != 0 {
		t.Fatalf("Cmp(1.5, 1.50) != 0")
	}
	if Cmp(mustParseExact(t, "1.5"), mustParseExact(t, "1.4")) <= 0 {
		t.Fatalf("Cmp(1.5, 1.4) should be > 0")
	}
	if Cmp(mustParseExact(t, "-1.5"), mustParseExact(t, "1.4")) >= 0 {
		t.Fatalf("Cmp(-1.5, 1.4) should be < 0")
	}
}

