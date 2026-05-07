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
	RoundingModeHalfUp
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
