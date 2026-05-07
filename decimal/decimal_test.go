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

func TestRoundingModes(t *testing.T) {
	cases := []struct {
		in    string
		scale int32
		mode  RoundingMode
		want  string
	}{
		// Exact half, positive — drives every halving rule.
		{"1.25", 1, RoundingModeDown, "1.2"},
		{"1.25", 1, RoundingModeUp, "1.3"},
		{"1.25", 1, RoundingModeHalfUp, "1.3"},
		{"1.25", 1, RoundingModeHalfEven, "1.2"}, // 2 already even
		{"1.35", 1, RoundingModeHalfEven, "1.4"}, // 3 odd → step
		{"2.5", 0, RoundingModeHalfEven, "2"},
		{"3.5", 0, RoundingModeHalfEven, "4"},
		{"1.25", 1, RoundingModeCeiling, "1.3"},
		{"1.25", 1, RoundingModeFloor, "1.2"},

		// Exact half, negative — Ceiling/Floor diverge from Up/HalfUp here.
		{"-1.25", 1, RoundingModeDown, "-1.2"},
		{"-1.25", 1, RoundingModeUp, "-1.3"},
		{"-1.25", 1, RoundingModeHalfUp, "-1.3"},
		{"-1.25", 1, RoundingModeHalfEven, "-1.2"},
		{"-1.35", 1, RoundingModeHalfEven, "-1.4"},
		{"-2.5", 0, RoundingModeHalfEven, "-2"},
		{"-1.25", 1, RoundingModeCeiling, "-1.2"}, // toward +∞
		{"-1.25", 1, RoundingModeFloor, "-1.3"},

		// Non-half residues — Ceiling/Floor step on any non-zero residue.
		{"1.21", 1, RoundingModeCeiling, "1.3"},
		{"1.29", 1, RoundingModeFloor, "1.2"},
		{"-1.21", 1, RoundingModeFloor, "-1.3"},
		{"-1.29", 1, RoundingModeCeiling, "-1.2"},

		// Sub-unit: quo collapses to 0 but the original sign still matters.
		{"-0.001", 0, RoundingModeFloor, "-1"},
		{"0.001", 0, RoundingModeCeiling, "1"},
		{"-0.001", 0, RoundingModeCeiling, "0"},
		{"0.001", 0, RoundingModeFloor, "0"},
	}

	for _, tc := range cases {
		ctx := Context{Scale: tc.scale, Mode: tc.mode}
		got := MustParse(ctx, tc.in).String()
		if got != tc.want {
			t.Errorf("Round(%s, scale=%d, mode=%d) = %q, want %q",
				tc.in, tc.scale, tc.mode, got, tc.want)
		}
	}
}

func TestDivRoundingModes(t *testing.T) {
	cases := []struct {
		num, den string
		scale    int32
		mode     RoundingMode
		want     string
	}{
		// 1/3 = 0.333... — never halfway, all modes pick a side.
		{"1", "3", 2, RoundingModeCeiling, "0.34"},
		{"1", "3", 2, RoundingModeFloor, "0.33"},
		{"-1", "3", 2, RoundingModeCeiling, "-0.33"},
		{"-1", "3", 2, RoundingModeFloor, "-0.34"},

		// Exact halfway, exercises HalfEven.
		{"5", "4", 1, RoundingModeHalfEven, "1.2"},  // 1.25, 2 even → 1.2
		{"15", "4", 1, RoundingModeHalfEven, "3.8"}, // 3.75, 7 odd → 3.8
		{"-5", "4", 1, RoundingModeHalfEven, "-1.2"},
	}

	for _, tc := range cases {
		ctx := Context{Scale: tc.scale, Mode: tc.mode}
		num, den := mustParseExact(t, tc.num), mustParseExact(t, tc.den)
		got, err := Div(ctx, num, den)
		if err != nil {
			t.Fatalf("Div(%s/%s) error: %v", tc.num, tc.den, err)
		}
		if got.String() != tc.want {
			t.Errorf("Div(%s/%s, scale=%d, mode=%d) = %q, want %q",
				tc.num, tc.den, tc.scale, tc.mode, got.String(), tc.want)
		}
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
