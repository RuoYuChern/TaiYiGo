package brain

import (
	"strings"
	"time"

	"taiyigo.com/common"
	"taiyigo.com/facade/tstock"
	"taiyigo.com/infra"
)

type loadActor struct {
	common.Actor
}

type loadCnBasic struct {
	common.Actor
}

type deltaLoadCnActor struct {
	common.Actor
}

func (la loadActor) Action() {
	today := common.GetDay(common.YYYYMMDD, time.Now())
	lastDay, err := infra.GetByKey(infra.CONF_TABLE, infra.KEY_DELTA)
	if err == nil && strings.Compare(today, lastDay) == 0 {
		common.Logger.Infof("Delta load has done")
		return
	}
	infra.SetKeyValue(infra.CONF_TABLE, infra.KEY_DELTA, today)
	var act common.Actor = loadCnBasic{}
	act.Action()
	act = deltaLoadCnActor{}
	act.Action()
}

func (lcb loadCnBasic) Action() {
	cnList, err := infra.QueryCnShareBasic("", "L")
	if err != nil {
		common.Logger.Infof("load cnbasic failed:%s", err)
		return
	}

	err = infra.SaveCnBasic(cnList)
	if err != nil {
		common.Logger.Infof("SaveCnBasic failed:%s", err)
	}
	infra.LoadMemData()
	common.Logger.Infof("basic loading over")
}

func (dlc deltaLoadCnActor) Action() {
	common.Logger.Infof("delta loading ......")
	now := common.GetDay(common.YYYYMMDD, time.Now())
	lastDay, err := infra.GetByKey(infra.CONF_TABLE, infra.KEY_CNLOADHISTORY)
	if err != nil {
		common.Logger.Infof("GetByKey failed: %s", err)
		return
	}
	if now == lastDay {
		common.Logger.Infof("delta loading done")
		return
	}
	if common.TodayIsWeek() {
		common.Logger.Infof("TodayIsWeek")
		return
	}
	cnList := &tstock.CnBasicList{}
	err = infra.GetCnBasic(cnList)
	if err != nil {
		common.Logger.Infof("GetCnBasic failed:%s", err)
		return
	}
	// 时间推后一天
	lastDay, _ = common.GetNextDay(lastDay)
	common.Logger.Infof("delta loading between [%s, %s], cnList:%d", lastDay, now, len(cnList.CnBasicList))
	cnShareStatus := make(map[string]string)
	timeStart := time.Now()
	rangeTotal, err := LoadSymbolDaily(cnList, lastDay, now, cnShareStatus)
	if err != nil {
		common.Logger.Infof("delta load failed:%s", err)
		return
	}
	timeUsed := time.Since(timeStart)
	common.Logger.Infof("delta loading between [%s, %s], total:%d, timeUsed:%f sec", lastDay, now, rangeTotal, timeUsed.Seconds())
	if rangeTotal > 0 {
		if err := infra.BatchSetKeyValue(infra.CONF_TABLE, cnShareStatus); err != nil {
			common.Logger.Infof("BatchSetKeyValue failed:%s", err)
			return
		}
		if err := infra.SetKeyValue(infra.CONF_TABLE, infra.KEY_CNLOADHISTORY, now); err != nil {
			common.Logger.Infof("SetKeyValue failed:%s", err)
			return
		}
		// 触发分析
		GetBrain().Subscript(TOPIC_STF, &FlowStart{})
	}
	common.Logger.Infof("delta load over")
}
