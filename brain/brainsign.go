package brain

import (
	"container/list"

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
	GetBrain().Subscript(TOPIC_ADMIN, &deltaLoadCnActor{})
}

func getSignalList() *list.List {
	sList := list.New()
	sList.PushBack(&deltaLoadSign{})
	return sList
}
