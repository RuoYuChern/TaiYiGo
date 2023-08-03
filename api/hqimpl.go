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
)

func calSymbolTrend(symbol string) ([]*dto.SymbolDaily, error) {
	lastDay, err := infra.GetByKey(infra.CONF_TABLE, infra.KEY_CNLOADHISTORY)
	if err != nil {
		common.Logger.Infof("GetByKey failed: %s", err)
		return nil, err
	}

	dataList, err := infra.GetSymbolNPoint(symbol, lastDay, MAX_GET_LEN)
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
		period := time.Unix(int64(candle.Period/1000), 0)
		tdata[off-startOff] = &dto.SymbolDaily{Day: common.GetDay(common.YYYYMMDD, period), Open: candle.Open, Close: candle.Close}
		tdata[off-startOff].High = candle.High
		tdata[off-startOff].Low = candle.Low
		tdata[off-startOff].Vol = common.FFloat(float64(candle.Volume)/10000, 3)
		tdata[off-startOff].Hld = common.FFloat(candle.High-candle.Low, 2)
		tdata[off-startOff].LSma = lsma.Calculate(off).FormatFloat(2)
		tdata[off-startOff].SSma = ssma.Calculate(off).FormatFloat(2)
		tdata[off-startOff].Mtm = mnt.Calculate(off).FormatFloat(2)
		tdata[off-startOff].FundIn = common.FFloat(float64(candle.Volume)*(candle.Close-candle.PreClose)/10000, 3)
	}
	return tdata, nil
}

func getStockNPoint(symbol string, lastDay string, num int) ([]*dto.CnDaily, error) {
	dataList, err := infra.GetSymbolNPoint(symbol, lastDay, num)
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

func calLatestDash() ([]*dto.DashDaily, error) {
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
	data := make([]*dto.DashDaily, 0, size)
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
			data = append(data, dto)
		}
	}
	return data, nil
}

func getLastUpDown() ([]*dto.UpDownItem, error) {
	dailyList := infra.GetLastNDayDash(5, false)
	if dailyList == nil {
		common.Logger.Infof("No data")
		return nil, errors.New("no data")
	}
	data := make([]*dto.UpDownItem, 0)
	for _, d := range dailyList {
		for _, u := range d.UpLimit {
			udi := &dto.UpDownItem{Day: d.Day, Symbol: u}
			udi.Flag = 1
			udi.Name = infra.GetSymbolName(u)
			data = append(data, udi)
		}
		for _, dn := range d.DownLimit {
			udi := &dto.UpDownItem{Day: d.Day, Symbol: dn}
			udi.Flag = 0
			udi.Name = infra.GetSymbolName(dn)
			data = append(data, udi)
		}
	}
	return data, nil
}

func getLatestHot() ([]*dto.HotItem, error) {
	dailyList := infra.GetLastNDayDash(20, true)
	if dailyList == nil {
		common.Logger.Infof("No data")
		return nil, errors.New("no data")
	}
	data := make([]*dto.HotItem, 0)
	for _, dbv := range dailyList {
		for _, dv := range dbv.Top20Vol {
			hi := &dto.HotItem{Symbol: dv.Name, Day: dbv.Day, Vol: dv.Vol, Open: dv.Open, Close: dv.Close}
			hi.Name = infra.GetSymbolName(hi.Symbol)
			data = append(data, hi)
		}
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
		off := orders.Len()
		for {
			t := lp.Top()
			if t == nil {
				break
			}
			ord := t.(*dto.TradingDto)
			if priceMap != nil {
				price, ok := priceMap[ord.Symbol]
				if ok {
					ord.CurePrice = common.FloatToStr(price.CurePrice, 2)
				}
			}
			statDto.Orders[off-1] = ord
			off -= 1
		}
	}
	return statDto, nil
}
