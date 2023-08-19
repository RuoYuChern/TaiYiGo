package indicators

import (
	"fmt"

	"taiyigo.com/facade/tstock"
)

type TimeSeries struct {
	Candles []*tstock.Candle
}

func (ts *TimeSeries) AddCandle(candle *tstock.Candle) bool {
	if candle == nil {
		panic(fmt.Errorf("error adding Candle: candle cannot be nil"))
	}

	last := ts.LastCandle()
	if last == nil || (candle.Period > last.Period) {
		ts.Candles = append(ts.Candles, candle)
		return true
	}

	return false
}

func NewTimeSeries(candles []*tstock.Candle) (t *TimeSeries) {
	t = new(TimeSeries)
	t.Candles = candles
	return t
}

func (ts *TimeSeries) LastCandle() *tstock.Candle {
	if len(ts.Candles) > 0 {
		return ts.Candles[ts.LastIndex()]
	}

	return nil
}

func (ts *TimeSeries) Get(i int) *tstock.Candle {
	return ts.Candles[i]
}

func (ts *TimeSeries) GetN(i int, opt GetValue) Decimal {
	return opt(ts.Get(i))
}

func (ts *TimeSeries) LastIndex() int {
	return len(ts.Candles) - 1
}
