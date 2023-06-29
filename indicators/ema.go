package indicators

type emaIndicator struct {
	ts     *TimeSeries
	opt    GetValue
	window int
	alpha  Decimal
	beta   Decimal
	preEma Decimal
}

func NewEMAIndicator(ts *TimeSeries, opt GetValue, window int) Indicator {
	ind := &emaIndicator{
		ts:     ts,
		opt:    opt,
		window: window,
		alpha:  ONE.Frac(2).Div(NewFromInt(window + 1)),
		preEma: ZERO,
	}
	ind.beta = ONE.Sub(ind.alpha)
	return ind
}

func (ema *emaIndicator) Calculate(index int) Decimal {
	if index < (ema.window - 1) {
		return ZERO
	}
	if index == (ema.window - 1) {
		sam := NewSimpleMovingAverage(ema.ts, ema.opt, ema.window)
		ema.preEma = sam.Calculate(index)
		return ema.preEma
	}

	todayVal := ema.opt(ema.ts.Get(index)).Mul(ema.alpha)
	emaValue := todayVal.Add(ema.preEma.Mul(ema.beta))
	ema.preEma = emaValue
	return emaValue
}
