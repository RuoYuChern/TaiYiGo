package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"taiyigo.com/common"
	"taiyigo.com/facade/dto"
	"taiyigo.com/facade/tsorder"
	"taiyigo.com/infra"
)

func tradingStat(c *gin.Context) {
	rsp := &dto.TradingStatRsp{Code: http.StatusOK, Msg: "OK"}
	data, err := doGetTradingStat()
	if err != nil {
		common.Logger.Infof("tradingStat failed:%s", err)
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = err.Error()
	} else {
		rsp.Data = data
	}
	c.JSON(http.StatusOK, rsp)
}

func modifyTrading(c *gin.Context) {
	cmd := &dto.CnAdminCmd{}
	rsp := &dto.CommonResponse{Code: http.StatusOK, Msg: "OK"}
	if err := c.BindJSON(cmd); err != nil {
		common.Logger.Infof("Can not find args:%+v", cmd)
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "Args error"
		c.JSON(http.StatusOK, rsp)
		return
	}
	if cmd.Opt == "BuyDate" {
		tOrd := infra.GetOrder(cmd.Cmd)
		if tOrd != nil {
			tOrd.BuyDay = cmd.Value
			infra.SaveObject(infra.ORDER_TABLE, tOrd.OrderId, tOrd)
		}
	}
	c.JSON(http.StatusOK, rsp)
}

func doTrading(c *gin.Context) {
	req := dto.TradingReq{}
	rsp := &dto.CommonResponse{Code: http.StatusOK, Msg: "OK"}
	if err := c.BindJSON(&req); err != nil {
		common.Logger.Infof("Can not find args:%+v", req)
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "Args error"
		c.JSON(http.StatusOK, rsp)
		return
	}

	symbol := infra.GetNameSymbol(req.Stock)
	if symbol == "" {
		common.Logger.Infoln("Finde none symbol")
		rsp.Code = http.StatusBadRequest
		rsp.Msg = "Forbidden"
		c.JSON(http.StatusOK, rsp)
		return

	}

	nickName := c.GetString("Username")
	if nickName == "" {
		common.Logger.Infoln("Finde none use name")
		rsp.Code = http.StatusForbidden
		rsp.Msg = "Forbidden"
		c.JSON(http.StatusOK, rsp)
		return
	}

	if b, er := infra.CheckExist(infra.ORDER_TABLE, req.OrderId); b || er != nil {
		common.Logger.Infoln("Order exist")
		rsp.Code = http.StatusConflict
		rsp.Msg = "Order exist"
		c.JSON(http.StatusOK, rsp)
		return
	}

	if req.Vol <= 0 {
		req.Vol = 200
	}
	order := tsorder.TOrder{Name: req.Stock, OrderId: req.OrderId, OrderPrice: req.Price, Vol: int32(req.Vol), Buyer: nickName, Status: dto.ORDER_IDLE}
	order.CreatDay = common.GetDay(common.YYYYMMDD, time.Now())
	order.Symbol = symbol
	err := infra.SaveObject(infra.ORDER_TABLE, req.OrderId, &order)
	if err != nil {
		common.Logger.Infof("SaveObject failed:%s", err.Error())
		rsp.Code = http.StatusInternalServerError
		rsp.Msg = err.Error()
		c.JSON(http.StatusOK, rsp)
		return
	}
	infra.AddOrder(&order)
	c.JSON(http.StatusOK, rsp)
	common.Logger.Infof("doTrading: orderId:%s, symbol:%s, price:%f, vol:%d success", req.OrderId, symbol, req.Price, req.Vol)
}
