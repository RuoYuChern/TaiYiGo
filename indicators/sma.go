package indicators

type smaForward struct {
	ts     *TimeSeries
	opt    GetValue
	window int
	preSma Decimal
}

type smaBack struct {
	ts     *TimeSeries
	opt    GetValue
	window int
}

func NewSimpleMovingAverage(ts *TimeSeries, opt GetValue, window int) Indicator {
	return &smaForward{ts: ts, opt: opt, window: window, preSma: ZERO}
}

func NewSimpleMovingAverage2(ts *TimeSeries, opt GetValue, window int) Indicator {
	return &smaBack{ts: ts, opt: opt, window: window}
}

func (sma *smaForward) Calculate(index int) Decimal {
	if index < (sma.window - 1) {
		return ZERO
	}
	if index == (sma.window - 1) {
		sum := ZERO
		start := index - sma.window + 1
		for i := start; i <= index; i++ {
			d := sma.opt(sma.ts.Get(i))
			sum = sum.Add(d)
		}
		sma.preSma = sum.Div(NewFromInt(sma.window))

	} else {
		lastStart := sma.opt(sma.ts.Get(index - sma.window))
		diff := sma.opt(sma.ts.Get(index)).Sub(lastStart)
		sma.preSma = sma.preSma.Add(diff.Div(NewFromInt(sma.window)))
	}
	return sma.preSma
}

func (sma *smaBack) Calculate(index int) Decimal {
	if index < (sma.window - 1) {
		return ZERO
	}

	sum := ZERO
	start := index - sma.window + 1
	for i := start; i <= index; i++ {
		d := sma.opt(sma.ts.Get(i))
		sum = sum.Add(d)
	}
	return sum.Div(NewFromInt(sma.window))
}
