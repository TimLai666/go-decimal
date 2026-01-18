package expr

import "github.com/tingzhen/go-decimal/decimal"

type tokenType uint8

const (
	tokenNumber tokenType = iota
	tokenIdent
	tokenOperator
	tokenLParen
	tokenRParen
)

type token struct {
	typ   tokenType
	text  string
	op    byte
	value decimal.Decimal
}

func tokenize(input string) ([]token, error) {
	tokens := make([]token, 0, len(input))

	for i := 0; i < len(input); {
		ch := input[i]
		switch {
		case isSpace(ch):
			i++
		case isDigit(ch) || ch == '.':
			start := i
			seenDot := false
			if ch == '.' {
				seenDot = true
				i++
				if i >= len(input) || !isDigit(input[i]) {
					return nil, ErrInvalidExpr
				}
			} else {
				i++
			}

			for i < len(input) {
				ch = input[i]
				if isDigit(ch) {
					i++
					continue
				}
				if ch == '.' {
					if seenDot {
						return nil, ErrInvalidExpr
					}
					seenDot = true
					i++
					continue
				}
				break
			}

			lit := input[start:i]
			dec, err := decimal.ParseExact(lit)
			if err != nil {
				return nil, ErrInvalidExpr
			}
			tokens = append(tokens, token{typ: tokenNumber, value: dec})
		case isIdentStart(ch):
			start := i
			i++
			for i < len(input) && isIdentPart(input[i]) {
				i++
			}
			tokens = append(tokens, token{typ: tokenIdent, text: input[start:i]})
		case ch == '+' || ch == '-' || ch == '*' || ch == '/':
			tokens = append(tokens, token{typ: tokenOperator, op: ch})
			i++
		case ch == '(':
			tokens = append(tokens, token{typ: tokenLParen})
			i++
		case ch == ')':
			tokens = append(tokens, token{typ: tokenRParen})
			i++
		default:
			return nil, ErrInvalidExpr
		}
	}

	return tokens, nil
}

func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isIdentPart(ch byte) bool {
	return isIdentStart(ch) || isDigit(ch)
}
