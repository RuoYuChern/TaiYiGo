package api

import (
	"container/list"
	"errors"
	"fmt"
	"strings"
	"time"

	"taiyigo.com/common"
	"taiyigo.com/facade/dto"
	"taiyigo.com/facade/tsorder"
	"taiyigo.com/facade/tstock"
	"taiyigo.com/indicators"
	"taiyigo.com/infra"
)

var (
	MAX_GET_LEN  = 300
	MAX_DAYS_LEN = 200
	WAN          = 10000.0
	SZ           = "sz399001"
	SH           = "sh000001"
	SZ_300       = "sz399300"
	HS_300       = "sh000300"
)

func calSymbolTrend(symbol string) ([]*dto.SymbolDaily, error) {
	lastDay, err := infra.GetByKey(infra.CONF_TABLE, infra.KEY_CNLOADHISTORY)
	if err != nil {
		common.Logger.Infof("GetByKey failed: %s", err)
		return nil, err
	}

	dataList, err := infra.FGetSymbolNPoint(symbol, lastDay, MAX_GET_LEN)
	if err != nil {
		common.Logger.Infof("GetSymbolNPoint failed: %s", err)
		return nil, err
	}
	datLen := len(dataList)
	if datLen == 0 {
		return nil, errors.New("not found")
	}
	if datLen <= 10 {
		return nil, errors.New("not enough data")
	}
	ts := indicators.NewTimeSeries(dataList)
	lsma := indicators.NewSimpleMovingAverage(ts, indicators.GetClose, 10)
	ssma := indicators.NewSimpleMovingAverage(ts, indicators.GetClose, 8)
	mnt := indicators.NewMtn(ts, indicators.GetClose, 10)
	startOff := (datLen - MAX_DAYS_LEN)
	if startOff < 10 {
		startOff = 10
	}
	tdata := make([]*dto.SymbolDaily, datLen-startOff)
	for off := 0; off < datLen; off++ {
		if off < startOff {
			lsma.Calculate(off)
			ssma.Calculate(off)
			continue
		}
		candle := ts.Get(off)
		period := time.UnixMilli(int64(candle.Period))
		tdata[off-startOff] = &dto.SymbolDaily{Day: common.GetDay(common.YYYYMMDD, period), Open: candle.Open, Close: candle.Close}
		tdata[off-startOff].High = candle.High
		tdata[off-startOff].Low = candle.Low
		tdata[off-startOff].Vol = common.FFloat(float64(candle.Volume)/WAN, 3)
		tdata[off-startOff].Hld = common.FFloat(candle.High-candle.Low, 2)
		tdata[off-startOff].LSma = lsma.Calculate(off).FormatFloat(2)
		tdata[off-startOff].SSma = ssma.Calculate(off).FormatFloat(2)
		tdata[off-startOff].Mtm = mnt.Calculate(off).FormatFloat(2)
		tdata[off-startOff].FundIn = common.FFloat(float64(candle.Volume)*(candle.Close-candle.PreClose)/WAN, 3)
	}
	return tdata, nil
}

func doPostQuantCal(name, symbol, method string) *dto.HqCommonRsp {
	lastDay, err := infra.GetByKey(infra.CONF_TABLE, infra.KEY_CNLOADHISTORY)
	hqRsp := &dto.HqCommonRsp{Code: 200, Msg: "OK"}
	if err != nil {
		common.Logger.Infof("GetByKey failed: %s", err)
		hqRsp.Code = 500
		hqRsp.Msg = err.Error()
		return hqRsp
	}
	dataList, err := infra.FGetSymbolNPoint(symbol, lastDay, 250)
	if err != nil {
		hqRsp.Code = 500
		hqRsp.Msg = err.Error()
		return hqRsp
	}
	tid := common.GetTid(common.Conf.Quotes.Sault)
	rsp := infra.DoPostQuant(tid, name, symbol, method, dataList)
	hqRsp.Code = rsp.Status
	hqRsp.Msg = rsp.Msg
	return hqRsp
}

func getStockNPoint(symbol string, lastDay string, num int) ([]*dto.CnDaily, error) {
	dataList, err := infra.FGetSymbolNPoint(symbol, lastDay, num)
	if err != nil {
		common.Logger.Infof("GetSymbolNPoint failed: %s", err)
		return nil, err
	}
	datLen := len(dataList)
	if datLen == 0 {
		return nil, errors.New("not found")
	}
	data := make([]*dto.CnDaily, datLen)
	for off := 0; off < datLen; off++ {
		stk := dataList[off]
		data[off] = &dto.CnDaily{Symbol: symbol, Open: stk.Open, Close: stk.Close, PreClose: stk.PreClose}
		data[off].Day = common.GetDay(common.YYYYMMDD, time.UnixMilli(int64(stk.Period)))
		data[off].High = stk.High
		data[off].Low = stk.Low
		data[off].Amount = stk.Amount
		data[off].Vol = stk.Volume
		data[off].PctChg = stk.Pcgp
		data[off].Change = stk.Pcg
	}
	return data, nil
}

