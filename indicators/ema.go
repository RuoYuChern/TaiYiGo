package indicators

type emaForward struct {
	ts     *TimeSeries
	opt    GetValue
	window int
	alpha  Decimal
	beta   Decimal
	preEma Decimal
}

type emaBack struct {
	ts     *TimeSeries
	opt    GetValue
	window int
	alpha  Decimal
	beta   Decimal
	cache  map[int]Decimal
}

func NewEMAIndicator(ts *TimeSeries, opt GetValue, window int) Indicator {
	m1 := ONE.Frac(2).Div(NewFromInt(window + 1))
	ind := &emaForward{
		ts:     ts,
		opt:    opt,
		window: window,
		alpha:  m1,
		preEma: ZERO,
	}
	ind.beta = ONE.Sub(ind.alpha)
	return ind
}

func NewEMAIndicator2(ts *TimeSeries, opt GetValue, window int) Indicator {
	m1 := ONE.Frac(2).Div(NewFromInt(window + 1))
	ind := &emaBack{
		ts:     ts,
		opt:    opt,
		window: window,
		alpha:  m1,
		cache:  make(map[int]Decimal),
	}
	ind.beta = ONE.Sub(ind.alpha)
	return ind
}

func (ema *emaForward) Calculate(index int) Decimal {
	if index < (ema.window - 1) {
		return ZERO
	}
	if index == (ema.window - 1) {
		sam := NewSimpleMovingAverage(ema.ts, ema.opt, ema.window)
		ema.preEma = sam.Calculate(index)
		return ema.preEma
	}

	todayVal := ema.opt(ema.ts.Get(index)).Mul(ema.alpha)
	preValue := ema.preEma.Mul(ema.beta)
	emaValue := todayVal.Add(preValue)
	ema.preEma = emaValue
	return emaValue
}

func (ema *emaBack) Calculate(index int) Decimal {
	value, ok := ema.cache[index]
	if ok {
		return value
	}
	if index < (ema.window - 1) {
		return ZERO
	}
	if index == (ema.window - 1) {
		sam := NewSimpleMovingAverage(ema.ts, ema.opt, ema.window)
		value = sam.Calculate(index)
		ema.cache[index] = value
		return value
	}

	todayVal := ema.opt(ema.ts.Get(index)).Mul(ema.alpha)
	preValue := ema.Calculate(index - 1).Mul(ema.beta)
	value = todayVal.Add(preValue)
	ema.cache[index] = value
	return value
}
