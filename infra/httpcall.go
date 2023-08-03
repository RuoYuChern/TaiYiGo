package infra

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin/binding"
	"taiyigo.com/common"
)

type httpConnector struct {
	common.TItemLife
	client *http.Client
	mu     sync.Mutex
}

var httpCon *httpConnector
var httpOnce sync.Once
var (
	defaultRetryWaitMin = 500 * time.Millisecond
	defaultRetryWaitMax = 3 * time.Second
	defaultRetryMax     = 4
	defReadLimit        = int64(4096)
)

func get() *httpConnector {
	httpOnce.Do(func() {
		tr := &http.Transport{
			MaxIdleConns:       10,
			MaxConnsPerHost:    50,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: true,
		}
		client := &http.Client{Transport: tr}
		httpCon = &httpConnector{client: client, mu: sync.Mutex{}}
		common.TaddLife(httpCon)
	})
	return httpCon
}

func (hc *httpConnector) Close() {

}

func doGet2(realUrl string, hearder map[string]string) ([]byte, error) {
	req, err := http.NewRequest("GET", realUrl, nil)
	if err != nil {
		common.Logger.Warnf("NewRequest:%s", err.Error())
		return nil, err
	}
	for k, v := range hearder {
		req.Header.Add(k, v)
	}
	rsp, err := doRetry(req)
	if err != nil {
		common.Logger.Warnf("Do Request err:%s", err.Error())
		return nil, err
	}
	defer rsp.Body.Close()
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		common.Logger.Warnf("Read Request err:%s", err.Error())
		return nil, err
	}
	return body, nil
}

func doGet(baseUrl string, path string, params map[string]string, hearder map[string]string, out any) error {
	base, err := url.Parse(baseUrl)
	if err != nil {
		common.Logger.Infof("parse failed %s\n", err)
		return err
	}
	qp := url.Values{}
	for k, v := range params {
		qp.Add(k, v)
	}
	base.Path = path
	base.RawQuery = qp.Encode()
	req, err := http.NewRequest("GET", base.String(), nil)
	if err != nil {
		common.Logger.Warnf("NewRequest:%s", err.Error())
		return err
	}
	for k, v := range hearder {
		req.Header.Add(k, v)
	}
	rsp, err := doRetry(req)
	if err != nil {
		common.Logger.Warnf("Do Request err:%s", err.Error())
		return err
	}
	defer rsp.Body.Close()
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		common.Logger.Warnf("Read Request err:%s", err.Error())
		return err
	}
	err = binding.JSON.BindBody(body, out)
	if err != nil {
		common.Logger.Warnf("BindBody err:%s", err.Error())
		return err
	}
	return err
}

func doPost(url string, reqBody any, out any) error {
	js, err := json.Marshal(reqBody)
	if err != nil {
		common.Logger.Warnf("marshal failed:%s", err.Error())
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(js))
	if err != nil {
		common.Logger.Warnf("NewRequest:%s", err.Error())
		return err
	}
	rsp, err := doRetry(req)
	if err != nil {
		common.Logger.Warnf("Do Request err:%s", err.Error())
		return err
	}
	defer rsp.Body.Close()
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		common.Logger.Warnf("Read Request err:%s", err.Error())
		return err
	}
	err = binding.JSON.BindBody(body, out)
	if err != nil {
		common.Logger.Warnf("BindBody err:%s", err.Error())
		return err
	}
	return err
}

func doRetry(req *http.Request) (*http.Response, error) {
	hc := get()
	times := 0
	for {
		rsp, err := hc.client.Do(req)
		checkOk, checkErr := checkRetry(rsp, err)
		if !checkOk {
			if checkErr != nil {
				err = checkErr
			}
			return rsp, err
		}
		if err == nil {
			drainBody(rsp)
		}
		times++
		if times >= defaultRetryMax {
			common.Logger.Infof("Retry times:%d >= max %d", times, defaultRetryMax)
			return rsp, err
		}
		wait := calBackOff(defaultRetryWaitMin, defaultRetryWaitMax, times)
		time.Sleep(wait)
	}
}

func drainBody(rsp *http.Response) {
	_, err := io.Copy(io.Discard, io.LimitReader(rsp.Body, defReadLimit))
	if err != nil {
		common.Logger.Infof("drainBody error:%s", err)
	}
	rsp.Body.Close()
}

func calBackOff(min, max time.Duration, attemptNum int) time.Duration {
	mult := math.Pow(2, float64(attemptNum)) * float64(min)
	sleep := time.Duration(mult)
	if float64(sleep) != mult || sleep > max {
		sleep = max
	}
	return sleep
}

func checkRetry(rsp *http.Response, err error) (bool, error) {
	if err != nil {
		return true, err
	}
	if rsp.StatusCode == 0 || rsp.StatusCode >= 500 {
		return true, nil
	}
	return false, err
}
