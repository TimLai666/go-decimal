package expr

import (
	"fmt"

	"github.com/tingzhen/go-decimal/decimal"
)

type opCode uint8

const (
	opPushConst opCode = iota
	opPushVar
	opAdd
	opSub
	opMul
	opDiv
	opNeg
)

type op struct {
	code opCode
	arg  int
}

type Program struct {
	ops    []op
	consts []decimal.Decimal
	vars   []string
}

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
		case opAdd, opSub, opMul, opDiv:
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
