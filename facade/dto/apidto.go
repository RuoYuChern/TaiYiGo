package dto

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

type DashDailyResponse struct {
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Data []*DashDaily `json:"data"`
}

type UpDownItem struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
	Day    string `json:"day"`
	Flag   int    `json:"flag"`
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
