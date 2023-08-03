package brain

import (
	"container/list"
	"time"

	"taiyigo.com/common"
)

type brainsign interface {
	isOnce() bool
	isTimeTo() bool
	doSign()
}

type deltaLoadSign struct {
	todayIsDone bool
}

type tradingSign struct {
	lastTime time.Time
}

func (dls *deltaLoadSign) isOnce() bool {
	return false
}

func (dls *deltaLoadSign) isTimeTo() bool {
	b := (common.GetTodayNHour() >= 18)
	if !b {
		dls.todayIsDone = false
	}
	return b && (!dls.todayIsDone)
}

func (dls *deltaLoadSign) doSign() {
	dls.todayIsDone = true
	GetBrain().Subscript(TOPIC_ADMIN, &loadActor{})
}

func (trade *tradingSign) isOnce() bool {
	return false
}

func (trade *tradingSign) isTimeTo() bool {
	diff := time.Now().Unix() - trade.lastTime.Unix()
	if diff < 300 {
		return false
	}
	h, m := common.GetTodayHAndM()
	// 09:30 ~ 11:30 || 13:00~15:00
	if (h == 9) && (m >= 30) {
		return true
	}
	if h == 10 {
		return true
	}
	if (h == 11) && (m <= 30) {
		return true
	}

	if (h >= 13) && (h < 15) {
		return true
	}
	return false
}

func (trade *tradingSign) doSign() {
	GetBrain().Subscript(TOPIC_STF, &TradeFlow{trade: trade})
}

func getSignalList() *list.List {
	sList := list.New()
	sList.PushBack(&deltaLoadSign{})
	sList.PushBack(&tradingSign{lastTime: time.Now()})
	return sList
}
