package indicators

import "math"

type wadback struct {
	ts     *TimeSeries
	window int
	cache  map[int]Decimal
}

type wadforwad struct {
	ts     *TimeSeries
	window int
	wady   Decimal
}

func NewWADIndicato(ts *TimeSeries, window int) Indicator {
	return &wadforwad{ts: ts, window: window, wady: ZERO}
}

func NewWADIndicato2(ts *TimeSeries, window int) Indicator {
	return &wadback{ts: ts, window: window, cache: map[int]Decimal{}}
}

func (wad *wadforwad) Calculate(index int) Decimal {
	todayClose := NewDecimal(wad.ts.Get(index).Close)
	yesClose := NewDecimal(wad.ts.Get(index).PreClose)
	vol := NewDecimal(float64(wad.ts.Get(index).Volume))
	var pm Decimal
	hr := NewDecimal(math.Max(wad.ts.Get(index).High, wad.ts.Get(index).PreClose))
	lr := NewDecimal(math.Min(wad.ts.Get(index).Low, wad.ts.Get(index).PreClose))
	if todayClose.Cmp(yesClose) > 0 {
		pm = todayClose.Sub(lr)
	} else if todayClose.Cmp(yesClose) < 0 {
		pm = todayClose.Sub(hr)
	} else {
		pm = ZERO
	}
	ad := pm.Mul(vol)
	wad.wady = ad.Add(wad.wady)
	return wad.wady
}

func (wad *wadback) Calculate(index int) Decimal {
	if index < 0 {
		return ZERO
	}
	wadValue, ok := wad.cache[index]
	if ok {
		return wadValue
	}
	todayClose := NewDecimal(wad.ts.Get(index).Close)
	yesClose := NewDecimal(wad.ts.Get(index).PreClose)
	vol := NewDecimal(float64(wad.ts.Get(index).Volume))
	var pm Decimal
	hr := NewDecimal(math.Max(wad.ts.Get(index).High, wad.ts.Get(index).PreClose))
	lr := NewDecimal(math.Min(wad.ts.Get(index).Low, wad.ts.Get(index).PreClose))
	if todayClose.Cmp(yesClose) > 0 {
		pm = todayClose.Sub(lr)
	} else if todayClose.Cmp(yesClose) < 0 {
		pm = todayClose.Sub(hr)
	} else {
		pm = ZERO
	}
	ad := pm.Mul(vol)
	wadValue = ad.Add(wad.Calculate(index - 1))
	wad.cache[index] = wadValue
	return wadValue
}
