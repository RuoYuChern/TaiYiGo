package infra

import (
	"container/list"
	"fmt"
	"os"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"
	"taiyigo.com/common"
	"taiyigo.com/facade/tsdb"
	"taiyigo.com/facade/tstock"
)

type DashBoardDao struct {
	dashMon *tstock.DashBoardMonth
}

type memData struct {
	cnShares        map[string]*tstock.CnBasic
	cnSharesNameMap map[string]string
}

var gmemData *memData
var gmenOnce sync.Once

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
	}
	dbd.dashMon = &tstock.DashBoardMonth{DailyDash: make([]*tstock.DashBoardV1, 0)}
	dbd.dashMon.Mon = month
	dbd.dashMon.DailyDash = append(dbd.dashMon.DailyDash, dbdv)
}

func (dbd *DashBoardDao) Save() {
	if dbd.dashMon == nil {
		return
	}
	tsf := tsFile{flag: os.O_RDONLY}
	oldDbd := &tstock.DashBoardMonth{}
	name := fmt.Sprintf("%s_dbd.dat", dbd.dashMon.Mon)
	err := tsf.read(name, oldDbd)
	if err == nil {
		//合并老数据
		dailyDash := make([]*tstock.DashBoardV1, 0, 30)
		nl := len(dbd.dashMon.DailyDash)
		ol := len(oldDbd.DailyDash)
		oldOff := 0
		for off := 0; off < nl; off++ {
			nd := dbd.dashMon.DailyDash[off]
			for oldOff < ol {
				od := oldDbd.DailyDash[oldOff]
				cmp := strings.Compare(nd.Day, od.Day)
				if cmp <= 0 {
					//nd <= od
					dailyDash = append(dailyDash, nd)
					if cmp == 0 {
						oldOff++
					}
					break
				} else {
					// nd > od
					dailyDash = append(dailyDash, od)
					oldOff++
				}
			}
		}
		//
		for oldOff < ol {
			od := oldDbd.DailyDash[oldOff]
			dailyDash = append(dailyDash, od)
			oldOff++
		}
		dbd.dashMon.DailyDash = dailyDash
	}
	tsf = tsFile{flag: os.O_CREATE}
	err = tsf.write(name, dbd.dashMon)
	if err != nil {
		common.Logger.Warnf("save %s, failed:%s", name, err)
	}
}

func LoadMemData() {
	gmemData = &memData{cnShares: make(map[string]*tstock.CnBasic), cnSharesNameMap: map[string]string{}}
	cnList := tstock.CnBasicList{}
	err := GetCnBasic(&cnList)
	if err != nil {
		common.Logger.Infof("GetCnBasic failed:%s", err)
		return
	}
	for _, v := range cnList.CnBasicList {
		gmemData.cnShares[v.Symbol] = v
		gmemData.cnSharesNameMap[v.Name] = v.Symbol
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
		ts := tsFile{flag: os.O_RDONLY}
		err := ts.read(f.Value.(string), dm)
		if err != nil {
			common.Logger.Infof("read %s failed:%s", f.Value.(string), err)
			continue
		}
		dbm[off] = dm
		off++
	}
	return dbm
}

func GetSymbolName(symbol string) string {
	gmenOnce.Do(LoadMemData)
	n, ok := gmemData.cnShares[symbol]
	if ok {
		return n.Name
	}
	return ""
}

func GetNameSymbol(name string) string {
	gmenOnce.Do(LoadMemData)
	symbol, ok := gmemData.cnSharesNameMap[name]
	if ok {
		return symbol
	}
	return ""
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
		share := front.Value.(*TjCnBasicInfo)
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
	tsf := tsFile{flag: os.O_CREATE}
	err := tsf.write(name, msg)
	if err != nil {
		common.Logger.Warnf("SaveMsg:%s failed:%s", name, err)
	}
	return err
}
