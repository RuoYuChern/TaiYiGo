package main

import (
	"taiyigo.com/api"
	"taiyigo.com/common"
)

func main() {
	app := common.AppLife{}
	app.Start()
	api.StartApi()
	app.Wait()
}