func calLatestDash() (*dto.DashData, error) {
	dbms := infra.GetLastNMonthDash(12)
	dailyList := list.New()
	totalMon := 0
	totalDays := 0
	if dbms != nil {
		totalMon = len(dbms)
		for off := 0; off < totalMon; off++ {
			dailyList.PushBack(dbms[off])
			totalDays += len(dbms[off].DailyDash)
		}
	}
	year := time.Now().Year() - 1
	for totalMon < 12 {
		fn := fmt.Sprintf("%d_hdbd.dat", year)
		dby := &tstock.DashBoardYear{}
		err := infra.GetMsg(fn, dby)
		if err != nil {
			common.Logger.Warnf("GetMsg:%s, failed:%s", fn, err)
			break
		}
		tm := len(dby.MonthDash)
		totalMon += tm
		for off := tm - 1; off >= 0; off-- {
			dailyList.PushFront(dby.MonthDash[off])
			totalDays += len(dby.MonthDash[off].DailyDash)
		}
		year -= 1
	}
	skipOff := (totalDays - MAX_DAYS_LEN)
	size := totalDays
	if size > MAX_DAYS_LEN {
		size = MAX_DAYS_LEN
	}
	data := &dto.DashData{}
	daily := make([]*dto.DashDaily, 0, size)
	for f := dailyList.Front(); f != nil; f = f.Next() {
		dbm := f.Value.(*tstock.DashBoardMonth)
		for _, dItem := range dbm.DailyDash {
			if skipOff > 0 {
				skipOff--
				continue
			}
			dto := &dto.DashDaily{Day: dItem.Day}
			dto.Amount = common.FFloat(dItem.TotalAmount/WAN, 4)
			dto.Vol = common.FFloat(dItem.TotalVol/WAN, 4)
			dto.DownLimit = len(dItem.DownLimit)
			dto.UpLimit = len(dItem.UpLimit)
			dto.UpStocks = int(dItem.UpStocks)
			dto.DownStocks = int(dItem.DownStocks)
			totalStocks := int(dItem.Stocks)
			topVol := 0.0
			for _, t := range dItem.Top20Vol {
				topVol += t.Vol
			}
			dto.Top20Vol = common.FFloat(topVol/WAN, 4)
			if totalStocks == 0 {
				totalStocks = 1
			}
			dto.Mood = int((dto.UpStocks * 100) / totalStocks)
			daily = append(daily, dto)
		}
	}
	data.Daily = daily
	hq := infra.GetKDataCache(SZ)
	if hq != nil {
		data.SZDaily = hq.Data
	} else {
		data.SZDaily = make([]*dto.CnStockKData, 0)
	}
	hq = infra.GetKDataCache(SH)
	if hq != nil {
		data.SHDaily = hq.Data
	} else {
		data.SHDaily = make([]*dto.CnStockKData, 0)
	}
	hq = infra.GetKDataCache(HS_300)
	if hq != nil {
		data.HSDaily = hq.Data
	} else {
		data.HSDaily = make([]*dto.CnStockKData, 0)
	}

	return data, nil
}

func getLastUpDown(day string) ([]*dto.UpDownItem, error) {
	dbv := infra.GetLastDayDash(day)
	if dbv == nil {
		common.Logger.Infof("No data")
		return nil, errors.New("no data")
	}
	data := make([]*dto.UpDownItem, len(dbv.UpLimit)+len(dbv.DownLimit))
	off := 0
	for _, u := range dbv.UpLimit {
		udi := &dto.UpDownItem{Day: dbv.Day, Symbol: u.Symbol}
		udi.Flag = 1
		udi.Name = infra.GetSymbolName(u.Symbol)
		udi.Close = float32(u.Close)
		udi.PreClose = float32(u.PreClose)
		udi.PctChg = float32(u.Rate)
		data[off] = udi
		off++
	}
	for _, dn := range dbv.DownLimit {
		udi := &dto.UpDownItem{Day: dbv.Day, Symbol: dn.Symbol}
		udi.Flag = 0
		udi.Name = infra.GetSymbolName(dn.Symbol)
		udi.Close = float32(dn.Close)
		udi.PreClose = float32(dn.PreClose)
		udi.PctChg = float32(dn.Rate)
		data[off] = udi
		off++
	}
	return data, nil
}

