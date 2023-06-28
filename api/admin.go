package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/ratelimit"
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

func toCandle(dIt *infra.CnSharesDaily) *tstock.Candle {
	candle := &tstock.Candle{}
	period, err := common.ToDay(common.YYYYMMDD, dIt.Day)
	if err != nil {
		common.Logger.Infof("ToDay failed:%s", err)
		return nil
	}
	candle.Period = uint64(period.UnixMilli())
	candle.Pcg = dIt.Change
	candle.Pcgp = dIt.PctChg
	candle.Open = dIt.Open
	candle.Close = dIt.Close
	candle.High = dIt.High
	candle.Low = dIt.Low
	candle.Volume = uint32(dIt.Vol)
	candle.PreClose = dIt.PreClose
	candle.Amount = dIt.Amount
	return candle
}

func (actor *loadHistoryActor) Action() {
	common.Logger.Infof("history loading ......")
	if actor.cmd.Opt != "LOAD" {
		common.Logger.Infof("cmd:%s is error", actor.cmd.Opt)
		return
	}
	now := common.GetDay(common.YYYYMMDD, time.Now())
	b, err := infra.CheckAndSet(infra.CONF_TABLE, "cn_history_load", now)
	if err != nil {
		common.Logger.Infof("do cmd:%s is failed:%s", actor.cmd.Opt, err)
		return
	}
	if !b {
		common.Logger.Infof("do cmd:%s has done", actor.cmd.Opt)
		return
	}
	cnList, err := infra.QueryCnShareBasic("", "L")
	if err != nil {
		common.Logger.Infof("do cmd:%s is failed:%s", actor.cmd.Opt, err)
		return
	}

	err = infra.SaveCnBasic(cnList)
	if err != nil {
		common.Logger.Infof("SaveCnBasic failed:%s", actor.cmd.Opt, err)
	}
	datRang := []yearItems{{"20200101", "20201231"}, {"20210101", "20211231"}, {"20220101", "20221231"}, {"20230101", now}}
	limter := ratelimit.New(500, ratelimit.Per(time.Minute))
	tsDb := infra.Gettsdb()
	cnShareStatus := make(map[string]string)
	for _, v := range datRang {
		isError := false
		for front := cnList.Front(); front != nil; front = front.Next() {
			cnBasic := front.Value.(*infra.CnSharesBasic)
			cnShareLastDay, err := infra.GetByKey(infra.CONF_TABLE, cnBasic.Symbol)
			startDay := v.start
			if err == nil && cnShareLastDay != "" {
				if strings.Compare(v.end, cnShareLastDay) <= 0 {
					continue
				}
				startDay = cnShareLastDay
			}
			limter.Take()
			daily, err := infra.QueryCnShareDailyRange(cnBasic.Symbol, startDay, v.end)
			if err != nil {
				common.Logger.Warnf("Load symbol:%s, range[%s,%s],failed:%s", cnBasic.Symbol, v.start, v.end, err)
				isError = true
				break
			}

			tbl := tsDb.OpenAppender(cnBasic.Symbol)
			defer tsDb.CloseAppender(tbl)
			for _, dIt := range daily {
				candle := toCandle(dIt)
				if candle == nil {
					common.Logger.Warnf("Load symbol:%s, range[%s,%s],failed", cnBasic.Symbol, v.start, v.end)
					isError = true
					break
				}
				out, err := proto.Marshal(candle)
				if err != nil {
					common.Logger.Warnf("Load symbol:%s, range[%s,%s],failed:%s", cnBasic.Symbol, v.start, v.end, err)
					isError = true
					break
				}
				tsData := &tsdb.TsdbData{Timestamp: candle.Period, Data: out}
				tbl.Append(tsData)
				cnShareStatus[dIt.Symbol] = dIt.Day
			}
		}
		if isError {
			break
		}
	}
	if err := infra.BatchSetKeyValue(infra.CONF_TABLE, cnShareStatus); err != nil {
		common.Logger.Infof("BatchSetKeyValue failed:%s", err)
	}
	common.Logger.Infof("history load over")
}

func loadCnSharesHistory(c *gin.Context) {
	cmd := dto.CnAdminCmd{}
	if err := c.BindJSON(&cmd); err != nil {
		common.Logger.Infoln("Can not find args")
		c.String(http.StatusBadRequest, "Can not find args")
		return
	}
	c.String(http.StatusOK, "Commond submitted")
	brain.GetBrain().Subscript(brain.TOPIC_ADMIN, &loadHistoryActor{cmd: &cmd})
}

func startCnSTFFlow(c *gin.Context) {
	cmd := dto.CnAdminCmd{}
	if err := c.BindJSON(&cmd); err != nil {
		common.Logger.Infoln("Can not find args")
		c.String(http.StatusBadRequest, "Can not find args")
		return
	}
	if cmd.Opt != "START" {
		common.Logger.Infoln("opt is error")
		c.String(http.StatusBadRequest, "opt is error")
		return
	}
	c.String(http.StatusOK, "Commond submitted")
	brain.GetBrain().Subscript(brain.TOPIC_STF, &brain.FlowStart{})
}
