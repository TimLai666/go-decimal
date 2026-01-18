package decimal

type RoundingMode uint8

const (
	RoundDown   RoundingMode = iota // toward zero
	RoundUp                         // away from zero
	RoundHalfUp                     // halves away from zero
)

type Context struct {
	Scale int32
	Mode  RoundingMode
}

func (c Context) Normalize(d Decimal) Decimal {
	return normalize(d, c.Scale, c.Mode)
}
