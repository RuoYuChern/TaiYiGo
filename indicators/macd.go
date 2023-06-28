package indicators

type differenceIndicator struct {
	shortIndicator Indicator
	longIndeicator Indicator
}

func NewMACDIndicator(ts *TimeSeries, baseIndicator GetValue, shortwindow, longwindow int) Indicator {
	return NewDifferenceIndicator(NewEMAIndicator(ts, baseIndicator, shortwindow), NewEMAIndicator(ts, baseIndicator, longwindow))
}

func NewDifferenceIndicator(shortIndicator, longIndeicator Indicator) Indicator {
	return differenceIndicator{
		shortIndicator: shortIndicator,
		longIndeicator: longIndeicator,
	}
}

func (di differenceIndicator) Calculate(index int) Decimal {
	return di.shortIndicator.Calculate(index).Sub(di.longIndeicator.Calculate(index))
}
