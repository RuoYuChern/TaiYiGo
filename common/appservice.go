package common

import (
	"container/list"
	"context"
	"os/signal"
	"syscall"
	"time"
)

type TaiyiApp interface {
	GetName() string
	Start(ctx *context.Context) error
	Stop(ctx *context.Context)
}

type AppLife struct {
	appList *list.List
	ctx     context.Context
	stop    context.CancelFunc
}

var gApp *AppLife

func GetApp() *AppLife {
	return gApp
}

func (app *AppLife) Start() error {
	conf := "../config/tao.yaml"
	BaseInit(conf)
	app.appList = list.New()
	app.ctx, app.stop = signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	gApp = app
	Logger.Info("AppLife startted")
	return nil
}

func (app *AppLife) Wait() {
	<-app.ctx.Done()
	app.stop()
	Logger.Info("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for app.appList.Len() > 0 {
		front := app.appList.Front()
		app.appList.Remove(front)
		serv := front.Value.(TaiyiApp)
		Logger.Infof("Server %s stopping.......", serv.GetName())
		serv.Stop(&ctx)
		Logger.Infof("Server %s shutdown", serv.GetName())
	}
	tcloseItems()
}

func (app *AppLife) AddService(serv TaiyiApp) error {
	err := serv.Start(&app.ctx)
	if err != nil {
		Logger.Infof("start %s failed:%s", serv.GetName(), err)
		return err
	}
	app.appList.PushBack(serv)
	return nil
}
