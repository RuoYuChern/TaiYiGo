package dto

import "encoding/json"

type CnAdminCmd struct {
	Opt   string `json:"opt" binding:"required"`
	Cmd   string `json:"cmd" binding:"required"`
	Value string `json:"value"`
}

type JustifyReq struct {
	Table string `json:"table" binding:"required"`
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

type CnDaily struct {
	Symbol   string  `json:"symbol"`
	Day      string  `json:"trade_date"`
	Open     float64 `json:"open"`
	Close    float64 `json:"close"`
	PreClose float64 `json:"pre_close"`
	High     float64 `json:"high"`
	Low      float64 `json:"low"`
	Vol      uint32  `json:"vol"`
	Amount   float64 `json:"amount"`
	PctChg   float64 `json:"pct_chg"`
	Change   float64 `json:"change"`
}

type SymbolDaily struct {
	Day    string  `json:"day"`
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Vol    float64 `json:"vol"`
	Mtm    float64 `json:"mtn"`
	Hld    float64 `json:"hld"`
	LSma   float64 `json:"lsma"`
	SSma   float64 `json:"ssma"`
	FundIn float64 `json:"fundIn"`
}

type PairDaily struct {
	Day    string  `json:"day"`
	FClose float64 `json:"fClose"`
	FVol   float64 `json:"fVol"`
	FMtm   float64 `json:"fMtn"`
	FLSma  float64 `json:"fLsma"`
	FSSma  float64 `json:"fSsma"`
	SClose float64 `json:"sClose"`
	SVol   float64 `json:"sVol"`
	SMtm   float64 `json:"sMtn"`
	SLSma  float64 `json:"sLsma"`
	SSSma  float64 `json:"sSsma"`
}

type SymbolTrendResponse struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	Data []*SymbolDaily `json:"data"`
}

type PaireSResponse struct {
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Data []*PairDaily `json:"data"`
}

