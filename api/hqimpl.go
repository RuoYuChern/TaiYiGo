package api

import (
	"container/list"
	"errors"
	"fmt"
	"time"

	"taiyigo.com/common"
	"taiyigo.com/facade/dto"
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
	}
	return tdata, nil
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
				topVol += t.Value
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
	hotMap := make(map[string]*dto.HotItem)
	for _, d := range dailyList {
		for _, tp := range d.Top20Vol {
			ht, ok := hotMap[tp.Name]
			if ok {
				ht.HighDay = d.Day
				ht.HotDays = ht.HotDays + 1
				ht.Vol = ht.Vol + tp.Value
			} else {
				ht = &dto.HotItem{LowDay: d.Day, HighDay: d.Day, HotDays: 1, Vol: tp.Value}
				ht.Symbol = tp.Name
				ht.Name = infra.GetSymbolName(ht.Symbol)
				hotMap[tp.Name] = ht
			}
		}
	}
	lhp := common.NewLp(len(hotMap)+1, func(a1, a2 any) int {
		v1 := a1.(*dto.HotItem)
		v2 := a2.(*dto.HotItem)
		if v1.Vol < v2.Vol {
			return -1
		} else if v1.Vol == v2.Vol {
			return 0
		} else {
			return 1
		}
	})
	totalVol := 0.0
	for _, v := range hotMap {
		v.Vol = common.FFloat(v.Vol/float64(v.HotDays), 4)
		totalVol += v.Vol
		lhp.Add(v)
	}
	data := make([]*dto.HotItem, lhp.Len())
	for lhp.Len() > 0 {
		off := lhp.Len() - 1
		di := lhp.Top().(*dto.HotItem)
		di.HotRate = int(di.Vol * 100 / totalVol)
		data[off] = di
	}
	return data, nil
}
