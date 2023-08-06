package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"taiyigo.com/brain"
	"taiyigo.com/common"
	"taiyigo.com/facade/dto"
	"taiyigo.com/infra"
)

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

func adminCommond(c *gin.Context) {
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
	if cmd.Cmd == "mstf" {
		brain.GetBrain().Subscript(brain.TOPIC_STF, &brain.MergeSTF{})
	} else if cmd.Cmd == "mgall" {
		brain.GetBrain().Subscript(brain.TOPIC_ADMIN, &brain.MergeAll{})
	} else if cmd.Cmd == "mstat" {
		brain.GetBrain().Subscript(brain.TOPIC_ADMIN, &brain.JustifyStat{StartDay: cmd.Value})
	} else if cmd.Cmd == "stfl" {
		brain.GetBrain().Subscript(brain.TOPIC_STF, &brain.FlowStart{})
	} else if cmd.Cmd == "tsdb" {
		brain.GetBrain().Subscript(brain.TOPIC_ADMIN, &tsdbRedo{})
	} else if cmd.Cmd == "forward" {
		brain.GetBrain().Subscript(brain.TOPIC_STF, &brain.ForwardFlow{})
	}

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
