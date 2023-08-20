package algor

import (
	"math"
	"strings"

	"taiyigo.com/facade/tstock"
	"taiyigo.com/indicators"
)

type bValueIndex struct {
	value    float32
	position int
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

func (boll *Boll) TAnalyze(dat []*tstock.Candle) (bool, string) {
	return false, ""
}

func (macd *Boll) FAnalyze(dat []*tstock.Candle) (bool, string) {
	return false, ""
}

func (ta *VolPrice) TAnalyze(dat []*tstock.Candle) (bool, string) {
	total := len(dat)
	if total < (ta.windows + ta.tWnd) {
		return false, "vp"
	}
	offset := (total - ta.windows - ta.tWnd)
	maxVol := bValueIndex{value: 0, position: 0}
	minVol := bValueIndex{value: math.MaxFloat32, position: 0}
	maxClose := bValueIndex{value: 0, position: 0}
	for off := 0; off < ta.windows; off++ {
		candle := dat[off+offset]
		if candle.Volume > uint32(maxVol.value) {
			maxVol.value = float32(candle.Volume)
			maxVol.position = off
		}
		if candle.Volume < uint32(minVol.value) {
			minVol.value = float32(candle.Volume)
			minVol.position = off
		}
		if candle.Close > float64(maxClose.value) {
			maxClose.value = float32(candle.Close)
			maxClose.position = off
		}
	}
	testAvgVol := float32(0)
	testAvgClose := float64(0)
	for off := 0; off < ta.tWnd; off++ {
		candle := dat[off+ta.windows+offset]
		testAvgVol += float32(candle.Volume)
		testAvgClose += candle.Close
	}
	testAvgVol = testAvgVol / float32(ta.tWnd)
	testAvgClose = testAvgClose / float64(ta.tWnd)
	if maxClose.value > float32(2*testAvgClose) && maxClose.position <= (total-100) {
		return true, BUY
	}
	if testAvgVol >= 3*maxVol.value {
		return true, BUY
	}
	if testAvgVol > 1.5*maxVol.value && testAvgClose > float64(maxClose.value) {
		return true, BUY
	}
	return false, ""
}

func (ta *VolPrice) FAnalyze(dat []*tstock.Candle) (bool, string) {
	return false, ""
}
