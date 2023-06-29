package algor

import (
	"container/list"

	"taiyigo.com/indicators"
)

type ThinkAlg interface {
	Name() string
	S(dat *list.List) (bool, string)
	T(dat *list.List) (bool, string)
	F(dat *list.List) (bool, string)
}

type Macd struct {
	ThinkAlg
}

func (macd *Macd) Name() string {
	return "mac"
}

func (macd *Macd) S(dat *list.List) (bool, string) {
	if dat.Len() < 12 {
		return false, ""
	}

	ts := indicators.NewTimeSeries(dat)
	hwd := 10
	lwd := 8
	endIdx := dat.Len() - 1
	hSMA := indicators.NewSimpleMovingAverage2(ts, indicators.GetHigh, hwd)
	lSMA := indicators.NewSimpleMovingAverage2(ts, indicators.GetLow, lwd)
	hma := hSMA.Calculate(endIdx)

	closeP1 := indicators.NewDecimal(ts.Get((endIdx - 1)).Close)
	closeP2 := indicators.NewDecimal(ts.Get((endIdx)).Close)
	if hma.LT(closeP1) && hma.LT(closeP2) {
		return true, "BUY"
	}
	lma := lSMA.Calculate(endIdx)
	if lma.GT(closeP1) && lma.GT(closeP2) {
		return true, "SELL"
	}
	return false, ""
}
func (macd *Macd) T(dat *list.List) (bool, string) {
	return false, ""
}
func (macd *Macd) F(dat *list.List) (bool, string) {
	return false, ""
}

func GetAlgList() *list.List {
	algs := list.New()
	algs.PushBack(&Macd{})
	return algs
}
