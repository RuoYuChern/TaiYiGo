package infra

import (
	"taiyigo.com/common"
	"taiyigo.com/facade/tstock"
)

type CnSharesDaily struct {
	Symbol   string  `json:"ts_code"`
	Day      string  `json:"trade_date"`
	Open     float64 `json:"open"`
	Close    float64 `json:"close"`
	PreClose float64 `json:"pre_close"`
	High     float64 `json:"high"`
	Low      float64 `json:"low"`
	Vol      float64 `json:"vol"`
	Amount   float64 `json:"amount"`
	PctChg   float64 `json:"pct_chg"`
	Change   float64 `json:"change"`
}

type CnSharesBasic struct {
	Symbol     string `json:"ts_code"`
	Name       string `json:"name"`
	Area       string `json:"area"`
	Industry   string `json:"industry"`
	FulName    string `json:"fullname"`
	EnName     string `json:"enname"`
	CnName     string `json:"cnspell"`
	Market     string `json:"market"`
	ExChange   string `json:"exchange"`
	Status     string `json:"list_status"`
	ListDate   string `json:"list_date"`
	DelistDate string `json:"delist_date"`
	IsHs       string `json:"is_hs"`
}

type TjDailyInfo struct {
	Symbol   string  `json:"symbol"`
	Day      string  `json:"tradeDate"`
	Open     float64 `json:"open"`
	Close    float64 `json:"close"`
	PreClose float64 `json:"preClose"`
	High     float64 `json:"high"`
	Low      float64 `json:"low"`
	Vol      float64 `json:"vol"`
	Amount   float64 `json:"amount"`
	PctChg   float64 `json:"pctChg"`
	Change   float64 `json:"change"`
}

type TjCnBasicInfo struct {
	Symbol     string `json:"symbol"`
	Name       string `json:"name"`
	Area       string `json:"area"`
	Industry   string `json:"industry"`
	FulName    string `json:"fullname"`
	EnName     string `json:"enname"`
	CnName     string `json:"cnspell"`
	Market     string `json:"market"`
	ExChange   string `json:"exchange"`
	Status     string `json:"status"`
	ListDate   string `json:"listDate"`
	DelistDate string `json:"delist_date"`
	IsHs       string `json:"is_hs"`
}

func ToTjDailyInfo(cnDaily *CnSharesDaily) *TjDailyInfo {
	tjDaily := &TjDailyInfo{Symbol: cnDaily.Symbol, Day: cnDaily.Day, Open: cnDaily.Open, Close: cnDaily.Close, PreClose: cnDaily.PreClose}
	tjDaily.High = cnDaily.High
	tjDaily.Low = cnDaily.Low
	tjDaily.Amount = cnDaily.Amount
	tjDaily.Vol = cnDaily.Vol
	tjDaily.PctChg = cnDaily.PctChg
	tjDaily.Change = cnDaily.Change
	return tjDaily
}

func ToCandle(dIt *TjDailyInfo) *tstock.Candle {
	candle := &tstock.Candle{}
	period, err := common.ToDay(common.YYYYMMDD, dIt.Day)
	if err != nil {
		common.Logger.Infof("ToDay failed:%s", err)
		return nil
	}
	candle.Period = uint64(period.UnixMilli())
	candle.Pcg = dIt.Change
	candle.Pcgp = dIt.PctChg
	candle.Open = dIt.Open
	candle.Close = dIt.Close
	candle.High = dIt.High
	candle.Low = dIt.Low
	candle.Volume = uint32(dIt.Vol)
	candle.PreClose = dIt.PreClose
	candle.Amount = dIt.Amount
	return candle
}

func ToCandle2(stkd *tstock.StockDaily) *tstock.Candle {
	candle := &tstock.Candle{}
	period, err := common.ToDay(common.YYYYMMDD, stkd.TradeDate)
	if err != nil {
		common.Logger.Infof("ToDay failed:%s", err)
		return nil
	}
	candle.Period = uint64(period.UnixMilli())
	candle.Pcg = stkd.Change
	candle.Pcgp = stkd.PctChg
	candle.Open = stkd.Open
	candle.Close = stkd.Close
	candle.High = stkd.High
	candle.Low = stkd.Low
	candle.Volume = uint32(stkd.Vol)
	candle.PreClose = stkd.PreClose
	candle.Amount = stkd.Amount
	return candle
}

func ToDaily(dIt *TjDailyInfo) *tstock.StockDaily {
	sdl := &tstock.StockDaily{}
	sdl.Symbol = dIt.Symbol
	sdl.TradeDate = dIt.Day
	sdl.Open = dIt.Open
	sdl.Close = dIt.Close
	sdl.PreClose = dIt.PreClose
	sdl.High = dIt.High
	sdl.Low = dIt.Low
	sdl.Change = dIt.Change
	sdl.PctChg = dIt.PctChg
	sdl.Amount = dIt.Amount
	sdl.Vol = dIt.Vol
	return sdl
}
