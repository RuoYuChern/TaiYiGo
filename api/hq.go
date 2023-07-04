package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"taiyigo.com/common"
	"taiyigo.com/facade/dto"
	"taiyigo.com/facade/tstock"
	"taiyigo.com/infra"
)

func getStfRecord(c *gin.Context) {
	opt := c.Query("opt")
	stfRecord := tstock.StfRecordList{}
	err := infra.GetStfRecord(&stfRecord)
	if err != nil {
		common.Logger.Infof("GetStfRecord failed:%s", err)
		c.String(http.StatusInternalServerError, "Internal error")
		return
	}

	rsp := dto.StfResponse{}
	rsp.Code = http.StatusOK
	rsp.Msg = "OK"
	rsp.Data = make([]*dto.StfItem, 0, len(stfRecord.Stfs))
	for _, v := range stfRecord.Stfs {
		if opt != "" && strings.Compare(opt, v.Opt) != 0 {
			continue
		}
		name := infra.GetSymbolName(v.Symbol)
		rsp.Data = append(rsp.Data, &dto.StfItem{Name: name, Symbol: v.Symbol, Status: v.Status,
			Opt: v.Opt, LowDay: v.LowDay, HighDay: v.HighDay})
	}
	c.JSON(http.StatusOK, &rsp)
}
