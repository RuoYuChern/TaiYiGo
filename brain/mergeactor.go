package brain

import (
	"fmt"
	"strings"
	"time"

	"taiyigo.com/common"
	"taiyigo.com/facade/tstock"
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

func (ma *MergeAll) Action() {
	act := dashBoardMerge{}
	act.Action()
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
		common.Logger.Infof("Fn:%s", fn)
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
