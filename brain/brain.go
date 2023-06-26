package brain

import (
	"container/list"
	"sync"

	"taiyigo.com/common"
)

var (
	TOPIC_ADMIN = "ADMIN"
	TOPIC_STF   = "STF"
)

type Topic struct {
	name   string
	q      *list.List
	isQuit bool
	cond   *sync.Cond
}

var gBrain *Brain

type Brain struct {
	topic map[string]*Topic
}

func GetBrain() *Brain {
	return gBrain
}

func doWork(topic *Topic) {
	common.Logger.Infof("worker:%s started", topic.name)
	for !topic.isQuit {
		topic.cond.L.Lock()
		if topic.q.Len() == 0 {
			topic.cond.Wait()
		}
		if topic.q.Len() == 0 {
			continue
		}
		front := topic.q.Front()
		topic.q.Remove(front)
		topic.cond.L.Unlock()
		actcor := front.Value.(common.Actor)
		actcor.Action()
	}
	common.Logger.Infof("worker:%s over", topic.name)
}

func (br *Brain) Start() {
	br.topic = make(map[string]*Topic)
	topicList := []string{TOPIC_ADMIN, TOPIC_STF}
	for _, v := range topicList {
		topic := &Topic{name: v, q: list.New(), isQuit: false, cond: sync.NewCond(&sync.Mutex{})}
		br.topic[v] = topic
		go doWork(topic)
	}
	gBrain = br
}

func (br *Brain) Stop() {
	for _, v := range br.topic {
		v.cond.L.Lock()
		v.cond.Broadcast()
		v.isQuit = true
		v.cond.L.Unlock()
	}
}

func (br *Brain) Subscript(topic string, actor common.Actor) {
	toppicQ, ok := br.topic[topic]
	if !ok {
		common.Logger.Infof("Find none such actor:%s", topic)
		return
	}
	toppicQ.cond.L.Lock()
	defer toppicQ.cond.L.Unlock()
	toppicQ.q.PushBack(actor)
	toppicQ.cond.Signal()
}
