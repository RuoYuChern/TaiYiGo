package algor

import "container/list"

type ThinkAlg interface {
	Name() string
	S(dat *list.List) bool
	T(dat *list.List) bool
	F(dat *list.List) bool
}

type Macd struct {
	ThinkAlg
}

func (macd *Macd) Name() string {
	return "mac"
}

func (macd *Macd) S(dat *list.List) bool {
	return false
}
func (macd *Macd) T(dat *list.List) bool {
	return false
}
func (macd *Macd) F(dat *list.List) bool {
	return false
}

func GetAlgList() *list.List {
	algs := list.New()
	algs.PushBack(&Macd{})
	return algs
}
