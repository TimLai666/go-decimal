package expr

import (
	"fmt"

	"github.com/TimLai666/go-decimal/decimal"
)

type opInfo struct {
	code       opCode
	precedence int
	rightAssoc bool
}

type opStackItem struct {
	op      opInfo
	isParen bool
}

func Compile(input string) (*Program, error) {
	tokens, err := tokenize(input)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, ErrInvalidExpr
	}

	var prog Program
	prog.ops = make([]op, 0, len(tokens))
	prog.consts = make([]decimal.Decimal, 0, len(tokens))
	prog.vars = make([]string, 0)

	var opStack []opStackItem
	prevWasValue := false

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]
		switch tok.typ {
		case tokenNumber:
			if prevWasValue {
				return nil, ErrInvalidExpr
			}
			idx := len(prog.consts)
			prog.consts = append(prog.consts, tok.value)
			prog.ops = append(prog.ops, op{code: opPushConst, arg: idx})
			prevWasValue = true
		case tokenIdent:
			if prevWasValue {
				return nil, ErrInvalidExpr
			}
			idx := prog.addVar(tok.text)
			prog.ops = append(prog.ops, op{code: opPushVar, arg: idx})
			prevWasValue = true
		case tokenOperator:
			if !prevWasValue {
				if tok.op == '+' {
					continue
				}
				if tok.op == '-' {
					if err := pushOperator(&prog, &opStack, opInfo{code: opNeg, precedence: 3, rightAssoc: true}); err != nil {
						return nil, err
					}
					prevWasValue = false
					continue
				}
				return nil, ErrInvalidExpr
			}

			info, ok := binaryOpInfo(tok.op)
			if !ok {
				return nil, ErrInvalidExpr
			}
			if err := pushOperator(&prog, &opStack, info); err != nil {
				return nil, err
			}
			prevWasValue = false
		case tokenLParen:
			if prevWasValue {
				return nil, ErrInvalidExpr
			}
			opStack = append(opStack, opStackItem{isParen: true})
			prevWasValue = false
		case tokenRParen:
			if !prevWasValue {
				return nil, ErrInvalidExpr
			}
			found := false
			for len(opStack) > 0 {
				top := opStack[len(opStack)-1]
				opStack = opStack[:len(opStack)-1]
				if top.isParen {
					found = true
					break
				}
				prog.ops = append(prog.ops, op{code: top.op.code})
			}
			if !found {
				return nil, ErrInvalidExpr
			}
			prevWasValue = true
		}
	}

	if !prevWasValue {
		return nil, ErrInvalidExpr
	}

	for len(opStack) > 0 {
		top := opStack[len(opStack)-1]
		opStack = opStack[:len(opStack)-1]
		if top.isParen {
			return nil, ErrInvalidExpr
		}
		prog.ops = append(prog.ops, op{code: top.op.code})
	}

	return &prog, nil
}

func pushOperator(prog *Program, stack *[]opStackItem, info opInfo) error {
	for len(*stack) > 0 {
		top := (*stack)[len(*stack)-1]
		if top.isParen {
			break
		}
		if info.rightAssoc {
			if info.precedence < top.op.precedence {
				*stack = (*stack)[:len(*stack)-1]
				prog.ops = append(prog.ops, op{code: top.op.code})
				continue
			}
		} else {
			if info.precedence <= top.op.precedence {
				*stack = (*stack)[:len(*stack)-1]
				prog.ops = append(prog.ops, op{code: top.op.code})
				continue
			}
		}
		break
	}

	*stack = append(*stack, opStackItem{op: info})
	return nil
}

func binaryOpInfo(op byte) (opInfo, bool) {
	switch op {
	case '+':
		return opInfo{code: opAdd, precedence: 1}, true
	case '-':
		return opInfo{code: opSub, precedence: 1}, true
	case '*':
		return opInfo{code: opMul, precedence: 2}, true
	case '/':
		return opInfo{code: opDiv, precedence: 2}, true
	default:
		return opInfo{}, false
	}
}

func (p *Program) addVar(name string) int {
	for i, existing := range p.vars {
		if existing == name {
			return i
		}
	}
	p.vars = append(p.vars, name)
	return len(p.vars) - 1
}

func (p *Program) String() string {
	return fmt.Sprintf("Program{ops:%d,consts:%d,vars:%d}", len(p.ops), len(p.consts), len(p.vars))
}
