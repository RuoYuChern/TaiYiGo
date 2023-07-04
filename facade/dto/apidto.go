package dto

type CnAdminCmd struct {
	Opt string `json:"opt"`
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
