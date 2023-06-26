package api

import (
	"net/http"

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
	cmd *dto.CnLoadHistoryCmd
}

func toCandle(dIt *infra.CnSharesDaily) *tstock.Candle {
	return nil
}

func (actor *loadHistoryActor) Action() {
	common.Logger.Infof("history loading ......")
	if actor.cmd.Opt != "LOAD" {
		common.Logger.Infof("cmd:%s is error", actor.cmd.Opt)
		return
	}
	b, err := infra.CheckAndSet(infra.CONF_TABLE, "cn_history_load", actor.cmd.Opt)
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
	datRang := []yearItems{{"20200101", "20201231"}, {"20210101", "20211231"}, {"20220101", "20221231"}, {"20230101", "20230626"}}
	limter := ratelimit.New(500)
	tsDb := infra.Gettsdb()
	for _, v := range datRang {
		isError := false
		for front := cnList.Front(); front != nil; front = front.Next() {
			limter.Take()
			cnBasic := front.Value.(*infra.CnSharesBasic)
			daily, err := infra.QueryCnShareDailyRange(cnBasic.Symbol, v.start, v.end)
			if err != nil {
				common.Logger.Warnf("Load symbol:%s, range[%s,%s],failed:%s", cnBasic.Symbol, v.start, v.end, err)
				isError = true
				break
			}

			tbl := tsDb.OpenAppender(cnBasic.Symbol)
			defer tsDb.CloseAppender(tbl)
			for _, dIt := range daily {
				candle := toCandle(dIt)
				out, err := proto.Marshal(candle)
				if err != nil {
					common.Logger.Warnf("Load symbol:%s, range[%s,%s],failed:%s", cnBasic.Symbol, v.start, v.end, err)
					isError = true
					break
				}
				tsData := &tsdb.TsdbData{Timestamp: candle.Period, Data: out}
				tbl.Append(tsData)
			}
		}
		if isError {
			break
		}
	}
	common.Logger.Infof("history load over")
}

func loadCnSharesHistory(c *gin.Context) {
	cmd := dto.CnLoadHistoryCmd{}
	if err := c.BindJSON(&cmd); err != nil {
		common.Logger.Infoln("Can not find args")
		c.String(http.StatusBadRequest, "Can not find args")
		return
	}
	c.String(http.StatusOK, "Hello")
	brain.GetBrain().Subscript(brain.TOPIC_ADMIN, &loadHistoryActor{cmd: &cmd})
}
