package infra

import (
	"container/list"
	"encoding/json"
	"errors"
	"math/rand"
	"strconv"
	"time"

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

type tjCnBasePage struct {
	PageSize int             `json:"pageSize"`
	Items    []TjCnBasicInfo `json:"items"`
}

type tjCnBasicRsp struct {
	Status int          `json:"status"`
	Msg    string       `json:"msg"`
	Data   tjCnBasePage `json:"data"`
}

type tjDailyRange struct {
	Symbol  string        `json:"tsCode"`
	DtoList []TjDailyInfo `json:"dtoList"`
}

type tjDailyRsp struct {
	Status int          `json:"status"`
	Msg    string       `json:"msg"`
	Data   tjDailyRange `json:"data"`
}

var (
	tuShareUrl = "http://api.tushare.pro"
	tjUrl      = "https://www.taiji666.top"
)

func GetDailyFromTj(tscode string, startDate string, endDate string) ([]TjDailyInfo, error) {
	params := make(map[string]string)
	headers := make(map[string]string)
	params["stock"] = tscode
	params["startDate"] = startDate
	params["endDate"] = endDate
	timeStr := strconv.FormatInt(time.Now().Unix(), 10)
	content := strconv.FormatInt(rand.Int63n(1000000000), 36)
	headers["X-TJ-TIME"] = timeStr
	headers["X-TJ-NOISE"] = content
	headers["X-TJ-SIGNATURE"] = common.MD5Sign(common.Conf.Quotes.Sault, content, timeStr)
	rsp := tjDailyRsp{}
	err := doGet(tjUrl, "/api/hq/get-symbol", params, headers, &rsp)
	if err != nil {
		common.Logger.Infof("GetBasicFromTj failed: %s", err)
		return nil, err
	}
	return rsp.Data.DtoList, nil
}

func GetBasicFromTj() (*list.List, error) {
	params := make(map[string]string)
	headers := make(map[string]string)
	pageNum := 0
	outList := list.New()
	for {
		params["pageNum"] = strconv.Itoa(pageNum)
		timeStr := strconv.FormatInt(time.Now().Unix(), 10)
		content := strconv.FormatInt(rand.Int63n(1000000000), 36)
		headers["X-TJ-TIME"] = timeStr
		headers["X-TJ-NOISE"] = content
		headers["X-TJ-SIGNATURE"] = common.MD5Sign(common.Conf.Quotes.Sault, content, timeStr)
		rsp := tjCnBasicRsp{}
		err := doGet(tjUrl, "/api/hq/get-cn-basic", params, headers, &rsp)
		if err != nil {
			common.Logger.Infof("GetBasicFromTj failed: %s", err)
			return nil, err
		}
		if rsp.Status != 200 {
			common.Logger.Infof("GetBasicFromTj failed: %d, msg:%s", rsp.Status, rsp.Msg)
			return nil, errors.New(rsp.Msg)
		}
		itemLen := len(rsp.Data.Items)
		for off := 0; off < itemLen; off++ {
			outList.PushBack(&rsp.Data.Items[off])
		}
		if len(rsp.Data.Items) < rsp.Data.PageSize {
			break
		}
		pageNum++
	}
	return outList, nil
}

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
	if rsp.Code != 0 {
		common.Logger.Infof("QueryCnShareBasic failed:%s", rsp.Msg)
		return nil, errors.New(rsp.Msg)
	}
	data := &rsp.Data
	outList := list.New()
	if len(data.Items) == 0 {
		return outList, nil
	}
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
