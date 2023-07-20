package brain

import (
	"fmt"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"
	"taiyigo.com/common"
	"taiyigo.com/facade/tsdb"
	"taiyigo.com/facade/tstock"
	"taiyigo.com/indicators"
	"taiyigo.com/infra"
)

type MergeSTF struct {
	common.Actor
}

type dashBoardMerge struct {
	common.Actor
}

type MergeAll struct {
	common.Actor
}

type JustifyStat struct {
	common.Actor
	StartDay string
}

func (ma *MergeAll) Action() {
	act := dashBoardMerge{}
	act.Action()
}

func (jst *JustifyStat) Action() {
	common.Logger.Infof("JustifyStat start ")
	cnList := &tstock.CnBasicList{}
	err := infra.GetCnBasic(cnList)
	if err != nil {
		common.Logger.Infof("GetCnBasic failed:%s", err)
		return
	}
	ndbc := indicators.NewNDbc()
	isLogger := false
	for _, basic := range cnList.CnBasicList {
		cnShareLastDay, err := infra.GetByKey(infra.CONF_TABLE, basic.Symbol)
		if err != nil {
			continue
		}
		if strings.Compare(cnShareLastDay, jst.StartDay) <= 0 {
			continue
		}

		start, _ := common.ToDay(common.YYYYMMDD, jst.StartDay)
		end, _ := common.ToDay(common.YYYYMMDD, cnShareLastDay)
		tsql := infra.Gettsdb().OpenQuery(basic.Symbol)
		dlist, err := tsql.GetRange(uint64(start.UnixMilli()), uint64(end.UnixMilli()), 0)
		infra.Gettsdb().CloseQuery(tsql)
		if err != nil {
			common.Logger.Infof("Symbol %s,between [%s, %s] error:%s", basic.Symbol, jst.StartDay, cnShareLastDay, err.Error())
			continue
		}
		common.Logger.Debugf("Symbol:%s, between [%s, %s], total:%d", basic.Symbol, jst.StartDay, cnShareLastDay, dlist.Len())
		filter := make(map[string]bool, 0)
		for f := dlist.Front(); f != nil; f = f.Next() {
			candle := &tstock.Candle{}
			value := f.Value.(*tsdb.TsdbData)
			err = proto.Unmarshal(value.Data, candle)
			if err != nil {
				common.Logger.Warnf("Unmarshal failed:%s", err)
				return
			}
			period := time.Unix(int64(candle.Period/1000), 0)
			day := common.GetDay(common.YYYYMMDD, period)
			daySymbol := fmt.Sprintf("%s.%s", day, basic.Symbol)
			_, ok := filter[daySymbol]
			if !ok {
				ndbc.Cal(day, basic.Symbol, candle)
				filter[daySymbol] = true
			} else {
				if !isLogger {
					common.Logger.Infof("%s is exist", daySymbol)
					isLogger = true
				}
			}
		}
	}
	ndbc.Save()
	common.Logger.Infof("JustifyStat over")
}

func (dbm *dashBoardMerge) Action() {
	common.Logger.Infof("merge dbm start...")
	dir := fmt.Sprintf("%s/meta", common.Conf.Infra.FsDir)
	fsList, err := common.GetFileList(dir, "dbd.dat", "hdbd.dat", 100)
	if err != nil {
		common.Logger.Warnf("GetFileList failed:%s", err)
		return
	}
	var dby *tstock.DashBoardYear
	curYear := common.GetYear(time.Now())
	for f := fsList.Front(); f != nil; f = f.Next() {
		fn := f.Value.(string)
		if strings.HasPrefix(fn, curYear) {
			break
		}
		dbm := &tstock.DashBoardMonth{}
		year := common.SubString(fn, 0, 4)
		err = infra.GetMsg(fn, dbm)
		if err != nil {
			common.Logger.Warnf("GetMsg:%s, failed:%s", fn, err)
			continue
		}
		if dby != nil {
			if strings.Compare(year, dby.Year) == 0 {
				dby.MonthDash = append(dby.MonthDash, dbm)
			} else {
				hn := fmt.Sprintf("%s_hdbd.dat", dby.Year)
				infra.SaveMsg(hn, dby)
				dby = &tstock.DashBoardYear{Year: year, MonthDash: make([]*tstock.DashBoardMonth, 0, 12)}
				dby.MonthDash = append(dby.MonthDash, dbm)
			}
		} else {
			dby = &tstock.DashBoardYear{Year: year, MonthDash: make([]*tstock.DashBoardMonth, 0, 12)}
			dby.MonthDash = append(dby.MonthDash, dbm)
		}
		//remove old file
		infra.RemoveMsg(fn)
	}
	if dby != nil {
		hn := fmt.Sprintf("%s_hdbd.dat", dby.Year)
		infra.SaveMsg(hn, dby)
	}
	common.Logger.Infof("merge dbm over")
}

func (mgs *MergeSTF) Action() {
	common.Logger.Infof("merge stf start...")
	dir := fmt.Sprintf("%s/meta", common.Conf.Infra.FsDir)
	fsList, err := common.GetFileList(dir, "S_stf.dat", "normal_S_stf.dat", 10)
	if err != nil {
		return
	}
	symbolMap := make(map[string]*tstock.StfRecord)
	highDay := ""
	for f := fsList.Front(); f != nil; f = f.Next() {
		parts := strings.SplitN(f.Value.(string), "_", 3)
		day := parts[0]
		stfList := tstock.StfList{}
		err := infra.GetStfList("S", day, &stfList)
		if err != nil {
			common.Logger.Warnf("Get day %s stflist failed:%s", day, err)
			continue
		}
		if highDay != "" {
			highDay = day
		} else if strings.Compare(highDay, day) < 0 {
			highDay = day
		}

		for _, v := range stfList.Stfs {
			str, ok := symbolMap[v.Symbol]
			if ok {
				if strings.Compare(day, str.HighDay) > 0 {
					str.HighDay = day
				}
				if strings.Compare(day, str.LowDay) < 0 {
					str.LowDay = day
				}
			} else {
				str = &tstock.StfRecord{Symbol: v.Symbol, Status: v.Status, Opt: v.Opt, HighDay: day, LowDay: day}
				symbolMap[str.Symbol] = str
			}
		}
	}
	num := int32(len(symbolMap))
	common.Logger.Infof("merge highday:%s, stf len:%d", highDay, num)
	hd, _ := common.ToDay(common.YYYYMMDD, highDay)
	strList := tstock.StfRecordList{Day: uint64(hd.UnixMilli()), Numbers: num, Stfs: make([]*tstock.StfRecord, num)}
	off := 0
	for _, v := range symbolMap {
		strList.Stfs[off] = v
		off++
	}
	err = infra.SaveStfRecord(&strList)
	if err != nil {
		common.Logger.Warnf("merge stf error:%s", err)
	}
	common.Logger.Infof("merge stf over")
}
