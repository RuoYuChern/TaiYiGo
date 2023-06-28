package indicators

import (
	"taiyigo.com/facade/tstock"
)

type GetValue func(cursor *tstock.Candle) Decimal

var (
	GetClose  = func(cursor *tstock.Candle) Decimal { return NewDecimal(cursor.Close) }
	GetOpen   = func(cursor *tstock.Candle) Decimal { return NewDecimal(cursor.Open) }
	GetHigh   = func(cursor *tstock.Candle) Decimal { return NewDecimal(cursor.High) }
	GetLow    = func(cursor *tstock.Candle) Decimal { return NewDecimal(cursor.Low) }
	GetAmount = func(cursor *tstock.Candle) Decimal { return NewDecimal(cursor.Amount) }
	GetVol    = func(cursor *tstock.Candle) Decimal { return NewFromInt(int(cursor.Volume)) }
)

type Indicator interface {
	Calculate(index int) Decimal
}
