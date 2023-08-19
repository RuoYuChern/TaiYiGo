package indicators

type mtnIndicator struct {
	ts     *TimeSeries
	opt    GetValue
	window int
}

func NewMtn(ts *TimeSeries, opt GetValue, window int) Indicator {
	return &mtnIndicator{ts: ts, opt: opt, window: window}
}

func (mnt *mtnIndicator) Calculate(index int) Decimal {
	if index < mnt.window {
		return ZERO
	}
	vnow := mnt.opt(mnt.ts.Get(index))
	vbefor := mnt.opt(mnt.ts.Get(index - mnt.window))
	diff := vnow.Sub(vbefor).Div(NewDecimal(float64(mnt.window)))
	return diff
}
