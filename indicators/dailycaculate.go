package indicators

import (
	"strings"

	"taiyigo.com/common"
	"taiyigo.com/facade/tstock"
	"taiyigo.com/infra"
)

var (
	UPLIMIT_LEVEL   = 9.5
	DOWNLIMIT_LEVEL = -9.5
	DAY_TOP_TOTAL   = 50
)

type DayBoradCal struct {
	dbdv *tstock.DashBoardV1
	hp   *common.LimitedHeap
}

type NameDashBoradCal struct {
	dbcm map[string]*DayBoradCal
}

func NewNDbc() *NameDashBoradCal {
	return &NameDashBoradCal{dbcm: make(map[string]*DayBoradCal)}
}

func (ndbc *NameDashBoradCal) Cal(day string, symbol string, candle *tstock.Candle) {
	dbc, ok := ndbc.dbcm[day]
	if !ok {
		dbc = &DayBoradCal{dbdv: newDashBoardV1(day)}
		ndbc.dbcm[day] = dbc
	}
	dbc.Cal(day, symbol, candle)
}

func (ndbc *NameDashBoradCal) Save() {
	size := len(ndbc.dbcm)
	if size == 0 {
		return
	}
	lsp := common.NewLp(size+1, func(a1, a2 any) int {
		v1 := a1.(string)
		v2 := a2.(string)
		return strings.Compare(v1, v2)
	})
	for _, v := range ndbc.dbcm {
		v.SetTop()
		lsp.Add(v.dbdv.Day)
	}
	dao := &infra.DashBoardDao{}
	for {
		v := lsp.Top()
		if v == nil {
			break
		}
		db := ndbc.dbcm[v.(string)]
		dao.Add(db.dbdv)
	}
	ndbc.dbcm = make(map[string]*DayBoradCal)
	dao.Save()
}

func newDashBoardV1(day string) *tstock.DashBoardV1 {
	dbdv := &tstock.DashBoardV1{Day: day, Top20Vol: make([]*tstock.TopSymbol, DAY_TOP_TOTAL)}
	dbdv.DownLimit = make([]*tstock.SymbolRate, 0)
	dbdv.UpLimit = make([]*tstock.SymbolRate, 0)
	return dbdv
}

func (dbc *DayBoradCal) SetTop() {
	off := 0
	for {
		v := dbc.hp.Top()
		if v == nil {
			break
		}
		dbc.dbdv.Top20Vol[off] = v.(*tstock.TopSymbol)
		off++
	}
}

func (dbc *DayBoradCal) Cal(day string, symbol string, candle *tstock.Candle) {
	symbolRate := &tstock.SymbolRate{Symbol: symbol, Close: candle.Close, PreClose: candle.PreClose, Rate: candle.Pcgp}
	if candle.Pcg < 0 {
		dbc.dbdv.DownStocks = dbc.dbdv.DownStocks + 1
		if candle.Pcgp <= DOWNLIMIT_LEVEL {
			dbc.dbdv.DownLimit = append(dbc.dbdv.DownLimit, symbolRate)
		}
	} else if candle.Pcg > 0 {
		dbc.dbdv.UpStocks = dbc.dbdv.UpStocks + 1
		if candle.Pcgp >= UPLIMIT_LEVEL {
			dbc.dbdv.UpLimit = append(dbc.dbdv.UpLimit, symbolRate)
		}
	}
	dbc.dbdv.TotalAmount += candle.Amount
	dbc.dbdv.TotalVol += float64(candle.Volume)
	dbc.dbdv.Stocks += 1
	nv := &tstock.TopSymbol{Name: symbol, Vol: float64(candle.Volume), Open: candle.Open, Close: candle.Close}
	if dbc.hp == nil {
		dbc.hp = common.NewLp(DAY_TOP_TOTAL, func(a1, a2 any) int {
			v1 := int64(a1.(*tstock.TopSymbol).Vol)
			v2 := int64(a2.(*tstock.TopSymbol).Vol)
			diff := int(v1 - v2)
			return diff
		})
	}
	dbc.hp.AddAndTrace(nv, func(flag string, v, old any) {})
}
