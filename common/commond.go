package common

import (
	"container/heap"
)

type Actor interface {
	Action()
}

type LessCmp func(any, any) int
type LimitedHeap struct {
	hp   []any
	lcmp LessCmp
	size int
	hpL  int
}

func (lp *LimitedHeap) Len() int           { return lp.hpL }
func (lp *LimitedHeap) Less(i, j int) bool { return (lp.lcmp(lp.hp[i], lp.hp[j]) < 0) }
func (lp *LimitedHeap) Swap(i, j int) {
	tmp := lp.hp[i]
	lp.hp[i] = lp.hp[j]
	lp.hp[j] = tmp
}

func NewLp(size int, lcmp LessCmp) *LimitedHeap {
	lp := &LimitedHeap{hp: make([]any, size), lcmp: lcmp, size: size, hpL: 0}
	return lp
}

func (lp *LimitedHeap) Push(x any) {
	lp.hp[lp.hpL] = x
	lp.hpL = lp.hpL + 1
}

func (lp *LimitedHeap) Pop() any {
	if lp.hpL <= 0 {
		return nil
	}
	n := lp.hpL
	x := lp.hp[n-1]
	lp.hpL = lp.hpL - 1
	return x
}

func (lp *LimitedHeap) Top() any {
	if lp.Len() == 0 {
		return nil
	}
	return heap.Pop(lp)
}

func (lp *LimitedHeap) Add(v any) {
	if lp.hpL < lp.size {
		heap.Push(lp, v)
		return
	}
	if lp.lcmp(v, lp.hp[lp.hpL-1]) <= 0 {
		return
	} else {
		heap.Pop(lp)
		heap.Push(lp, v)
	}
}
