package api

import (
	"errors"
	"time"

	"taiyigo.com/common"
	"taiyigo.com/facade/dto"
	"taiyigo.com/indicators"
	"taiyigo.com/infra"
)

func calSymbolTrend(symbol string) ([]*dto.SymbolDaily, error) {
	lastDay, err := infra.GetByKey(infra.CONF_TABLE, infra.KEY_CNLOADHISTORY)
	if err != nil {
		common.Logger.Infof("GetByKey failed: %s", err)
		return nil, err
	}

	dataList, err := infra.GetSymbolNPoint(symbol, lastDay, 250)
	if err != nil {
		common.Logger.Infof("GetSymbolNPoint failed: %s", err)
		return nil, err
	}
	datLen := len(dataList)
	if datLen == 0 {
		return nil, errors.New("not found")
	}
	ts := indicators.NewTimeSeries(dataList)
	lsma := indicators.NewSimpleMovingAverage(ts, indicators.GetClose, 10)
	ssma := indicators.NewSimpleMovingAverage(ts, indicators.GetClose, 8)
	mnt := indicators.NewMtn(ts, indicators.GetClose, 10)
	startOff := 10
	tdata := make([]*dto.SymbolDaily, datLen-startOff)
	for off := 0; off < datLen; off++ {
		if off < startOff {
			lsma.Calculate(off)
			ssma.Calculate(off)
			continue
		}
		candle := ts.Get(off)
		period := time.Unix(int64(candle.Period/1000), 0)
		tdata[off-startOff] = &dto.SymbolDaily{Day: common.GetDay(common.YYYYMMDD, period), Open: candle.Open, Close: candle.Close}
		tdata[off-startOff].High = candle.High
		tdata[off-startOff].Low = candle.Low
		tdata[off-startOff].Vol = float64(candle.Volume)
		tdata[off-startOff].Hld = common.FFloat(candle.High-candle.Low, 2)
		tdata[off-startOff].LSma = lsma.Calculate(off).FormatFloat(2)
		tdata[off-startOff].SSma = ssma.Calculate(off).FormatFloat(2)
		tdata[off-startOff].Mtm = mnt.Calculate(off).FormatFloat(2)
	}
	return tdata, nil
}
