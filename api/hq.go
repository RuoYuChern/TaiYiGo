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
	rsp := dto.SymbolTrendResponse{}
	data, err := calSymbolTrend(symbol)
	if err != nil {
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = err.Error()
	} else {
		rsp.Code = http.StatusOK
		rsp.Msg = "OK"
		rsp.Data = data
	}
	c.JSON(http.StatusOK, &rsp)
}

func getSymbolPairTrend(c *gin.Context) {
	first := c.Query("first")
	second := c.Query("second")
	if (first == "") || (second == "") {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	fsym := infra.GetNameSymbol(first)
	ssym := infra.GetNameSymbol(second)
	if (fsym == "") || (ssym == "") {
		c.String(http.StatusNotFound, "Not found")
		return
	}

	rsp := dto.PaireSResponse{}
	rsp.Code = http.StatusOK
	rsp.Msg = "OK"
	fdata, err := calSymbolTrend(fsym)
	if err != nil {
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = err.Error()
		c.JSON(http.StatusOK, &rsp)
		return
	}

	sdata, err := calSymbolTrend(ssym)
	if err != nil {
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = err.Error()
		c.JSON(http.StatusOK, &rsp)
		return
	}

	//找出时间段相同
	firstLen := len(fdata)
	firstOff := firstLen
	secondLen := len(sdata)
	secondOff := secondLen
	for (firstOff > 0) && (secondOff > 0) {
		if strings.Compare(fdata[firstOff-1].Day, sdata[secondOff-1].Day) != 0 {
			break
		}
		firstOff -= 1
		secondOff -= 1
	}
	if (firstOff >= firstLen) || (secondOff >= secondLen) {
		rsp.Code = http.StatusNoContent
		rsp.Msg = err.Error()
		c.JSON(http.StatusOK, &rsp)
		return
	}
	dataLen := (firstLen - firstOff)
	rsp.Data = make([]*dto.PairDaily, dataLen)
	for off := 0; off < dataLen; off++ {
		fDaily := fdata[firstOff+off]
		sDaily := sdata[secondOff+off]
		rsp.Data[off] = &dto.PairDaily{Day: fDaily.Day}
		rsp.Data[off].FClose = fDaily.Close
		rsp.Data[off].FVol = fDaily.Vol
		rsp.Data[off].FMtm = fDaily.Mtm
		rsp.Data[off].FLSma = fDaily.LSma
		rsp.Data[off].FSSma = fDaily.SSma

		rsp.Data[off].SClose = sDaily.Close
		rsp.Data[off].SVol = sDaily.Vol
		rsp.Data[off].SMtm = sDaily.Mtm
		rsp.Data[off].SLSma = sDaily.LSma
		rsp.Data[off].SSSma = sDaily.SSma

	}

	c.JSON(http.StatusOK, &rsp)
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
