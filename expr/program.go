package expr

import (
	"fmt"

	"github.com/TimLai666/go-decimal/decimal"
)

type opCode uint8

const (
	opPushConst opCode = iota
	opPushVar
	opAdd
	opSub
	opMul
	opDiv
	opPow
	opNeg
)

type op struct {
	code opCode
	arg  int
}

// Program is a compiled expression ready to be evaluated.
//
// Internally it holds an RPN bytecode plus its constant pool and the
// list of variable names referenced by the source. A Program is safe to
// reuse and to share across goroutines as long as no one mutates it
// (which the public API does not allow). Compile to obtain one.
type Program struct {
	ops    []op
	consts []decimal.Decimal
	vars   []string
}

// Eval runs the program against ctx and the provided Vars binding,
// returning the resulting decimal value.
//
// All arithmetic is performed through the decimal package, so every
// intermediate result is normalized to ctx.Scale and ctx.Mode just as
// if the caller had used Add / Sub / Mul / Div / Pow directly.
//
// Possible errors:
//   - ErrInvalidExpr   if the program is nil or its bytecode is malformed.
//     A well-formed Compile output should never trigger this at Eval time.
//   - ErrUnknownVar    if vars cannot resolve a referenced name. The
//     wrapped error includes the offending name.
//   - decimal.ErrDivisionByZero  on division by zero.
//   - decimal.ErrInvalidPow      from "^" with a negative base and a
//     non-integer exponent.
//   - decimal.ErrNonPositiveLog  surfaced through "^" when computing
//     ln(base) for non-positive base under a fractional exponent.
//
// A nil Vars is acceptable as long as the program references no variables.
func (p *Program) Eval(ctx decimal.Context, vars Vars) (decimal.Decimal, error) {
	if p == nil {
		return decimal.Decimal{}, ErrInvalidExpr
	}

	stack := make([]decimal.Decimal, 0, len(p.ops))

	for _, inst := range p.ops {
		switch inst.code {
		case opPushConst:
			if inst.arg < 0 || inst.arg >= len(p.consts) {
				return decimal.Decimal{}, ErrInvalidExpr
			}
			stack = append(stack, p.consts[inst.arg])
		case opPushVar:
			if inst.arg < 0 || inst.arg >= len(p.vars) {
				return decimal.Decimal{}, ErrInvalidExpr
			}
			if vars == nil {
				return decimal.Decimal{}, ErrUnknownVar
			}
			name := p.vars[inst.arg]
			value, ok := vars.Get(name)
			if !ok {
				return decimal.Decimal{}, fmt.Errorf("%w: %s", ErrUnknownVar, name)
			}
			stack = append(stack, value)
		case opNeg:
			if len(stack) < 1 {
				return decimal.Decimal{}, ErrInvalidExpr
			}
			v := stack[len(stack)-1]
			stack[len(stack)-1] = decimal.Neg(v)
		case opAdd, opSub, opMul, opDiv, opPow:
			if len(stack) < 2 {
				return decimal.Decimal{}, ErrInvalidExpr
			}
			b := stack[len(stack)-1]
			a := stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			var res decimal.Decimal
			var err error
			switch inst.code {
			case opAdd:
				res = decimal.Add(ctx, a, b)
			case opSub:
				res = decimal.Sub(ctx, a, b)
			case opMul:
				res = decimal.Mul(ctx, a, b)
			case opDiv:
				res, err = decimal.Div(ctx, a, b)
				if err != nil {
					return decimal.Decimal{}, err
				}
			case opPow:
				res, err = decimal.Pow(ctx, a, b)
				if err != nil {
					return decimal.Decimal{}, err
				}
			}
			stack = append(stack, res)
		default:
			return decimal.Decimal{}, ErrInvalidExpr
		}
	}

	if len(stack) != 1 {
		return decimal.Decimal{}, ErrInvalidExpr
	}

	return stack[0], nil
}
