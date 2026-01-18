package expr

import "github.com/TimLai666/go-decimal/decimal"

type Vars interface {
	Get(name string) (decimal.Decimal, bool)
}

type MapVars map[string]decimal.Decimal

func (m MapVars) Get(name string) (decimal.Decimal, bool) {
	v, ok := m[name]
	return v, ok
}
