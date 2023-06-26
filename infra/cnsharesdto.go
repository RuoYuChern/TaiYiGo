package infra

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

var (
	gIsCnEmpty = CnEmptyError{}
)

type CnEmptyError struct {
}

func (ee CnEmptyError) Error() string {
	return "CnEmptyError"
}
