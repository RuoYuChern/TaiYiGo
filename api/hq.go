package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"taiyigo.com/common"
	"taiyigo.com/facade/dto"
	"taiyigo.com/facade/tstock"
	"taiyigo.com/infra"
)

func postQuantPredit(c *gin.Context) {
	name := c.Query("stock")
	method := c.Query("method")
	if name == "" || method == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	symbol := infra.GetNameSymbol(name)
	if symbol == "" {
		c.String(http.StatusNotFound, "Not found")
		return
	}
	rsp := doPostQuantCal(symbol, method)
	c.JSON(http.StatusOK, rsp)
}

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
	rsp := &dto.SymbolTrendResponse{}
	data, err := calSymbolTrend(symbol)
	if err != nil {
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = err.Error()
	} else {
		rsp.Code = http.StatusOK
		rsp.Msg = "OK"
		rsp.Data = data
	}
	c.JSON(http.StatusOK, rsp)
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

	rsp := &dto.PaireSResponse{}
	rsp.Code = http.StatusOK
	rsp.Msg = "OK"
	fdata, err := calSymbolTrend(fsym)
	if err != nil {
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = err.Error()
		c.JSON(http.StatusOK, rsp)
		return
	}

	sdata, err := calSymbolTrend(ssym)
	if err != nil {
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = err.Error()
		c.JSON(http.StatusOK, rsp)
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
		c.JSON(http.StatusOK, rsp)
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

	c.JSON(http.StatusOK, rsp)
}

func getStfRecord(c *gin.Context) {
	opt := c.Query("opt")
	page := c.Query("page")
	orderDay := c.Query("orderday")
	if orderDay == "" {
		lastDay, err := infra.GetByKey(infra.CONF_TABLE, infra.KEY_CNLOADHISTORY)
		if err != nil {
			common.Logger.Infof("GetByKey failed:%s", err)
			c.String(http.StatusInternalServerError, "Internal error")
			return
		}
		orderDay = lastDay
	}
	stfRecord := tstock.StfList{}
	err := infra.GetStfList("S", orderDay, &stfRecord)
	if err != nil {
		common.Logger.Infof("GetStfList failed:%s", err)
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
	rsp := &dto.StfResponse{}
	rsp.Code = http.StatusOK
	rsp.Msg = "OK"
	rsp.Data = make([]*dto.StfItem, 0)
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
			Opt: v.Opt, LowDay: orderDay, HighDay: orderDay})
		total++
		if total >= pageSize {
			break
		}
	}
	c.JSON(http.StatusOK, rsp)
}

func getDashboard(c *gin.Context) {
	rsp := &dto.DashDailyResponse{}
	rsp.Code = http.StatusOK
	rsp.Msg = "OK"
	data, err := calLatestDash()
	if err != nil {
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = err.Error()
	} else {
		rsp.Data = data
	}
	c.JSON(http.StatusOK, rsp)
}

func getUpDown(c *gin.Context) {
	rsp := &dto.UpDownResponse{}
	rsp.Code = http.StatusOK
	rsp.Msg = "OK"
	orderDay := c.Query("orderDay")
	if orderDay == "" {
		lastDay, err := infra.GetByKey(infra.CONF_TABLE, infra.KEY_CNLOADHISTORY)
		if err != nil {
			rsp.Code = http.StatusNoContent
			rsp.Msg = err.Error()
			c.JSON(http.StatusOK, rsp)
			return
		}
		orderDay = lastDay
	}

	data, err := getLastUpDown(orderDay)
	if err != nil {
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = err.Error()
	} else {
		rsp.Data = data
	}
	c.JSON(http.StatusOK, rsp)
}

func getHot(c *gin.Context) {
	rsp := &dto.GetHotResponse{}
	rsp.Code = http.StatusOK
	rsp.Msg = "OK"

	orderDay := c.Query("orderDay")
	if orderDay == "" {
		lastDay, err := infra.GetByKey(infra.CONF_TABLE, infra.KEY_CNLOADHISTORY)
		if err != nil {
			rsp.Code = http.StatusNoContent
			rsp.Msg = err.Error()
			c.JSON(http.StatusOK, rsp)
			return
		}
		orderDay = lastDay
	}
	data, err := getLatestHot(orderDay)
	if err != nil {
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = err.Error()
	} else {
		rsp.Data = data
	}
	c.JSON(http.StatusOK, rsp)
}

func getSymbolLastN(c *gin.Context) {
	stock := c.Query("stock")
	name := c.Query("name")
	date := c.Query("date")
	num := c.Query("num")
	rsp := &dto.GetDailyResponse{}
	rsp.Code = http.StatusOK
	rsp.Msg = "OK"
	total, err := strconv.Atoi(num)
	if name != "" {
		stock = infra.GetNameSymbol(name)
	}
	if (stock == "") || (date == "") || (err != nil) {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "bad request"
		c.JSON(http.StatusOK, rsp)
		return
	}

	data, err := getStockNPoint(stock, date, int(total))
	if err != nil {
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = err.Error()
	} else {
		rsp.Data = data
	}
	c.JSON(http.StatusOK, rsp)
}

func getCnRtPrice(c *gin.Context) {
	stock := c.Query("stock")
	name := c.Query("name")
	way := c.Query("way")
	rsp := &dto.HqCommonRsp{}
	rsp.Code = http.StatusOK
	rsp.Msg = "OK"
	if way != "name" && way != "symbol" {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "bad request"
		c.JSON(http.StatusOK, rsp)
		return
	}

	if (stock == "" && way == "symbol") || (way == "name" && name == "") {
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "bad request"
		c.JSON(http.StatusOK, rsp)
		return
	}

	if way == "name" {
		stock = infra.GetNameSymbol(name)
		if stock == "" {
			rsp.Code = http.StatusBadRequest
			rsp.Msg = "bad request"
			common.Logger.Infof("Find none such symbol")
			c.JSON(http.StatusOK, rsp)
			return
		}
	}

	price, err := infra.GetRealDaily(stock)
	if err != nil {
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = err.Error()
	} else {
		rsp.Data = price
	}
	c.JSON(http.StatusOK, rsp)
}

func getForward(c *gin.Context) {
	mon := c.Query("month")
	if mon == "" {
		mon = common.GetYearMon(time.Now())
	}

	rsp := &dto.HqCommonRsp{}
	rsp.Code = http.StatusOK
	rsp.Msg = "OK"
	record := &tstock.ForwardStatRecord{}
	err := infra.GetForwardRecord(mon, record)
	if err != nil {
		rsp.Code = http.StatusNotFound
		rsp.Msg = err.Error()
		c.JSON(http.StatusOK, rsp)
		return
	}
	total := len(record.Items)
	items := make([]*dto.HqForwardItem, total)
	for off := 0; off < total; off++ {
		item := record.Items[off]
		items[off] = &dto.HqForwardItem{Day: item.Day, Total: int(item.Total), Success: int(item.Success), Failed: int(item.Failed)}
	}
	rsp.Data = items
	c.JSON(http.StatusOK, rsp)
}
