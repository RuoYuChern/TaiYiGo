package brain

import (
	"google.golang.org/protobuf/proto"
	"taiyigo.com/brain/algor"
	"taiyigo.com/common"
	"taiyigo.com/facade/tsdb"
	"taiyigo.com/facade/tstock"
	"taiyigo.com/infra"
)

type FlowStart struct {
	common.Actor
}

func (fs *FlowStart) Action() {
	cnList := &tstock.CnBasicList{}
	err := infra.GetCnBasic(cnList)
	if err != nil {
		common.Logger.Infof("GetCnBasic failed: %s", err)
		return
	}
	lastDay, err := infra.GetByKey(infra.CONF_TABLE, infra.KEY_CNLOADHISTORY)
	if err != nil {
		common.Logger.Infof("GetByKey failed: %s", err)
		return
	}

	dayTime, err := common.ToDay(common.YYYYMMDD, lastDay)
	if err != nil {
		common.Logger.Infof("%s ToDay failed: %s", lastDay, err)
		return
	}
	common.Logger.Infof("Think for day:%s start", lastDay)
	algs := algor.GetAlgList()
	stfList := tstock.StfList{Numbers: 0, Stfs: make([]*tstock.StfInfo, 0)}
	for _, basic := range cnList.GetCnBasicList() {
		if common.FillteST(basic.Name) {
			continue
		}
		tql := infra.Gettsdb().OpenQuery(basic.Symbol)
		out, err := tql.GetPointN(uint64(dayTime.UnixMilli()), common.Conf.Brain.StfPriceCount)
		infra.Gettsdb().CloseQuery(tql)
		if err != nil {
			if !infra.IsEmpty(err) {
				common.Logger.Warnf("%s GetPointN failed:%s", basic.Symbol, err)
				break
			}
			continue
		}
		//转成candle
		candles := make([]*tstock.Candle, out.Len())
		off := 0
		for f := out.Front(); f != nil; f = f.Next() {
			candles[off] = &tstock.Candle{}
			dat := f.Value.(*tsdb.TsdbData)
			err = proto.Unmarshal(dat.Data, candles[off])
			if err != nil {
				common.Logger.Infof("Unmarshal:%s", err)
				return
			}
			off++
		}
		for front := algs.Front(); front != nil; front = front.Next() {
			think := front.Value.(algor.ThinkAlg)
			b, o := think.TAnalyze(candles)
			if b {
				stf := &tstock.StfInfo{Symbol: basic.Symbol, Status: "S", Name: basic.Name, Opt: o, Day: uint64(dayTime.UnixMilli())}
				stfList.Stfs = append(stfList.Stfs, stf)
				break
			}
		}
	}
	common.Logger.Infof("Think for day:%s over, find:%d", lastDay, len(stfList.Stfs))
	if len(stfList.Stfs) > 0 {
		err = infra.SaveStfList("S", lastDay, &stfList)
		if err != nil {
			common.Logger.Infof("Save day %s failed:%s", lastDay, err)
		}
		GetBrain().Subscript(TOPIC_ADMIN, &MergeSTF{})
	}
}
