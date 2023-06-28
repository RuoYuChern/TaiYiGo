package indicators

type emaIndicator struct {
	ts     *TimeSeries
	opt    GetValue
	window int
	alpha  Decimal
}

func NewEMAIndicator(ts *TimeSeries, opt GetValue, window int) Indicator {
	return &emaIndicator{
		ts:     ts,
		opt:    opt,
		window: window,
		alpha:  ONE.Frac(2).Div(NewFromInt(window + 1)),
	}
}

func (ema *emaIndicator) Calculate(index int) Decimal {
	if index < (ema.window - 1) {
		return ZERO
	}
	emaValue := ZERO
	for start := index - ema.window + 1; start <= index; start++ {
		todayVal := ema.opt(ema.ts.Get(start)).Mul(ema.alpha)
		emaValue = todayVal.Add(emaValue.Mul(ONE.Sub(ema.alpha)))
	}
	return emaValue
}
