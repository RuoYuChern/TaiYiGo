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

func GetLastNDayDash(day int, up bool) []*tstock.DashBoardV1 {
	month := (day / 30) + 1
	dir := fmt.Sprintf("%s/meta", common.Conf.Infra.FsDir)
	fsList, err := common.GetFileList(dir, "dbd.dat", "hdbd.dat", month)
	if err != nil {
		common.Logger.Warnf("GetFileList:%s", err)
		return nil
	}
	dList := list.New()
	for t := fsList.Back(); t != nil; t = t.Prev() {
		dm := &tstock.DashBoardMonth{}
		err = GetMsg(t.Value.(string), dm)
		if err != nil {
			common.Logger.Infof("read %s failed:%s", t.Value.(string), err)
			return nil
		}
		off := len(dm.DailyDash) - 1
		for off >= 0 {
			dList.PushFront(dm.DailyDash[off])
			off -= 1
			if dList.Len() == day {
				break
			}
		}
		if dList.Len() == day {
			break
		}
	}
	if dList.Len() == 0 {
		return nil
	}
	dbdvs := make([]*tstock.DashBoardV1, dList.Len())
	off := 0
	if up {
		for f := dList.Front(); f != nil; f = f.Next() {
			dbdvs[off] = f.Value.(*tstock.DashBoardV1)
			off++
		}
	} else {
		for f := dList.Back(); f != nil; f = f.Prev() {
			dbdvs[off] = f.Value.(*tstock.DashBoardV1)
			off++
		}
	}
	return dbdvs
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
