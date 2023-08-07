package brain

import (
	"container/list"
	"sync"
	"time"

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
var gTracheHour = -1

type Brain struct {
	topic map[string]*Topic
	bc    chan int
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
			topic.cond.L.Unlock()
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

func brainWork(br *Brain, c chan int) {
	isStop := false
	common.Logger.Infof("brainWork started")
	sList := getSignalList()
	for !isStop {
		hour := time.Now().Hour()
		isTrace := (gTracheHour != hour)
		if isTrace {
			common.Logger.Infof("brainWork poll begin")
			gTracheHour = hour
		}
		select {
		case res := <-c:
			common.Logger.Infof("Get signal:%d", res)
			isStop = true
		case <-time.After(5 * time.Second):
			poll(isTrace, sList)
		}
		if isTrace {
			common.Logger.Infof("brainWork poll over")
		}
	}
	common.Logger.Infof("brainWork stopped")
}

func poll(isTrace bool, sList *list.List) {
	for front := sList.Front(); front != nil; {
		bs := front.Value.(brainsign)
		old := front
		front = front.Next()
		if !bs.isTimeTo() {
			continue
		}
		bs.doSign()
		if bs.isOnce() {
			sList.Remove(old)
		}
	}
	if isTrace {
		common.Logger.Infof("size:%d", sList.Len())
	}
}

func (br *Brain) Start() {
	br.topic = make(map[string]*Topic)
	br.bc = make(chan int, 1)
	go brainWork(br, br.bc)
	topicList := []string{TOPIC_ADMIN, TOPIC_STF}
	for _, v := range topicList {
		topic := &Topic{name: v, q: list.New(), isQuit: false, cond: sync.NewCond(&sync.Mutex{})}
		br.topic[v] = topic
		go doWork(topic)
	}
	gBrain = br
}

func (br *Brain) Stop() {
	br.bc <- 0
	close(br.bc)
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
