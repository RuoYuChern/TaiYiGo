package infra

import (
	"container/list"
	"encoding/json"
	"errors"

	"github.com/gin-gonic/gin/binding"
	"taiyigo.com/common"
)

type tuShareReq struct {
	ApiName string         `json:"api_name"`
	Token   string         `json:"token"`
	Params  map[string]any `json:"params"`
	Fields  []string       `json:"fields"`
}

type tuDailyItem struct {
	Fields  []string `json:"fields"`
	Items   [][]any  `json:"items"`
	HasMore bool     `json:"has_more"`
}

type tuDailyRsp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data tuDailyItem `json:"data"`
}

var (
	tuShareUrl = "http://api.tushare.pro"
)

func QueryCnShareBasic(exchange string, listStatus string) (*list.List, error) {
	params := make(map[string]any)
	params["exchange"] = exchange
	params["list_status"] = listStatus
	req := &tuShareReq{ApiName: "stock_basic", Token: common.Conf.Quotes.TuToken, Params: params}
	rsp := tuDailyRsp{}
	err := doPost(tuShareUrl, req, &rsp)
	if err != nil {
		common.Logger.Infof("CnShareDailyRange failed:%s", err)
		return nil, err
	}
	data := &rsp.Data
	if len(data.Items) == 0 {
		return nil, gIsCnEmpty
	}
	outList := list.New()
	vo := make(map[string]any)
	for _, item := range data.Items {
		for idx, v := range data.Fields {
			vo[v] = item[idx]
		}
		jstr, err := json.Marshal(vo)
		if err != nil {
			common.Logger.Infof("CnShareDailyRange Marshal failed:%s", err)
			return nil, nil
		}
		basicVo := &CnSharesBasic{}
		err = binding.JSON.BindBody(jstr, basicVo)
		if err != nil {
			common.Logger.Infof("CnShareDailyRange Marshal failed:%s", err)
			return nil, nil
		}
		outList.PushBack(basicVo)
	}

	return outList, nil
}

func QueryCnShareDaily(tscode string, tradeDate string) (*CnSharesDaily, error) {
	params := make(map[string]any)
	params["ts_code"] = tscode
	params["trade_date"] = tradeDate
	req := &tuShareReq{ApiName: "daily", Token: common.Conf.Quotes.TuToken, Params: params}
	rsp := tuDailyRsp{}
	err := doPost(tuShareUrl, req, &rsp)
	if err != nil {
		common.Logger.Infof("CnShareDailyRange failed:%s", err)
		return nil, err
	}
	if rsp.Code != 0 {
		common.Logger.Infof("CnShareDailyRange failed:%s", rsp.Msg)
		return nil, errors.New(rsp.Msg)
	}
	data := &rsp.Data
	if len(data.Items) == 0 {
		return nil, gIsCnEmpty
	}
	vo := make(map[string]any)
	dailyOut := &CnSharesDaily{}
	for _, item := range data.Items {
		for idx, v := range data.Fields {
			vo[v] = item[idx]
		}
		jstr, err := json.Marshal(vo)
		if err != nil {
			common.Logger.Infof("CnShareDailyRange Marshal failed:%s", err)
			return nil, nil
		}
		err = binding.JSON.BindBody(jstr, dailyOut)
		if err != nil {
			common.Logger.Infof("CnShareDailyRange Marshal failed:%s", err)
			return nil, nil
		} else {
			break
		}
	}
	return dailyOut, nil
}

func QueryCnShareDailyRange(tscode string, startDate string, endDate string) ([]*CnSharesDaily, error) {
	params := make(map[string]any)
	params["ts_code"] = tscode
	params["start_date"] = startDate
	params["end_date"] = endDate
	req := &tuShareReq{ApiName: "daily", Token: common.Conf.Quotes.TuToken, Params: params}
	rsp := tuDailyRsp{}
	err := doPost(tuShareUrl, req, &rsp)
	if err != nil {
		common.Logger.Infof("CnShareDailyRange failed:%s", err)
		return nil, err
	}
	if rsp.Code != 0 {
		common.Logger.Infof("CnShareDailyRange failed:%s", rsp.Msg)
		return nil, errors.New(rsp.Msg)
	}
	data := &rsp.Data
	if len(data.Items) == 0 {
		return nil, gIsCnEmpty
	}
	vo := make(map[string]any)
	dailyOut := make([]*CnSharesDaily, len(data.Items))
	for itIdx, item := range data.Items {
		for idx, v := range data.Fields {
			vo[v] = item[idx]
		}
		jstr, err := json.Marshal(vo)
		if err != nil {
			common.Logger.Infof("CnShareDailyRange Marshal failed:%s", err)
			return nil, nil
		}
		dailyVo := &CnSharesDaily{}
		err = binding.JSON.BindBody(jstr, dailyVo)
		if err != nil {
			common.Logger.Infof("CnShareDailyRange Marshal failed:%s", err)
			return nil, nil
		}
		dailyOut[itIdx] = dailyVo
	}
	return dailyOut, nil
}
