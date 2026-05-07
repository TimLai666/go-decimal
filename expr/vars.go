// Package expr is a small compile-once / evaluate-many expression engine
// over decimal.Decimal values.
//
// Compile turns a textual expression into a Program (an RPN bytecode
// produced by a shunting-yard parser); Program.Eval then runs that
// program against a decimal.Context and a Vars binding. The split lets
// callers pay the parsing cost once and reuse the compiled program in
// hot loops.
//
// Supported syntax:
//   - Numeric literals (decimal point optional, no scientific notation)
//   - Identifiers as variable names (look up via the Vars argument)
//   - Binary operators: +  -  *  /  ^
//   - Unary +/-, parentheses for grouping
//
// The ^ operator is right-associative and binds tighter than * and /;
// 2^3^2 evaluates to 2^(3^2) = 512, and -2^2 evaluates to -(2^2) = -4.
package expr

import "github.com/TimLai666/go-decimal/decimal"

// Vars resolves variable names to decimal values during Eval.
//
// Implementations should return ok = false when the name is unknown so
// Eval can surface ErrUnknownVar with the offending name attached.
type Vars interface {
	// Get looks up a single variable by name. It returns the value and
	// true on hit, or the zero Decimal and false on miss.
	Get(name string) (decimal.Decimal, bool)
}

// MapVars is a Vars implementation backed by an ordinary map. It is the
// fastest way to plug a fixed set of bindings into a Program.
type MapVars map[string]decimal.Decimal

// Get satisfies Vars by looking name up in the underlying map. Missing
// keys produce (zero Decimal, false).
func (m MapVars) Get(name string) (decimal.Decimal, bool) {
	v, ok := m[name]
	return v, ok
}
