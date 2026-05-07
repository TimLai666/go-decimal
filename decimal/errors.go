package decimal

import "errors"

var (
	// ErrInvalidDecimal is returned by ParseExact and Parse when the input
	// is not a well-formed decimal literal (empty, illegal characters, or
	// more than one decimal point).
	ErrInvalidDecimal = errors.New("invalid decimal")

	// ErrDivisionByZero is returned by Div, and by Pow when raising 0 to a
	// negative integer power (which would require dividing by zero).
	ErrDivisionByZero = errors.New("division by zero")

	// ErrNegativeSqrt is returned by Sqrt when the input is negative.
	ErrNegativeSqrt = errors.New("square root of negative number")

	// ErrNonPositiveLog is returned by Log when the input is zero or
	// negative — neither is in the domain of the real-valued logarithm.
	ErrNonPositiveLog = errors.New("logarithm of non-positive number")

	// ErrInvalidPow is returned by Pow when the base is negative and the
	// exponent is not an integer (the result would be complex-valued).
	ErrInvalidPow = errors.New("invalid power: negative base with non-integer exponent")
)
