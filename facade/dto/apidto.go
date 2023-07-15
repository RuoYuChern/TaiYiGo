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
