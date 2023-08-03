package brain

import (
	"time"

	"google.golang.org/protobuf/proto"
	"taiyigo.com/brain/algor"
	"taiyigo.com/common"
	"taiyigo.com/facade/dto"
	"taiyigo.com/facade/tsdb"
	"taiyigo.com/facade/tsorder"
	"taiyigo.com/facade/tstock"
	"taiyigo.com/infra"
)

type FlowStart struct {
	common.Actor
}

type TradeFlow struct {
	common.Actor
	trade *tradingSign
}

func (ts *TradeFlow) Action() {
	orders := infra.GetUnFinishOrders()
	trade := ts.trade
	if orders.Len() == 0 {
		trade.lastTime = time.Now()
		return
	}
	symbols := make([]string, 0, orders.Len())
	for f := orders.Front(); f != nil; f = f.Next() {
		tOrd := f.Value.(*tsorder.TOrder)
		symbols = append(symbols, tOrd.Symbol)
	}
	priceMap, err := infra.BatchGetRealPrice(symbols)
	if err != nil {
		common.Logger.Infof("BatchGetRealPrice failed:%s", err)
		trade.lastTime = time.Now()
		return
	}
	for f := orders.Front(); f != nil; f = f.Next() {
		tOrd := f.Value.(*tsorder.TOrder)
		if tOrd.Status == dto.ORDER_BUY {
			doSell(tOrd, priceMap)
		} else if tOrd.Status == dto.ORDER_IDLE {
			doBuy(tOrd, priceMap)
		}
	}
	trade.lastTime = time.Now()
}

func doBuy(tOrd *tsorder.TOrder, priceMap map[string]*infra.CnStockPrice) {
	orderDay, err := common.ToDay(common.YYYYMMDD, tOrd.CreatDay)
	if err != nil {
		common.Logger.Infof("%s toDay error:%s", tOrd.CreatDay, err)
		tOrd.Status = dto.ORDER_CANCLE
		infra.SaveObject(infra.ORDER_TABLE, tOrd.OrderId, tOrd)
		return
	}
	price, ok := priceMap[tOrd.Symbol]
	if !ok {
		common.Logger.Infof("Find none such price:%s", tOrd.Symbol)
	} else {
		highPrice := tOrd.OrderPrice * 1.05
		if price.CurePrice <= float64(highPrice) {
			tOrd.Status = dto.ORDER_BUY
			tOrd.BuyDay = common.GetDay(common.YYYYMMDD, time.Now())
			tOrd.BuyPrice = float32(price.CurePrice)
			infra.SaveObject(infra.ORDER_TABLE, tOrd.OrderId, tOrd)
			common.Logger.Infof("OrderId:%s, Symbol:%s, Buy Price:%f, on Day:%s", tOrd.OrderId, tOrd.Symbol, tOrd.BuyPrice, tOrd.BuyDay)
			return
		}
	}
	diff := time.Since(orderDay)
	days := 10.0
	if (diff.Hours() / 24) < days {
		return
	}
	tOrd.Status = dto.ORDER_CANCLE
	infra.SaveObject(infra.ORDER_TABLE, tOrd.OrderId, tOrd)
	common.Logger.Infof("OrderId:%s, Symbol:%s is time out, and cancel", tOrd.OrderId, tOrd.Symbol)
}

func doSell(tOrd *tsorder.TOrder, priceMap map[string]*infra.CnStockPrice) {
	orderDay, _ := common.ToDay(common.YYYYMMDD, tOrd.CreatDay)
	price, ok := priceMap[tOrd.Symbol]
	if !ok {
		common.Logger.Infof("Find none such price:%s", tOrd.Symbol)
		return
	}
	lowPrice := tOrd.BuyPrice * 1.20
	if float64(lowPrice) <= price.CurePrice {
		tOrd.Status = dto.ORDER_SELL
		tOrd.SellDay = common.GetDay(common.YYYYMMDD, time.Now())
		tOrd.SellPrice = float32(price.CurePrice)
		infra.SaveObject(infra.ORDER_TABLE, tOrd.OrderId, tOrd)
		common.Logger.Infof("OrderId:%s, Symbol:%s, Sell Price:%f, on Day:%s", tOrd.OrderId, tOrd.Symbol, tOrd.BuyPrice, tOrd.BuyDay)
		return
	}

	diff := time.Since(orderDay)
	days := 10.0
	if (diff.Hours() / 24) < days {
		return
	}

	tOrd.Status = dto.ORDER_SELL
	tOrd.SellDay = common.GetDay(common.YYYYMMDD, time.Now())
	tOrd.SellPrice = float32(price.CurePrice)
	infra.SaveObject(infra.ORDER_TABLE, tOrd.OrderId, tOrd)
	common.Logger.Infof("OrderId:%s, Symbol:%s is time out, and sell", tOrd.OrderId, tOrd.Symbol)
}

func (fs *FlowStart) Action() {
	cnList := &tstock.CnBasicList{}
	err := infra.GetCnBasic(cnList)
	if err != nil {
		common.Logger.Infof("GetCnBasic failed: %s", err)
		return
	}
	lastDay, err := infra.GetByKey(infra.CONF_TABLE, infra.KEY_CNLOADHISTORY)
	if err != nil {
		common.Logger.Infof("GetByKey failed: %s", err)
		return
	}

	dayTime, err := common.ToDay(common.YYYYMMDD, lastDay)
	if err != nil {
		common.Logger.Infof("%s ToDay failed: %s", lastDay, err)
		return
	}
	common.Logger.Infof("Think for day:%s start", lastDay)
	algs := algor.GetAlgList()
	stfList := tstock.StfList{Numbers: 0, Stfs: make([]*tstock.StfInfo, 0)}
	for _, basic := range cnList.GetCnBasicList() {
		if common.FillteST(basic.Name) {
			continue
		}
		tql := infra.Gettsdb().OpenQuery(basic.Symbol)
		out, err := tql.GetPointN(uint64(dayTime.UnixMilli()), common.Conf.Brain.StfPriceCount)
		infra.Gettsdb().CloseQuery(tql)
		if err != nil {
			if !infra.IsEmpty(err) {
				common.Logger.Warnf("%s GetPointN failed:%s", basic.Symbol, err)
				break
			}
			continue
		}
		//转成candle
		candles := make([]*tstock.Candle, out.Len())
		off := 0
		for f := out.Front(); f != nil; f = f.Next() {
			candles[off] = &tstock.Candle{}
			dat := f.Value.(*tsdb.TsdbData)
			err = proto.Unmarshal(dat.Data, candles[off])
			if err != nil {
				common.Logger.Infof("Unmarshal:%s", err)
				return
			}
			off++
		}
		for front := algs.Front(); front != nil; front = front.Next() {
			think := front.Value.(algor.ThinkAlg)
			b, o := think.TAnalyze(candles)
			if b {
				stf := &tstock.StfInfo{Symbol: basic.Symbol, Status: "S", Name: basic.Name, Opt: o, Day: uint64(dayTime.UnixMilli())}
				stfList.Stfs = append(stfList.Stfs, stf)
				break
			}
		}
	}
	common.Logger.Infof("Think for day:%s over, find:%d", lastDay, len(stfList.Stfs))
	if len(stfList.Stfs) > 0 {
		err = infra.SaveStfList("S", lastDay, &stfList)
		if err != nil {
			common.Logger.Infof("Save day %s failed:%s", lastDay, err)
		}
		GetBrain().Subscript(TOPIC_STF, &MergeSTF{})
	}
}
