package expr

import "errors"

var (
	// ErrInvalidExpr is returned when parsing fails.
	//
	// Common triggers: empty input, unbalanced parentheses, misplaced
	// operators, malformed numeric literals, and unfinished expressions
	// (e.g. a trailing operator with no operand).
	ErrInvalidExpr = errors.New("invalid expression")

	// ErrUnknownVar is returned by Eval when the program references a
	// variable name that the supplied Vars cannot resolve.
	//
	// The actual error wraps the variable name in "%w: name", so use
	// errors.Is to compare against this sentinel.
	ErrUnknownVar = errors.New("unknown variable")
)
