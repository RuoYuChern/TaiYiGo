package infra

import (
	"container/list"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"
	"taiyigo.com/common"
	"taiyigo.com/facade/dto"
	"taiyigo.com/facade/tsdb"
	"taiyigo.com/facade/tsorder"
	"taiyigo.com/facade/tstock"
)

type DashBoardDao struct {
	dashMon *tstock.DashBoardMonth
}

type HqCacheData struct {
	Symbol string
	Name   string
	Data   []*dto.CnStockKData
}

type memData struct {
	cnShares        map[string]*tstock.CnBasic
	cnSharesNameMap map[string]string
	orderMap        map[string]*tsorder.TOrder
	hqData          map[string]*HqCacheData
	lock            sync.Mutex
}

var gmemData *memData
var gmenOnce sync.Once

func LoadData() {
	gmenOnce.Do(func() {
		gmemData = &memData{cnShares: make(map[string]*tstock.CnBasic), cnSharesNameMap: map[string]string{}, orderMap: make(map[string]*tsorder.TOrder)}
		gmemData.hqData = make(map[string]*HqCacheData)
		LoadMemData()
	})
}

func (dbd *DashBoardDao) Add(dbdv *tstock.DashBoardV1) {
	month := common.SubString(dbdv.Day, 0, 6)
	if dbd.dashMon != nil {
		if strings.Compare(month, dbd.dashMon.Mon) < 0 {
			return
		}
		if strings.Compare(month, dbd.dashMon.Mon) == 0 {
			dbd.dashMon.DailyDash = append(dbd.dashMon.DailyDash, dbdv)
			return
		}
		//先保存
		dbd.Save()
		dbd.dashMon = nil
	}
	dbd.dashMon = &tstock.DashBoardMonth{DailyDash: make([]*tstock.DashBoardV1, 0)}
	dbd.dashMon.Mon = month
	dbd.dashMon.DailyDash = append(dbd.dashMon.DailyDash, dbdv)
}

func (dao *DashBoardDao) Save() {
	if dao.dashMon == nil {
		return
	}
	oldDbd := &tstock.DashBoardMonth{}
	name := fmt.Sprintf("%s_dbd.dat", dao.dashMon.Mon)
	err := GetMsg(name, oldDbd)
	if err == nil {
		//合并老数据
		mergeDash := make([]*tstock.DashBoardV1, 0)
		newLen := len(dao.dashMon.DailyDash)
		oldLen := len(oldDbd.DailyDash)
		oldOff := 0
		common.Logger.Infof("merge old data:%s, oldLen:%d, newLen:%d", oldDbd.Mon, oldLen, newLen)
		for off := 0; off < newLen; off++ {
			nd := dao.dashMon.DailyDash[off]
			for oldOff < oldLen {
				od := oldDbd.DailyDash[oldOff]
				cmp := strings.Compare(nd.Day, od.Day)
				if cmp > 0 {
					// nd > od
					mergeDash = append(mergeDash, od)
					oldOff++
				} else {
					if cmp == 0 {
						oldOff++
					}
					break
				}
			} //end oldOff
			mergeDash = append(mergeDash, nd)
		}
		//
		for oldOff < oldLen {
			od := oldDbd.DailyDash[oldOff]
			mergeDash = append(mergeDash, od)
			oldOff++
		}
		mgrMonDash := &tstock.DashBoardMonth{Mon: dao.dashMon.Mon, DailyDash: mergeDash}
		common.Logger.Infof("merge data:%s, newLen:%d", mgrMonDash.Mon, len(mgrMonDash.DailyDash))
		err = SaveMsg(name, mgrMonDash)
	} else {
		common.Logger.Infof("save data:%s, newLen:%d", dao.dashMon.Mon, len(dao.dashMon.DailyDash))
		err = SaveMsg(name, dao.dashMon)
	}
	if err != nil {
		common.Logger.Warnf("save %s, failed:%s", name, err)
	}
}

func LoadMemData() {
	cnList := tstock.CnBasicList{}
	err := GetCnBasic(&cnList)
	if err != nil {
		common.Logger.Infof("GetCnBasic failed:%s", err)
		return
	}

	gmemData.cnShares = make(map[string]*tstock.CnBasic)
	gmemData.cnSharesNameMap = make(map[string]string)

	for _, v := range cnList.CnBasicList {
		gmemData.cnShares[v.Symbol] = v
		gmemData.cnSharesNameMap[v.Name] = v.Symbol
	}
	orderList, err := Scan(ORDER_TABLE)
	if err != nil {
		common.Logger.Infof("Scan %s, failed:%s", ORDER_TABLE, err)
		return
	}

	gmemData.orderMap = make(map[string]*tsorder.TOrder)
	for f := orderList.Front(); f != nil; f = f.Next() {
		order := &tsorder.TOrder{}
		kv := f.Value.(*KvPair)
		err = proto.Unmarshal(kv.Value, order)
		if err != nil {
			common.Logger.Infof("Unmarshal %s, failed:%s", kv.Key, err)
			continue
		}
		gmemData.orderMap[order.OrderId] = order
	}
	gmemData.hqData = make(map[string]*HqCacheData)
	symbols := [4]string{"sz399001", "sz399300", "sh000001", "sh000300"}
	names := [4]string{"深圳成指", "沪深300", "上证指数", "沪深300"}
	for i := 0; i < 4; i++ {
		data, err := GetCnKData(symbols[i], "", 240, 200)
		if err != nil {
			common.Logger.Infof("Get %s failed:%s", symbols[i], err)
		}
		gmemData.hqData[symbols[i]] = &HqCacheData{Symbol: symbols[i], Name: names[i], Data: data}
	}
	common.Logger.Infof("cnShares.Size:%d, order.Size:%d, hq:%d", len(gmemData.cnShares), len(gmemData.orderMap), len(gmemData.hqData))
}

