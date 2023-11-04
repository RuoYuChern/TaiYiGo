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
	"taiyigo.com/facade/dto"
	"taiyigo.com/facade/tstock"
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

type tjQuantReq struct {
	Symbol string        `json:"symbol"`
	Tid    string        `json:"tid"`
	Method string        `json:"method"`
	Salt   string        `json:"salt"`
	CbUrl  string        `json:"cbUrl"`
	Noise  int64         `json:"noise"`
	Items  []*CnQuantDto `json:"items"`
}

type tjQuantRsp struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
	Tid    string `json:"tid"`
	Symbol string `json:"symbol"`
	Action string `json:"action"`
}

var (
	tuShareUrl = "http://api.tushare.pro"
	sinaUrl    = "http://hq.sinajs.cn"
	sinaKUrl   = "https://quotes.sina.cn/cn/api/jsonp_v2.php"
)

func DoPostQuant(tid, symbol, method string, candleList []*tstock.Candle) *tjQuantRsp {
	req := &tjQuantReq{Symbol: symbol, Tid: tid, Method: method, Noise: time.Now().Unix(), Items: make([]*CnQuantDto, len(candleList))}
	salt := common.QuantMd5(tid, method, symbol, fmt.Sprintf("%d", req.Noise), common.Conf.Quotes.Sault)
	req.Salt = salt
	if strings.HasPrefix(common.Conf.Quotes.Quantify, "http://127.0.0.1") {
		req.CbUrl = "http://127.0.0.1:9090/taiyi/hq/quant-cb"
	} else {
		req.CbUrl = "https://www.taiji666.top/taiyi/hq/quant-cb"
	}
	for offset := 0; offset < len(candleList); offset++ {
		req.Items[offset] = ToQuantFromCandle(candleList[offset])
	}
	tjRsp := &tjQuantRsp{Symbol: symbol, Tid: tid}
	err := doPost(common.Conf.Quotes.Quantify, req, tjRsp)
	if err != nil {
		tjRsp.Status = 500
		tjRsp.Msg = err.Error()
		common.Logger.Infof("doPost: url=%s faild", common.Conf.Quotes.Quantify)
	}
	return tjRsp
}

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

func GetCnKData(symbol string, ma string, scale int, dataLen int) ([]*dto.CnStockKData, error) {
	tsCode := toSinaSymbol(symbol)
	realUrl := fmt.Sprintf("%s/%s_%s_%d_%d=/CN_MarketDataService.getKLineData", sinaKUrl, "var%20", tsCode,
		scale, time.Now().Unix())
	queryMap := make(map[string]string)
	queryMap["symbol"] = tsCode
	queryMap["scale"] = strconv.Itoa(scale)
	if ma != "" {
		queryMap["ma"] = ma
	}
	queryMap["datalen"] = strconv.Itoa(dataLen)
	headers := make(map[string]string)
	headers["Referer"] = "http://finance.sina.com.cn"
	body, err := doGet2(realUrl, queryMap, headers)
	if err != nil {
		return nil, err
	}
	values := string(body)
	start := strings.Index(values, "[")
	end := strings.LastIndex(values, "]")
	if start < 0 || end < 0 {
		return nil, errors.New("data error")
	}
	datas := make([]*dto.CnStockKData, 0)
	err = binding.JSON.BindBody([]byte(common.SubString(values, start, end+1)), &datas)
	if err == nil {
		for _, d := range datas {
			d.Symbol = symbol
			d.Day = strings.ReplaceAll(d.Day, "-", "")
		}
	}
	return datas, nil
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
	body, err := doGet2(realUrl, nil, headers)
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

func GetRealDaily(symbol string) (*dto.CnStockDaily, error) {
	realUrl := fmt.Sprintf("%s/list=%s", sinaUrl, toSinaSymbol(symbol))
	headers := make(map[string]string)
	headers["Referer"] = "http://finance.sina.com.cn"
	body, err := doGet2(realUrl, nil, headers)
	if err != nil {
		return nil, err
	}
	values := strings.ReplaceAll(string(body), "\n", "")
	valueList := strings.Split(values, ";")
	for _, v := range valueList {
		if v == "" {
			continue
		}
		obj, err := toSinaVo2(v)
		if err != nil {
			common.Logger.Infof("GetRealDaily %s to failed:%s", symbol, err)
			return nil, err
		}
		return obj, nil
	}
	return nil, errors.New("find none value")
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
	price.Open = common.StrToF32(valueList[1])
	price.PreClose = common.StrToF32(valueList[2])
	price.CurePrice = common.StrToF32(valueList[3])
	price.High = common.StrToF32(valueList[4])
	price.Low = common.StrToF32(valueList[5])
	price.Date = strings.ReplaceAll(valueList[30], "-", "")
	price.Time = valueList[31]
	return price, nil
}

func toSinaVo2(value string) (*dto.CnStockDaily, error) {
	pos := strings.Index(value, "\"")
	end := strings.LastIndex(value, "\"")
	if pos < 0 || end < 0 {
		return nil, errors.New("pos or end < 0")
	}
	valueList := strings.Split(common.SubString(value, pos+1, end), ",")
	price := &dto.CnStockDaily{}
	price.Name = valueList[0]
	price.Symbol = fromSinaSymbol(value)
	price.Open = common.StrToF32(valueList[1])
	price.PreClose = common.StrToF32(valueList[2])
	price.CurePrice = common.StrToF32(valueList[3])
	price.High = common.StrToF32(valueList[4])
	price.Low = common.StrToF32(valueList[5])
	price.BuyPrice = common.StrToF32(valueList[6])
	price.SellPrice = common.StrToF32(valueList[7])
	price.Vol, _ = strconv.Atoi(valueList[8])
	price.Amount, _ = strconv.ParseFloat(valueList[9], 64)
	price.Buy1Vol, _ = strconv.Atoi(valueList[10])
	price.Buy1Price, _ = strconv.ParseFloat(valueList[11], 64)
	price.Buy2Vol, _ = strconv.Atoi(valueList[12])
	price.Buy2Price, _ = strconv.ParseFloat(valueList[13], 64)
	price.Buy3Vol, _ = strconv.Atoi(valueList[14])
	price.Buy3Price, _ = strconv.ParseFloat(valueList[15], 64)
	price.Buy4Vol, _ = strconv.Atoi(valueList[16])
	price.Buy4Price, _ = strconv.ParseFloat(valueList[17], 64)
	price.Buy5Vol, _ = strconv.Atoi(valueList[18])
	price.Buy5Price, _ = strconv.ParseFloat(valueList[19], 64)

	price.Sell1Vol, _ = strconv.Atoi(valueList[20])
	price.Sell1Price, _ = strconv.ParseFloat(valueList[21], 64)
	price.Sell2Vol, _ = strconv.Atoi(valueList[22])
	price.Sell2Price, _ = strconv.ParseFloat(valueList[23], 64)
	price.Sell3Vol, _ = strconv.Atoi(valueList[24])
	price.Sell3Price, _ = strconv.ParseFloat(valueList[25], 64)
	price.Sell4Vol, _ = strconv.Atoi(valueList[26])
	price.Sell4Price, _ = strconv.ParseFloat(valueList[27], 64)
	price.Sell5Vol, _ = strconv.Atoi(valueList[28])
	price.Sell5Price, _ = strconv.ParseFloat(valueList[29], 64)

	price.Date = strings.ReplaceAll(valueList[30], "-", "")
	price.Time = valueList[31]
	price.Status = valueList[32]
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
