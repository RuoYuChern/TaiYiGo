package algor

import (
	"container/list"

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

type Boll struct {
	ThinkAlg
	ts      *indicators.TimeSeries
	windows int
	tWnd    int
}

type kDJ struct {
	ThinkAlg
	ts      *indicators.TimeSeries
	windows int
	tWnd    int
}

func GetAlgList() *list.List {
	algs := list.New()
	algs.PushBack(&Macd{highWd: 10, lowWd: 8, sWnd: 3, wadWd: 57, tWnd: 3})
	algs.PushBack(&Boll{windows: 20, tWnd: 3})
	return algs
}
