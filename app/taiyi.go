package main

import (
	"fmt"

	"github.com/tao/faststore"
	fapi "github.com/tao/faststore/api"
	"taiyigo.com/api"
	"taiyigo.com/brain"
	"taiyigo.com/common"
	"taiyigo.com/infra"
)

func startAll() {
	if err := infra.StartDb(); err != nil {
		panic(err)
	}
	infra.LoadData()
	dir := fmt.Sprintf("%s/ftsdb", common.Conf.Infra.FsDir)
	conf := &fapi.TsdbConf{Level: common.Conf.Log.Level, File: common.Conf.Log.File, MaxSize: common.Conf.Log.MaxSize,
		MaxBackups: common.Conf.Log.MaxBackups, MaxAge: common.Conf.Log.MaxAge, Env: common.Conf.Log.Env,
		DataDir: dir}
	err := faststore.Start(conf)
	if err != nil {
		common.Logger.Infof("migrate failed:%s", err)
		return
	}
}

func main() {
	app := common.AppLife{}
	brain := brain.Brain{}
	app.Start()
	startAll()
	brain.Start()
	api.StartApi()
	app.Wait()
	brain.Stop()
	faststore.Stop()
}
