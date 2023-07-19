package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"taiyigo.com/brain"
	"taiyigo.com/common"
	"taiyigo.com/facade/dto"
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
	//now := common.GetDay(common.YYYYMMDD, time.Now())
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

	// datRang := []yearItems{{"20200101", "20201231"}, {"20210101", "20211231"}, {"20220101", "20221231"}, {"20230101", now}}
	datRang := []yearItems{{"20230101", now}}
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

func loadCnBasic(c *gin.Context) {
	cmd := dto.CnAdminCmd{}
	if err := c.BindJSON(&cmd); err != nil {
		common.Logger.Infoln("Can not find args")
		c.String(http.StatusBadRequest, "Can not find args")
		return
	}
	c.String(http.StatusOK, "Commond submitted")
	brain.GetBrain().Subscript(brain.TOPIC_ADMIN, &loadCnBasciActor{cmd: &cmd})
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

func mergeSTF(c *gin.Context) {
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
	brain.GetBrain().Subscript(brain.TOPIC_ADMIN, &brain.MergeSTF{})
}

func mergeAll(c *gin.Context) {
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
	brain.GetBrain().Subscript(brain.TOPIC_ADMIN, &brain.MergeAll{})
}

func justifyKeyValue(c *gin.Context) {
	cmd := dto.JustifyReq{}
	if err := c.BindJSON(&cmd); err != nil {
		common.Logger.Infoln("Can not find args")
		c.String(http.StatusBadRequest, "Can not find args")
		return
	}
	err := infra.SetKeyValue(cmd.Table, cmd.Key, cmd.Value)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	} else {
		c.String(http.StatusOK, "OK")
	}
}

func justifyStat(c *gin.Context) {
	cmd := dto.CnAdminCmd{}
	if err := c.BindJSON(&cmd); err != nil {
		common.Logger.Infoln("Can not find args")
		c.String(http.StatusBadRequest, "Can not find args")
		return
	}
	if cmd.Opt != "START" || cmd.Value == "" {
		common.Logger.Infoln("opt is error")
		c.String(http.StatusBadRequest, "opt is error")
		return
	}
	c.String(http.StatusOK, "Commond submitted")
	brain.GetBrain().Subscript(brain.TOPIC_ADMIN, &brain.JustifyStat{StartDay: cmd.Value})
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
