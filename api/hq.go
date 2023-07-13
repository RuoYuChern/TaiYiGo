package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"taiyigo.com/common"
	"taiyigo.com/facade/dto"
	"taiyigo.com/facade/tstock"
	"taiyigo.com/infra"
)

func getSymbolTrend(c *gin.Context) {
	name := c.Query("stock")
	if name == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	symbol := infra.GetNameSymbol(name)
	if symbol == "" {
		c.String(http.StatusNotFound, "Not found")
		return
	}
	_, err := infra.GetByKey(infra.CONF_TABLE, infra.KEY_CNLOADHISTORY)
	if err != nil {
		common.Logger.Infof("GetByKey failed: %s", err)
		c.String(http.StatusNotFound, "Not found")
		return
	}

}

func getStfRecord(c *gin.Context) {
	opt := c.Query("opt")
	page := c.Query("page")
	stfRecord := tstock.StfRecordList{}
	err := infra.GetStfRecord(&stfRecord)
	if err != nil {
		common.Logger.Infof("GetStfRecord failed:%s", err)
		c.String(http.StatusInternalServerError, "Internal error")
		return
	}

	startOff := 0
	pageSize := 500
	if page != "" {
		startOff, err = strconv.Atoi(page)
		if err != nil {
			startOff = 0
		}
		startOff = startOff * pageSize
	}

	rsp := dto.StfResponse{}
	rsp.Code = http.StatusOK
	rsp.Msg = "OK"
	rsp.Data = make([]*dto.StfItem, 0, len(stfRecord.Stfs))
	off := 0
	total := 0
	for _, v := range stfRecord.Stfs {
		if opt != "" && strings.Compare(opt, v.Opt) != 0 {
			continue
		}
		if off < startOff {
			off++
			continue
		}

		name := infra.GetSymbolName(v.Symbol)
		rsp.Data = append(rsp.Data, &dto.StfItem{Name: name, Symbol: v.Symbol, Status: v.Status,
			Opt: v.Opt, LowDay: v.LowDay, HighDay: v.HighDay})
		total++
		if total >= pageSize {
			break
		}
	}
	c.JSON(http.StatusOK, &rsp)
}
