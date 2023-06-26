package common

import (
	"container/list"
	"sync"
)

type TItemLife interface {
	Close()
}

var items *list.List = list.New()
var mu sync.Mutex = sync.Mutex{}

func TaddLife(c TItemLife) {
	mu.Lock()
	defer mu.Unlock()
	items.PushBack(c)
}

func tcloseItems() {
	mu.Lock()
	defer mu.Unlock()
	for c := items.Front(); c != nil; c = c.Next() {
		if c != nil {
			ac := c.Value.(TItemLife)
			ac.Close()
		}
	}
}
