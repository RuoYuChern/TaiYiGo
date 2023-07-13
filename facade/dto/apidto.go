package dto

type CnAdminCmd struct {
	Opt string `json:"opt"`
}

type SymbolDaily struct {
	Day   string  `json:"day"`
	Open  float64 `json:"open"`
	Close float64 `json:"close"`
	High  float64 `json:"high"`
	Low   float64 `json:"low"`
	Vol   float64 `json:"vol"`
	Mtm   float64 `json:"mtn"`
	Hld   float64 `json:"hld"`
	LSma  float64 `json:"lsma"`
	SSma  float64 `json:"ssma"`
}

type SymbolTrendResponse struct {
	Code int            `json:"code"`
	Msg  string         `json:"msg"`
	Data []*SymbolDaily `json:"data"`
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
