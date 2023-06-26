package main

import (
	"taiyigo.com/api"
	"taiyigo.com/brain"
	"taiyigo.com/common"
)

func main() {
	app := common.AppLife{}
	brain := brain.Brain{}
	app.Start()
	brain.Start()
	api.StartApi()
	app.Wait()
	brain.Stop()
}
