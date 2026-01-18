package decimal

import "errors"

var (
	ErrInvalidDecimal = errors.New("invalid decimal")
	ErrDivisionByZero = errors.New("division by zero")
)
