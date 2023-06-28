package main

import (
	"taiyigo.com/api"
	"taiyigo.com/brain"
	"taiyigo.com/common"
	"taiyigo.com/infra"
)

func startAll() {
	if err := infra.StartDb(); err != nil {
		panic(err)
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
}
