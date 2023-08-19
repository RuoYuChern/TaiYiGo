package api

import (
	"fmt"
	"strings"
	"time"

	"github.com/tao/faststore"
	"github.com/tao/faststore/api"
	"google.golang.org/protobuf/proto"
	"taiyigo.com/brain"
	"taiyigo.com/common"
	"taiyigo.com/facade/dto"
	"taiyigo.com/facade/tsdb"
	"taiyigo.com/facade/tstock"
	"taiyigo.com/infra"
)

type yearItems struct {
	start string
	end   string
}
type loadHistoryActor struct {
	common.Actor
	cmd *dto.CnAdminCmd
}

type loadCnBasciActor struct {
	common.Actor
	cmd *dto.CnAdminCmd
}

type tsdbRedo struct {
	common.Actor
	tda *infra.TsdbAppender
}

type tsdbMigrate struct {
	common.Actor
}

func (actor *loadCnBasciActor) Action() {
	common.Logger.Infof("basic loading ......")
	if actor.cmd.Opt != "LOAD" {
		common.Logger.Infof("cmd:%s is error", actor.cmd.Opt)
		return
	}
	cnList, err := infra.GetBasicFromTj()
	if err != nil {
		common.Logger.Infof("do cmd:%s is failed:%s", actor.cmd.Opt, err)
		return
	}

	err = infra.SaveCnBasic(cnList)
	if err != nil {
		common.Logger.Infof("SaveCnBasic failed:%s", err)
	}
	common.Logger.Infof("basic loading over")
}

func (actor *loadHistoryActor) Action() {
	common.Logger.Infof("history loading ......")
	if actor.cmd.Opt != "LOAD" {
		common.Logger.Infof("cmd:%s is error", actor.cmd.Opt)
		return
	}
	now := common.Conf.Quotes.HistoryEnd
	if actor.cmd.Value != "FORCE" {
		b, err := infra.CheckAndSet(infra.CONF_TABLE, infra.KEY_CNLOADHISTORY, now)
		if err != nil {
			common.Logger.Infof("do cmd:%s is failed:%s", actor.cmd.Opt, err)
			return
		}
		if !b {
			common.Logger.Infof("do cmd:%s has done", actor.cmd.Opt)
			return
		}
	}

	cnList := &tstock.CnBasicList{}
	err := infra.GetCnBasic(cnList)
	if err != nil {
		common.Logger.Infof("GetCnBasic failed:%s", err)
		return
	}

	datRang := []yearItems{{"20200101", "20201231"}, {"20210101", "20211231"}, {"20220101", "20221231"}, {"20230101", now}}
	// datRang := []yearItems{{"20230101", now}}
	cnShareStatus := make(map[string]string)
	timeStart := time.Now()
	for _, v := range datRang {
		common.Logger.Infof("day between [%s, %s] started", v.start, v.end)
		rangeTotal, err := brain.LoadSymbolDaily(cnList, v.start, v.end, cnShareStatus)
		if err != nil {
			common.Logger.Warnf("LoadSymbolDaily failed:%s", err)
			return
		}
		common.Logger.Infof("day between [%s, %s] over, total:%d", v.start, v.end, rangeTotal)
	}
	timeUsed := time.Since(timeStart)
	if err := infra.BatchSetKeyValue(infra.CONF_TABLE, cnShareStatus); err != nil {
		common.Logger.Infof("BatchSetKeyValue failed:%s", err)
	}
	common.Logger.Infof("history load over, time Used:%f sec", timeUsed.Seconds())
}

func (tsd *tsdbRedo) Action() {
	common.Logger.Infof("redo start ......")
	dir := fmt.Sprintf("%s/dlog", common.Conf.Infra.FsDir)
	dlist, err := common.GetFileList(dir, ".dlog", ".dat", 50)
	if err != nil {
		common.Logger.Infof("redo failed:%s", err)
	}

	for f := dlist.Front(); f != nil; f = f.Next() {
		name := f.Value.(string)
		err = infra.GetDLog(name, func(sd *tstock.StockDaily) error {
			candle := infra.ToCandle2(sd)
			out, cberr := proto.Marshal(candle)
			if cberr != nil {
				return cberr
			}
			if tsd.tda == nil {
				tsd.tda = infra.Gettsdb().OpenAppender(sd.Symbol)
			} else {
				if strings.Compare(sd.Symbol, tsd.tda.GetId()) != 0 {
					infra.Gettsdb().CloseAppender(tsd.tda)
					tsd.tda = infra.Gettsdb().OpenAppender(sd.Symbol)
				}
			}
			tsData := &tsdb.TsdbData{Timestamp: candle.Period, Data: out}
			return tsd.tda.Append(tsData)
		})
		if err != nil {
			common.Logger.Infof("Get dlog %s failed:%s", name, err)
			break
		}
		if tsd.tda != nil {
			infra.Gettsdb().CloseAppender(tsd.tda)
			tsd.tda = nil
		}
		common.Logger.Infof("redo :%s over", name)
	}
	if tsd.tda != nil {
		infra.Gettsdb().CloseAppender(tsd.tda)
	}
	common.Logger.Infof("redo over")
}

func (m *tsdbMigrate) Action() {
	common.Logger.Infof("migrate action ......")
	cnList := &tstock.CnBasicList{}
	err := infra.GetCnBasic(cnList)
	if err != nil {
		common.Logger.Infof("GetCnBasic failed:%s", err)
		return
	}
	datRang := []yearItems{{"20200101", "20201231"}, {"20210101", "20211231"}, {"20220101", "20221231"}, {"20230101", "20230819"}}
	isFaild := false
	common.Logger.Infof("migrate start......")
	dlg := faststore.FsTsdbLogGet("cnshares")
	for _, cn := range cnList.CnBasicList {
		call := faststore.FsTsdbGet("cnshares", cn.Symbol)
		for _, dr := range datRang {
			datas, err := infra.GetDataDayBetween(cn.Symbol, dr.start, dr.end, 0)
			if err != nil {
				common.Logger.Debugf("Get symbol=%s, failed:%s", cn.Symbol, err)
				continue
			}
			for f := datas.Front(); f != nil; f = f.Next() {
				value := f.Value.(*tsdb.TsdbData)
				fv := api.FstTsdbValue{Timestamp: int64(value.Timestamp), Data: value.Data}
				err = call.Append(&fv)
				dlg.Append(cn.Symbol, &fv)
				if err != nil {
					key := common.GetDay(common.YYYYMMDD, time.UnixMilli(int64(value.Timestamp)))
					common.Logger.Infof("migreate symbol=%s,key=%s, failed:%s", cn.Symbol, key, err)
					isFaild = true
					break
				}
			}
			if isFaild {
				break
			}
		}
		call.Close()
		if isFaild {
			break
		}
	}
	dlg.Close()
	common.Logger.Infof("migrate over")
}
