package infra

import (
	"bytes"
	"encoding/json"
	"io"
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
	hc := get()
	rsp, err := hc.client.Do(req)
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
	hc := get()
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
	rsp, err := hc.client.Do(req)
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
