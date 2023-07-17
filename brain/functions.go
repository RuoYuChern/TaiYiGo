package brain

import (
	"errors"
	"strings"
	"time"

	"go.uber.org/ratelimit"
	"google.golang.org/protobuf/proto"
	"taiyigo.com/common"
	"taiyigo.com/facade/tsdb"
	"taiyigo.com/facade/tstock"
	"taiyigo.com/indicators"
	"taiyigo.com/infra"
)

func LoadSymbolDaily(cnList *tstock.CnBasicList, lowDay string, highDay string, cnShareStatus map[string]string) (int, error) {
	limter := ratelimit.New(500, ratelimit.Per(time.Minute))
	tsDb := infra.Gettsdb()
	rangeTotal := 0
	ndbc := indicators.NewNDbc()
	for _, cnBasic := range cnList.CnBasicList {
		cnShareLastDay, err := infra.GetByKey(infra.CONF_TABLE, cnBasic.Symbol)
		startDay := lowDay
		if err == nil && cnShareLastDay != "" {
			if strings.Compare(highDay, cnShareLastDay) <= 0 {
				continue
			}
			if strings.Compare(startDay, cnShareLastDay) < 0 {
				startDay = cnShareLastDay
			}
		}
		limter.Take()
		daily, err := infra.GetDailyFromTj(cnBasic.Symbol, startDay, highDay)
		if err != nil {
			common.Logger.Warnf("Load symbol:%s, range[%s,%s],failed:%s", cnBasic.Symbol, startDay, highDay, err)
			return 0, err
		}

		total := len(daily)
		rangeTotal += total
		if total == 0 {
			continue
		}
		tbl := tsDb.OpenAppender(cnBasic.Symbol)
		for dOff := 0; dOff < total; dOff++ {
			dailyInfo := &daily[dOff]
			candle := infra.ToCandle(dailyInfo)
			if candle == nil {
				common.Logger.Warnf("Load symbol:%s, range[%s,%s],failed", cnBasic.Symbol, startDay, highDay)
				tsDb.CloseAppender(tbl)
				return 0, errors.New("to candle failed")
			}
			out, err := proto.Marshal(candle)
			if err != nil {
				common.Logger.Warnf("Load symbol:%s, range[%s,%s],failed:%s", cnBasic.Symbol, startDay, highDay, err)
				tsDb.CloseAppender(tbl)
				return 0, err
			}
			tsData := &tsdb.TsdbData{Timestamp: candle.Period, Data: out}
			err = tbl.Append(tsData)
			if err != nil {
				common.Logger.Warnf("Save symbol:%s, range[%s,%s],failed:%s", cnBasic.Symbol, startDay, highDay, err)
				tsDb.CloseAppender(tbl)
				return 0, err
			}
			ndbc.Cal(daily[dOff].Day, daily[dOff].Symbol, candle)
			cnShareStatus[daily[dOff].Symbol] = daily[dOff].Day
		}
		tsDb.CloseAppender(tbl)
	}
	ndbc.Save()
	return rangeTotal, nil
}
