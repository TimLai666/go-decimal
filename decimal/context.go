package decimal

type RoundingMode uint8

const (
	RoundingModeDown   RoundingMode = iota // toward zero
	RoundingModeUp                         // away from zero
	RoundingModeHalfUp                     // halves away from zero
)

type Context struct {
	Scale int32
	Mode  RoundingMode
}

func (c Context) Normalize(d Decimal) Decimal {
	return normalize(d, c.Scale, c.Mode)
}
