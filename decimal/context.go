package decimal

// RoundingMode selects the rounding strategy used by Context.Normalize and
// every operation that produces results at a target scale (Add, Sub, Mul,
// Div, Sqrt, Exp, Log, Pow).
type RoundingMode uint8

const (
	// RoundingModeDown truncates toward zero (the absolute value never grows).
	// 1.235 and -1.235 both become 1.23 / -1.23 at Scale = 2.
	RoundingModeDown RoundingMode = iota

	// RoundingModeUp rounds away from zero whenever any non-zero residue
	// remains. 1.231 and -1.231 both become 1.24 / -1.24 at Scale = 2.
	RoundingModeUp

	// RoundingModeHalfUp is the everyday "round half away from zero" rule.
	// Exact halves move away from zero: 1.235 → 1.24, -1.235 → -1.24.
	// Matches Java BigDecimal.ROUND_HALF_UP and Python decimal.ROUND_HALF_UP.
	RoundingModeHalfUp

	// RoundingModeHalfEven is banker's rounding: the IEEE 754 default and
	// Python decimal's default. Behaves like HalfUp except that exact halves
	// pick the neighbour whose last digit is even, eliminating the systematic
	// upward bias of HalfUp on long sums. 1.225 → 1.22, 1.235 → 1.24,
	// -1.225 → -1.22.
	RoundingModeHalfEven

	// RoundingModeCeiling rounds toward +∞ whenever any non-zero residue
	// remains. Positive values round away from zero; negative values round
	// toward zero. 1.231 → 1.24, -1.231 → -1.23 at Scale = 2.
	RoundingModeCeiling

	// RoundingModeFloor rounds toward -∞ whenever any non-zero residue
	// remains. Positive values round toward zero; negative values round
	// away from zero. 1.231 → 1.23, -1.231 → -1.24 at Scale = 2.
	RoundingModeFloor

	// RoundingModeHalfDown is "round half toward zero": behaves like HalfUp
	// except that exact halves stay put rather than stepping away. 1.235 →
	// 1.23, 1.236 → 1.24, -1.235 → -1.23. Matches Java BigDecimal.ROUND_HALF_DOWN
	// and Python decimal.ROUND_HALF_DOWN.
	RoundingModeHalfDown

	// RoundingMode05Up implements Python's decimal.ROUND_05UP rule: after
	// truncating toward zero, if the kept last digit is 0 or 5 then any
	// non-zero residue causes a step away from zero; otherwise the residue
	// is dropped. Used in some accounting contexts to avoid producing 5s
	// as final digits unless they are exact.
	RoundingMode05Up

	// RoundingModeUnnecessary asserts that no rounding will be required.
	// If an operation under this mode has to discard a non-zero residue,
	// it panics with ErrRoundingNecessary. Equivalent to Java's
	// RoundingMode.UNNECESSARY (which throws ArithmeticException).
	RoundingModeUnnecessary
)

// Context bundles the target precision and rounding policy used by every
// Context-aware operation in this package.
//
// Scale is the number of fractional digits in the output (so 10^-Scale is
// the smallest representable unit). Mode chooses how excess digits are
// dropped when an operation would otherwise produce more.
//
// Context is plain data: copy it freely, share it across goroutines, and
// reuse it in concurrent calls.
type Context struct {
	// Scale is the number of fractional digits in normalized results.
	Scale int32
	// Mode selects the rounding strategy; see RoundingMode.
	Mode RoundingMode
}

// Normalize re-scales d to c.Scale: it pads with zeros when the input has
// fewer digits, and rounds according to c.Mode when it has more.
//
// When d already matches c.Scale the original value is returned without
// allocating a new big.Int.
func (c Context) Normalize(d Decimal) Decimal {
	return normalize(d, c.Scale, c.Mode)
}
