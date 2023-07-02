package algor

import (
	"container/list"
	"strings"

	"taiyigo.com/facade/tstock"
	"taiyigo.com/indicators"
)

var (
	BUY  = "BUY"
	SELL = "SELL"
)

type ThinkAlg interface {
	Name() string
	TAnalyze(dat []*tstock.Candle) (bool, string)
	FAnalyze(dat []*tstock.Candle) (bool, string)
}

type Macd struct {
	ThinkAlg
	ts     *indicators.TimeSeries
	highWd int
	lowWd  int
	sWnd   int
	wadWd  int
	tWnd   int
}

func (macd *Macd) Name() string {
	return "mac"
}

func (macd *Macd) sp(dat []*tstock.Candle) (bool, string) {
	dlen := len(dat)
	day1 := macd.wadWd + macd.sWnd + macd.tWnd
	if dlen < day1 {
		return false, ""
	}

	macd.ts = indicators.NewTimeSeries(dat)
	sEndIdx := (dlen - 1) - macd.tWnd
	hSMA := indicators.NewSimpleMovingAverage2(macd.ts, indicators.GetHigh, macd.highWd)
	lSMA := indicators.NewSimpleMovingAverage2(macd.ts, indicators.GetLow, macd.lowWd)
	hma := hSMA.Calculate(sEndIdx)
	start := sEndIdx - macd.sWnd + 1
	isHit := indicators.IsLessIn(hma, macd.ts, indicators.GetClose, start, sEndIdx)
	if isHit {
		return true, BUY
	}
	lma := lSMA.Calculate(sEndIdx)
	isHit = indicators.IsBiggerIn(lma, macd.ts, indicators.GetClose, start, sEndIdx)
	if isHit {
		return true, SELL
	}
	return false, ""
}

func (macd *Macd) TAnalyze(dat []*tstock.Candle) (bool, string) {
	b, s := macd.sp(dat)
	if !b {
		return b, s
	}
	endIdx := len(dat) - 1
	wad := indicators.NewWADIndicato(macd.ts, macd.wadWd)
	startId := endIdx - macd.wadWd + 1
	adList := make([]indicators.Decimal, macd.wadWd)
	for off := 0; off < macd.wadWd; off++ {
		adList[off] = wad.Calculate(startId + off)
	}
	wadSma := indicators.NewWadSma(adList, macd.wadWd)
	wad57 := indicators.ZERO
	for off := 0; off < macd.wadWd; off++ {
		wad57 = wadSma.Calculate(off)
	}

	adHigh := len(adList) - 1
	adLow := adHigh - macd.tWnd + 1
	if strings.Compare(s, BUY) == 0 {
		return indicators.IsLessInDec(wad57, adList, adLow, adHigh), s
	} else {
		return indicators.IsBiggerInDec(wad57, adList, adLow, adHigh), s
	}
}

func (macd *Macd) FAnalyze(dat []*tstock.Candle) (bool, string) {
	return false, ""
}

func GetAlgList() *list.List {
	algs := list.New()
	algs.PushBack(&Macd{highWd: 10, lowWd: 8, sWnd: 2, wadWd: 57, tWnd: 2})
	return algs
}
