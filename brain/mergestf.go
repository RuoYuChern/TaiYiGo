package brain

import (
	"container/list"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"taiyigo.com/common"
	"taiyigo.com/facade/tstock"
	"taiyigo.com/infra"
)

type MergeSTF struct {
	common.Actor
}

func (mgs *MergeSTF) Action() {
	common.Logger.Infof("merge stf start...")
	dir := fmt.Sprintf("%s/meta", common.Conf.Infra.FsDir)
	fsList := list.New()
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), "S_stf.dat") || (strings.Compare(d.Name(), "normal_S_stf.dat") == 0) {
			return nil
		}
		parts := strings.SplitN(d.Name(), "_", 3)
		if common.IsDayBeforN(common.YYYYMMDD, parts[0], 10) {
			return nil
		}
		fsList.PushBack(parts[0])
		return nil
	})
	if err != nil {
		return
	}

	symbolMap := make(map[string]*tstock.StfRecord)
	highDay := ""
	for f := fsList.Front(); f != nil; f = f.Next() {
		day := f.Value.(string)
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
