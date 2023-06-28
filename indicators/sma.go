package indicators

type smaIndicator struct {
	ts     *TimeSeries
	opt    GetValue
	window int
}

func NewSimpleMovingAverage(ts *TimeSeries, opt GetValue, window int) Indicator {
	return smaIndicator{ts, opt, window}
}

func (sma smaIndicator) Calculate(index int) Decimal {
	if index < (sma.window - 1) {
		return ZERO
	}

	sum := ZERO
	for i := index; i > (index - sma.window); i-- {
		d := sma.opt(sma.ts.Get(i))
		sum = sum.Add(d)
	}
	result := sum.Div(NewFromInt(sma.window))
	return result
}
