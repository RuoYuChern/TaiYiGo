package indicators

type wadSmaForward struct {
	ts     []Decimal
	window int
	wady   Decimal
}

type wadSmaBack struct {
	ts     []Decimal
	window int
}

func NewWadSma(ts []Decimal, window int) Indicator {
	return &wadSmaForward{ts: ts, window: window, wady: ZERO}
}

func NewWadSma2(ts []Decimal, window int) Indicator {
	return &wadSmaBack{ts: ts, window: window}
}

func (sma *wadSmaForward) Calculate(index int) Decimal {
	if index < (sma.window - 1) {
		return ZERO
	}
	if index == (sma.window - 1) {
		sum := ZERO
		start := index - sma.window + 1
		for i := start; i <= index; i++ {
			sum = sum.Add(sma.ts[i])
		}
		sma.wady = sum.Div(NewFromInt(sma.window))

	} else {
		lastStart := sma.ts[index-sma.window]
		diff := sma.ts[index].Sub(lastStart)
		sma.wady = sma.wady.Add(diff.Div(NewFromInt(sma.window)))
	}
	return sma.wady
}

func (sma *wadSmaBack) Calculate(index int) Decimal {
	if index < (sma.window - 1) {
		return ZERO
	}

	sum := ZERO
	start := index - sma.window + 1
	for i := start; i <= index; i++ {
		sum = sum.Add(sma.ts[i])
	}
	return sum.Div(NewFromInt(sma.window))
}