type StfItem struct {
	Symbol  string `json:"symbol"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Opt     string `json:"opt"`
	LowDay  string `json:"lowday"`
	HighDay string `json:"highday"`
}

type StfResponse struct {
	Code int        `json:"code"`
	Msg  string     `json:"msg"`
	Data []*StfItem `json:"data"`
}

type DashDaily struct {
	Day        string  `json:"day"`
	Vol        float64 `json:"vol"`
	Top20Vol   float64 `json:"topVol"`
	Amount     float64 `json:"amount"`
	UpStocks   int     `json:"upStocks"`
	DownStocks int     `json:"downStocks"`
	UpLimit    int     `json:"upLimit"`
	DownLimit  int     `json:"downLimit"`
	Mood       int     `json:"mood"`
}

type CnStockKData struct {
	Symbol     string      `json:"symbol"`
	Open       json.Number `json:"open"`
	High       json.Number `json:"high"`
	Low        json.Number `json:"low"`
	Close      json.Number `json:"close"`
	Volume     json.Number `json:"volume"`
	MaPrice5   float32     `json:"ma_price5"`
	MaVolume5  int64       `json:"ma_volume5"`
	MaPrice10  float32     `json:"ma_price10"`
	MaVolume10 int64       `json:"ma_volume10"`
	MaPrice30  float32     `json:"ma_price30"`
	MaVolume30 int64       `json:"ma_volume30"`
	Day        string      `json:"day"`
}

type DashData struct {
	Daily   []*DashDaily    `json:"daily"`
	SZDaily []*CnStockKData `json:"sz_daily"`
	SHDaily []*CnStockKData `json:"sh_daily"`
	HSDaily []*CnStockKData `json:"hs_daily"`
}

type DashDailyResponse struct {
	Code int       `json:"code"`
	Msg  string    `json:"msg"`
	Data *DashData `json:"data"`
}

type UpDownItem struct {
	Name     string  `json:"name"`
	Symbol   string  `json:"symbol"`
	Close    float32 `json:"close"`
	PreClose float32 `json:"preClose"`
	PctChg   float32 `json:"pctChange"`
	Day      string  `json:"day"`
	Flag     int     `json:"flag"`
}

type UpDownResponse struct {
	Code int           `json:"code"`
	Msg  string        `json:"msg"`
	Data []*UpDownItem `json:"data"`
}

type HotItem struct {
	Name   string  `json:"name"`
	Symbol string  `json:"symbol"`
	Vol    float64 `json:"vol"`
	Open   float64 `json:"open"`
	Close  float64 `json:"close"`
	Day    string  `json:"day"`
}

type GetHotResponse struct {
	Code int        `json:"code"`
	Msg  string     `json:"msg"`
	Data []*HotItem `json:"data"`
}

type GetDailyResponse struct {
	Code int        `json:"code"`
	Msg  string     `json:"msg"`
	Data []*CnDaily `json:"data"`
}

type UserPwdReq struct {
	Name  string `json:"name" binding:"required"`
	Pwd   string `json:"pwd" binding:"required"`
	Noice string `json:"noice" binding:"required"`
}

type CommonResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

type TradingReq struct {
	Stock   string  `json:"stock" binding:"required"`
	Price   float32 `json:"price" binding:"required"`
	Vol     int     `json:"vol" binding:"required"`
	OrderId string  `json:"orderId" binding:"required"`
}

type TradingDto struct {
	Name       string `json:"name"`
	Symbol     string `json:"symbol"`
	OrderPrice string `json:"orderPrice"`
	BuyPrice   string `json:"buyPrice"`
	SellPrice  string `json:"sellPrice"`
	CurePrice  string `json:"price"`
	Status     string `json:"status"`
	OrderDate  string `json:"orderDate"`
	BuyDate    string `json:"buyDate"`
	SellDate   string `json:"sellDate"`
	Flag       string `json:"flag"`
}

type TradingStatDto struct {
	Vol           int           `json:"vol"`
	Amount        float64       `json:"amount"`
	Pnl           float64       `json:"pnl"`
	FailedOrders  int           `json:"failed"`
	SuccessOrders int           `json:"success"`
	BuyOrders     int           `json:"buy"`
	CancelOrders  int           `json:"cancel"`
	Orders        []*TradingDto `json:"orders"`
}

type TradingStatRsp struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data *TradingStatDto `json:"data"`
}

type CnStockDaily struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Open      float32 `json:"open"`
	PreClose  float32 `json:"preClose"`
	CurePrice float32 `json:"curePrice"`
	High      float32 `json:"high"`
	Low       float32 `json:"low"`
	BuyPrice  float32 `json:"buyPrice"`
	SellPrice float32 `json:"sellPrice"`
	Vol       int     `json:"vol"`
	Amount    float64 `json:"amount"`
	Buy1Vol   int     `json:"buy_1_vol"`
	Buy1Price float64 `json:"buy_1_price"`
	Buy2Vol   int     `json:"buy_2_vol"`
	Buy2Price float64 `json:"buy_2_price"`
	Buy3Vol   int     `json:"buy_3_vol"`
	Buy3Price float64 `json:"buy_3_price"`
	Buy4Vol   int     `json:"buy_4_vol"`
	Buy4Price float64 `json:"buy_4_price"`
	Buy5Vol   int     `json:"buy_5_vol"`
	Buy5Price float64 `json:"buy_5_price"`

	Sell1Vol   int     `json:"sell_1_vol"`
	Sell1Price float64 `json:"sell_1_price"`
	Sell2Vol   int     `json:"sell_2_vol"`
	Sell2Price float64 `json:"sell_2_price"`
	Sell3Vol   int     `json:"sell_3_vol"`
	Sell3Price float64 `json:"sell_3_price"`
	Sell4Vol   int     `json:"sell_4_vol"`
	Sell4Price float64 `json:"sell_4_price"`
	Sell5Vol   int     `json:"sell_5_vol"`
	Sell5Price float64 `json:"sell_5_price"`
	Date       string  `json:"date"`
	Time       string  `json:"time"`
	Status     string  `json:"status"`
}

type HqForwardItem struct {
	Day     string `json:"day"`
	Total   int    `json:"total"`
	Success int    `json:"success"`
	Failed  int    `json:"failed"`
}

type HqCommonRsp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}
