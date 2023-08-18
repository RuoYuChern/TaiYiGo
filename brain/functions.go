package brain

import (
	"errors"
	"strings"
	"time"

	"github.com/tao/faststore"
	"github.com/tao/faststore/api"
	"go.uber.org/ratelimit"
	"google.golang.org/protobuf/proto"
	"taiyigo.com/common"
	"taiyigo.com/facade/tsdb"
	"taiyigo.com/facade/tstock"
	"taiyigo.com/indicators"
	"taiyigo.com/infra"
)

func LoadSymbolDaily(cnList *tstock.CnBasicList, lowDay string, highDay string, cnShareStatus map[string]string) (int, error) {
	limter := ratelimit.New(400, ratelimit.Per(time.Minute))
	tsDb := infra.Gettsdb()
	rangeTotal := 0
	ndbc := indicators.NewNDbc()
	tdlg := infra.OpenTlg()
	for _, cnBasic := range cnList.CnBasicList {
		cnShareLastDay, _ := infra.GetByKey(infra.CONF_TABLE, cnBasic.Symbol)
		startDay := lowDay
		if cnShareLastDay != "" {
			if strings.Compare(highDay, cnShareLastDay) <= 0 {
				continue
			}
			if strings.Compare(startDay, cnShareLastDay) < 0 {
				startDay = cnShareLastDay
			}
		}
		limter.Take()
		daily, err := infra.QueryCnShareDailyRange(cnBasic.Symbol, startDay, highDay)
		if err != nil {
			common.Logger.Warnf("Load symbol:%s, range[%s,%s],failed:%s", cnBasic.Symbol, startDay, highDay, err)
			ndbc.Save()
			tdlg.Close()
			return 0, err
		}
		if daily == nil {
			continue
		}
		total := len(daily)
		rangeTotal += total
		if total == 0 {
			continue
		}
		tbl := tsDb.OpenAppender(cnBasic.Symbol)
		call := faststore.FsTsdbGet("cnshares", cnBasic.Symbol)
		for dOff := 0; dOff < total; dOff++ {
			dailyInfo := daily[dOff]
			candle := infra.ToCandle(dailyInfo)
			if candle == nil {
				common.Logger.Warnf("Load symbol:%s, range[%s,%s],failed", cnBasic.Symbol, startDay, highDay)
				tsDb.CloseAppender(tbl)
				ndbc.Save()
				tdlg.Close()
				return 0, errors.New("to candle failed")
			}
			out, err := proto.Marshal(candle)
			if err != nil {
				common.Logger.Warnf("Load symbol:%s, range[%s,%s],failed:%s", cnBasic.Symbol, startDay, highDay, err)
				tsDb.CloseAppender(tbl)
				ndbc.Save()
				tdlg.Close()
				return 0, err
			}
			tsData := &tsdb.TsdbData{Timestamp: candle.Period, Data: out}
			fv := api.FstTsdbValue{Timestamp: int64(candle.Period), Data: out}
			err = tbl.Append(tsData)
			if err != nil {
				common.Logger.Warnf("Save symbol:%s, range[%s,%s],failed:%s", cnBasic.Symbol, startDay, highDay, err)
				tsDb.CloseAppender(tbl)
				ndbc.Save()
				tdlg.Close()
				return 0, err
			}
			call.Append(&fv)
			ndbc.Cal(dailyInfo.Day, dailyInfo.Symbol, candle)
			sbdl := infra.ToDaily(dailyInfo)
			tdlg.Append(dailyInfo.Day, sbdl)
			cnShareStatus[dailyInfo.Symbol] = dailyInfo.Day
		}
		tsDb.CloseAppender(tbl)
		call.Close()
	}
	ndbc.Save()
	tdlg.Close()
	return rangeTotal, nil
}