func GetKDataCache(symbol string) *HqCacheData {
	hq, ok := gmemData.hqData[symbol]
	if ok {
		return hq
	} else {
		return nil
	}
}

func GetLastNMonthDash(month int) []*tstock.DashBoardMonth {
	dir := fmt.Sprintf("%s/meta", common.Conf.Infra.FsDir)
	fsList, err := common.GetFileList(dir, "dbd.dat", "hdbd.dat", month)
	if err != nil {
		common.Logger.Warnf("GetFileList:%s", err)
		return nil
	}
	dbm := make([]*tstock.DashBoardMonth, fsList.Len())
	off := 0
	for f := fsList.Front(); f != nil; f = f.Next() {
		dm := &tstock.DashBoardMonth{}
		err = GetMsg(f.Value.(string), dm)
		if err != nil {
			common.Logger.Infof("read %s failed:%s", f.Value.(string), err)
			return nil
		}
		dbm[off] = dm
		off++
	}
	return dbm
}

func GetLastDayDash(day string) *tstock.DashBoardV1 {
	month := common.SubString(day, 0, 6)
	name := fmt.Sprintf("%s_dbd.dat", month)
	dm := &tstock.DashBoardMonth{}
	err := GetMsg(name, dm)
	if err != nil {
		common.Logger.Infof("read %s failed:%s", name, err)
		return nil
	}
	for _, v := range dm.DailyDash {
		if v.Day == day {
			return v
		}
	}
	return nil
}

func GetSymbolName(symbol string) string {
	n, ok := gmemData.cnShares[symbol]
	if ok {
		return n.Name
	}
	return ""
}

func GetNameSymbol(name string) string {
	symbol, ok := gmemData.cnSharesNameMap[name]
	if ok {
		return symbol
	}
	return ""
}

func AddOrder(order *tsorder.TOrder) {
	gmemData.lock.Lock()
	gmemData.orderMap[order.OrderId] = order
	gmemData.lock.Unlock()
}

func GetOrder(orderId string) *tsorder.TOrder {
	gmemData.lock.Lock()
	order, ok := gmemData.orderMap[orderId]
	gmemData.lock.Unlock()
	if ok {
		return order
	} else {
		return nil
	}
}

func GetOrdersByYear(year string) (*list.List, error) {
	orders := list.New()
	gmemData.lock.Lock()
	for _, v := range gmemData.orderMap {
		if !strings.HasPrefix(v.CreatDay, year) {
			continue
		}
		orders.PushBack(v)
	}
	gmemData.lock.Unlock()
	return orders, nil
}

func GetUnFinishOrders() *list.List {
	orders := list.New()
	gmemData.lock.Lock()
	for _, v := range gmemData.orderMap {
		if v.Status == dto.ORDER_IDLE || v.Status == dto.ORDER_BUY {
			orders.PushBack(v)
		}
	}
	gmemData.lock.Unlock()
	return orders
}

func GetDayBetween(symbol, low, high string, offset int) (*list.List, error) {
	start, _ := common.ToDay(common.YYYYMMDD, low)
	end, _ := common.ToDay(common.YYYYMMDD, high)
	tsd := Gettsdb()
	tql := tsd.OpenQuery(symbol)
	datList, err := tql.GetRange(uint64(start.UnixMilli()), uint64(end.UnixMilli()), offset)
	tsd.CloseQuery(tql)
	tsd.Close()
	if err != nil {
		return nil, err
	}
	candleList := list.New()
	for front := datList.Front(); front != nil; front = front.Next() {
		candle := &tstock.Candle{}
		value := front.Value.(*tsdb.TsdbData)
		err = proto.Unmarshal(value.Data, candle)
		if err != nil {
			common.Logger.Warnf("Unmarshal failed:%s", err)
			return nil, err
		}
		candleList.PushBack(candle)
	}
	return candleList, nil
}

