package infra

import (
	"container/list"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
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
	sinaUrl    = "http://hq.sinajs.cn"
)

func GetDailyFromTj(tscode string, startDate string, endDate string) ([]TjDailyInfo, error) {
	params := make(map[string]string)
	headers := make(map[string]string)
	params["stock"] = tscode
	params["start_date"] = startDate
	params["end_date"] = endDate
	timeStr := strconv.FormatInt(time.Now().Unix(), 10)
	content := strconv.FormatInt(rand.Int63n(1000000000), 36)
	headers["X-TJ-TIME"] = timeStr
	headers["X-TJ-NOISE"] = content
	headers["X-TJ-SIGNATURE"] = common.MD5Sign(common.Conf.Quotes.Sault, content, timeStr)
	rsp := tjDailyRsp{}
	err := doGet(common.Conf.Quotes.TjUrl, "/api/hq/get-symbol", params, headers, &rsp)
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
		params["page_num"] = strconv.Itoa(pageNum)
		timeStr := strconv.FormatInt(time.Now().Unix(), 10)
		content := strconv.FormatInt(rand.Int63n(1000000000), 36)
		headers["X-TJ-TIME"] = timeStr
		headers["X-TJ-NOISE"] = content
		headers["X-TJ-SIGNATURE"] = common.MD5Sign(common.Conf.Quotes.Sault, content, timeStr)
		rsp := tjCnBasicRsp{}
		err := doGet(common.Conf.Quotes.TjUrl, "/api/hq/get-cn-basic", params, headers, &rsp)
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

func toTsCnShare(item []any, fields []string, obj any) error {
	vo := make(map[string]any)
	for idx, v := range fields {
		vo[v] = item[idx]
	}
	jstr, err := json.Marshal(vo)
	if err != nil {
		common.Logger.Infof("CnShareDailyRange Marshal failed:%s", err)
		return err
	}
	err = binding.JSON.BindBody(jstr, obj)
	if err != nil {
		common.Logger.Infof("CnShareDailyRange Marshal failed:%s", err)
		return err
	}
	return nil
}

func QueryCnShareBasic(exchange string, listStatus string) (*list.List, error) {
	params := make(map[string]any)
	params["exchange"] = exchange
	params["list_status"] = listStatus
	req := &tuShareReq{ApiName: "stock_basic", Token: common.Conf.Quotes.TuBToKen, Params: params}
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

	for _, item := range data.Items {
		basicVo := &CnSharesBasic{}
		err = toTsCnShare(item, data.Fields, basicVo)
		if err != nil {
			common.Logger.Infof("CnShareDailyRange Marshal failed:%s", err)
			return nil, nil
		}
		if basicVo.Status == "" && listStatus != "" {
			basicVo.Status = listStatus
		}
		outList.PushBack(basicVo)
	}

	return outList, nil
}

func QueryCnShareDaily(tscode string, tradeDate string) (*TjDailyInfo, error) {
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
	dailyOut := &CnSharesDaily{}
	err = toTsCnShare(data.Items[0], data.Fields, dailyOut)
	if err != nil {
		common.Logger.Infof("CnShareDailyRange Marshal failed:%s", err)
		return nil, nil
	}
	return ToTjDailyInfo(dailyOut), nil
}

func QueryCnShareDailyRange(tscode string, startDate string, endDate string) ([]*TjDailyInfo, error) {
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
		return nil, nil
	}
	dailyOut := make([]*TjDailyInfo, len(data.Items))
	taiOff := len(data.Items) - 1
	for _, item := range data.Items {
		dailyVo := &CnSharesDaily{}
		err = toTsCnShare(item, data.Fields, dailyVo)
		if err != nil {
			common.Logger.Infof("CnShareDailyRange Marshal failed:%s", err)
			return nil, nil
		}
		dailyOut[taiOff] = ToTjDailyInfo(dailyVo)
		taiOff--
	}
	return dailyOut, nil
}

func BatchGetRealPrice(symbols []string) (map[string]*CnStockPrice, error) {
	var buf strings.Builder
	for _, v := range symbols {
		if buf.Len() > 0 {
			buf.WriteByte(',')
		}
		buf.WriteString(toSinaSymbol(v))
	}
	realUrl := fmt.Sprintf("%s/list=%s", sinaUrl, buf.String())
	headers := make(map[string]string)
	headers["Referer"] = "http://finance.sina.com.cn"
	body, err := doGet2(realUrl, headers)
	if err != nil {
		return nil, err
	}
	values := strings.ReplaceAll(string(body), "\n", "")
	valueList := strings.Split(values, ";")
	priceMap := make(map[string]*CnStockPrice)
	for _, v := range valueList {
		if v == "" {
			continue
		}
		price, err := toSinaVo(v)
		if err != nil {
			return nil, err
		}
		priceMap[price.Symbol] = price
	}
	return priceMap, nil
}

func toSinaVo(value string) (*CnStockPrice, error) {
	pos := strings.Index(value, "\"")
	end := strings.LastIndex(value, "\"")
	if pos < 0 || end < 0 {

		return nil, errors.New("pos or end < 0")
	}

	valueList := strings.Split(common.SubString(value, pos+1, end), ",")
	price := &CnStockPrice{}
	price.Name = valueList[0]
	price.Symbol = fromSinaSymbol(value)
	price.Open, _ = strconv.ParseFloat(valueList[1], 64)
	price.PreClose, _ = strconv.ParseFloat(valueList[2], 64)
	price.CurePrice, _ = strconv.ParseFloat(valueList[3], 64)
	price.High, _ = strconv.ParseFloat(valueList[4], 64)
	price.Low, _ = strconv.ParseFloat(valueList[5], 64)
	price.Date = strings.ReplaceAll(valueList[30], "-", "")
	price.Time = valueList[31]
	return price, nil
}

func toSinaSymbol(symbol string) string {
	subStr := common.SubString(symbol, 0, len(symbol)-3)
	if strings.HasSuffix(symbol, ".SZ") {
		return fmt.Sprintf("sz%s", subStr)
	} else if strings.HasSuffix(symbol, ".SH") {
		return fmt.Sprintf("sh%s", subStr)
	} else if strings.HasSuffix(symbol, ".BJ") {
		return fmt.Sprintf("bj%s", subStr)
	}
	return symbol
}

func fromSinaSymbol(symbol string) string {
	pos := strings.Index(symbol, "hq_str_")
	end := strings.Index(symbol, "=")
	if pos < 0 || end < 0 {
		return ""
	}
	pos += len("hq_str_")
	ts := common.SubString(symbol, pos, end)
	if strings.HasPrefix(ts, "sz") {
		return fmt.Sprintf("%s.SZ", common.PreSubString(ts, 2))
	} else if strings.HasPrefix(ts, "sh") {
		return fmt.Sprintf("%s.SH", common.PreSubString(ts, 2))
	} else if strings.HasPrefix(ts, "bj") {
		return fmt.Sprintf("%s.BJ", common.PreSubString(ts, 2))
	}
	return ""
}
