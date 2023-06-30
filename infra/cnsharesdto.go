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