func GetSymbolNPoint(symbol, date string, n int) ([]*tstock.Candle, error) {
	lastTime, err := common.ToDay(common.YYYYMMDD, date)
	if err != nil {
		return nil, err
	}
	tsd := Gettsdb()
	tql := tsd.OpenQuery(symbol)
	datList, err := tql.GetPointN(uint64(lastTime.UnixMilli()), n)
	tsd.CloseQuery(tql)
	tsd.Close()
	if err != nil {
		return nil, err
	}
	items := make([]*tstock.Candle, datList.Len())
	off := 0
	for front := datList.Front(); front != nil; front = front.Next() {
		items[off] = &tstock.Candle{}
		value := front.Value.(*tsdb.TsdbData)
		err = proto.Unmarshal(value.Data, items[off])
		if err != nil {
			common.Logger.Warnf("Unmarshal failed:%s", err)
			return nil, err
		}
		off++
	}
	return items, err
}

func SaveCnBasic(basics *list.List) error {
	cnList := &tstock.CnBasicList{Numbers: int32(basics.Len()), CnBasicList: make([]*tstock.CnBasic, basics.Len())}
	off := 0
	for front := basics.Front(); front != nil; front = front.Next() {
		share := front.Value.(*CnSharesBasic)
		cnb := &tstock.CnBasic{}
		cnb.Symbol = share.Symbol
		cnb.Name = share.Name
		cnb.Area = share.Area
		cnb.Industry = share.Industry
		cnb.FulName = share.FulName
		cnb.EnName = share.EnName
		cnb.CnName = share.CnName
		cnb.Market = share.Market
		cnb.ExChange = share.ExChange
		cnb.Status = share.Status
		cnb.ListDate = share.ListDate
		cnb.DelistDate = share.DelistDate
		cnb.IsHs = share.IsHs
		cnList.CnBasicList[off] = cnb
		off++
	}

	return SaveMsg("cnbasic.dat", cnList)
}

func SaveStfList(status string, day string, msg proto.Message) error {
	name := fmt.Sprintf("%s_%s_stf.dat", day, status)
	return SaveMsg(name, msg)
}

func GetStfList(status string, day string, msg proto.Message) error {
	name := fmt.Sprintf("%s_%s_stf.dat", day, status)
	err := GetMsg(name, msg)
	return err
}

func SaveForwardRecord(mon string, record *tstock.ForwardStatRecord) error {
	name := fmt.Sprintf("%s_fwd.dat", mon)
	old := &tstock.ForwardStatRecord{}
	err := GetMsg(name, old)
	if err != nil {
		err = SaveMsg(name, record)
	} else {
		old.Items = append(old.Items, record.Items...)
		err = SaveMsg(name, old)
	}
	return err
}

func GetForwardRecord(mon string, msg proto.Message) error {
	name := fmt.Sprintf("%s_fwd.dat", mon)
	err := GetMsg(name, msg)
	return err
}

func SaveStfRecord(msg proto.Message) error {
	name := "normal_S_stf.dat"
	return SaveMsg(name, msg)
}

func GetStfRecord(msg proto.Message) error {
	name := "normal_S_stf.dat"
	return GetMsg(name, msg)
}

func GetCnBasic(cnList *tstock.CnBasicList) error {
	err := GetMsg("cnbasic.dat", cnList)
	return err
}

func GetMsg(name string, msg proto.Message) error {
	tsf := tsFile{flag: os.O_RDONLY}
	err := tsf.read(name, msg)
	return err
}

func RemoveMsg(name string) {
	fn := fmt.Sprintf("%s/meta/%s", common.Conf.Infra.FsDir, name)
	os.Remove(fn)
}

func SaveMsg(name string, msg proto.Message) error {
	tsf := tsFile{flag: os.O_CREATE | os.O_WRONLY | os.O_TRUNC}
	err := tsf.write(name, msg)
	if err != nil {
		common.Logger.Warnf("SaveMsg:%s failed:%s", name, err)
	}
	return err
}

func GetDLog(name string, cb func(*tstock.StockDaily) error) error {
	parts := strings.SplitN(name, ".", 2)
	tdlg := tsDataLog{}
	if err := tdlg.open(parts[0], os.O_RDONLY); err != nil {
		common.Logger.Infof("open %s failed:%s", name, err)
		return err
	}
	hb := make([]byte, 4)
	var err error
	for {
		err = tdlg.read(hb)
		if err != nil {
			break
		}
		ol := int(GetIntFromB(hb))
		if ol < 0 || ol >= (1<<20) {
			tdlg.close()
			err = fmt.Errorf("ol =%d is error", ol)
			break
		}
		obf := make([]byte, ol)
		err = tdlg.read(obf)
		if err != nil {
			break
		}

		stdl := &tstock.StockDaily{}
		err = proto.Unmarshal(obf, stdl)
		if err != nil {
			break
		}
		err = cb(stdl)
		if err != nil {
			break
		}
	}
	tdlg.close()
	if isTargetError(err, io.EOF) {
		return nil
	}
	common.Logger.Infof("Get %s failed:%s", name, err)
	return err
}