func getLatestHot(day string) ([]*dto.HotItem, error) {
	dbv := infra.GetLastDayDash(day)
	if dbv == nil {
		common.Logger.Infof("No data")
		return nil, errors.New("no data")
	}
	total := len(dbv.Top20Vol)
	lp := common.NewLp(total+1, func(a1, a2 any) int {
		v1 := int64(a1.(*tstock.TopSymbol).Vol)
		v2 := int64(a2.(*tstock.TopSymbol).Vol)
		diff := int(v1 - v2)
		return diff
	})
	for _, dv := range dbv.Top20Vol {
		lp.Add(dv)
	}

	data := make([]*dto.HotItem, total)
	for off := total; off > 0; off-- {
		tv := lp.Top()
		if tv == nil {
			break
		}
		dv := tv.(*tstock.TopSymbol)
		idx := off - 1
		data[idx] = &dto.HotItem{Symbol: dv.Name, Day: dbv.Day, Vol: dv.Vol, Open: dv.Open, Close: dv.Close}
		data[idx].Name = infra.GetSymbolName(data[idx].Symbol)
	}
	return data, nil
}

func doGetTradingStat() (*dto.TradingStatDto, error) {
	year := common.GetYear(time.Now())
	orders, err := infra.GetOrdersByYear(year)
	if err != nil {
		return nil, err
	}
	statDto := &dto.TradingStatDto{Orders: make([]*dto.TradingDto, orders.Len())}
	if orders.Len() > 0 {
		lp := common.NewLp(orders.Len(), func(a1, a2 any) int {
			o1 := a1.(*dto.TradingDto)
			o2 := a2.(*dto.TradingDto)
			return strings.Compare(o1.OrderDate, o2.OrderDate)
		})

		symbols := make([]string, 0)
		for f := orders.Front(); f != nil; f = f.Next() {
			order := &dto.TradingDto{}
			tOrd := f.Value.(*tsorder.TOrder)
			order.Name = tOrd.Name
			order.Symbol = tOrd.Symbol
			order.OrderPrice = common.FloatToStr(float64(tOrd.OrderPrice), 2)
			order.OrderDate = tOrd.CreatDay

			if tOrd.Status == dto.ORDER_BUY || tOrd.Status == dto.ORDER_IDLE {
				order.Status = "IDLE"
				symbols = append(symbols, tOrd.Symbol)
				if tOrd.Status == dto.ORDER_BUY {
					statDto.BuyOrders = statDto.BuyOrders + 1
				}
			}

			if tOrd.Status == dto.ORDER_BUY || tOrd.Status == dto.ORDER_SELL {
				order.BuyPrice = common.FloatToStr(float64(tOrd.BuyPrice), 2)
				order.BuyDate = tOrd.BuyDay
				statDto.Amount += (float64(tOrd.BuyPrice) * float64(tOrd.Vol))
				order.Status = "BUY"
			}

			if tOrd.Status == dto.ORDER_SELL {
				order.SellPrice = common.FloatToStr(float64(tOrd.SellPrice), 2)
				order.SellDate = tOrd.SellDay
				order.Status = "SELL"
				diff := (tOrd.SellPrice - tOrd.BuyPrice)
				statDto.Pnl += float64(diff * float32(tOrd.Vol))
				if diff > 0 {
					order.Flag = "+"
					statDto.SuccessOrders += 1
				} else if diff < 0 {
					order.Flag = "-"
					statDto.FailedOrders += 1
				} else {
					order.Flag = "="
				}
			}

			if tOrd.Status == dto.ORDER_CANCLE {
				order.Status = "CANCLE"
				statDto.CancelOrders += 1
			}
			lp.Push(order)
		}
		priceMap, err := infra.BatchGetRealPrice(symbols)
		statDto.Vol = lp.Len()
		if err != nil {
			common.Logger.Infof("BatchGetRealPrice failed:%s", err)
		}
		off := lp.Len()
		for {
			t := lp.Top()
			if t == nil {
				break
			}
			ord := t.(*dto.TradingDto)
			if priceMap != nil {
				price, ok := priceMap[ord.Symbol]
				if ok {
					ord.CurePrice = common.FloatToStr(float64(price.CurePrice), 2)
				}
			}
			statDto.Orders[off-1] = ord
			off -= 1
		}
		statDto.Amount = common.FFloat(statDto.Amount, 4)
		statDto.Pnl = common.FFloat(statDto.Pnl, 4)
	}
	return statDto, nil
}
