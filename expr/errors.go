package expr

import "errors"

var (
	ErrInvalidExpr = errors.New("invalid expression")
	ErrUnknownVar  = errors.New("unknown variable")
)
