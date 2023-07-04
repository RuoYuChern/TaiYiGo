package infra

import (
	"container/list"
	"fmt"
	"sync"

	"google.golang.org/protobuf/proto"
	"taiyigo.com/common"
	"taiyigo.com/facade/tstock"
)

type memData struct {
	cnShares        map[string]*tstock.CnBasic
	cnSharesNameMap map[string]string
}

var gmemData *memData
var gmenOnce sync.Once

func LoadMemData() {
	gmemData = &memData{cnShares: make(map[string]*tstock.CnBasic), cnSharesNameMap: map[string]string{}}
	cnList := tstock.CnBasicList{}
	err := GetCnBasic(&cnList)
	if err != nil {
		common.Logger.Infof("GetCnBasic failed:%s", err)
		return
	}
	for _, v := range cnList.CnBasicList {
		gmemData.cnShares[v.Symbol] = v
		gmemData.cnSharesNameMap[v.Name] = v.Symbol
	}
}

func GetSymbolName(symbol string) string {
	gmenOnce.Do(LoadMemData)
	n, ok := gmemData.cnShares[symbol]
	if ok {
		return n.Name
	}
	return ""
}

func SaveCnBasic(basics *list.List) error {
	cnList := &tstock.CnBasicList{Numbers: int32(basics.Len()), CnBasicList: make([]*tstock.CnBasic, basics.Len())}
	off := 0
	for front := basics.Front(); front != nil; front = front.Next() {
		share := front.Value.(*TjCnBasicInfo)
		cnb := &tstock.CnBasic{}
		cnb.Symbol = share.Symbol
		cnb.Name = share.Name
		cnb.Area = share.Area
		cnb.Industry = share.Industry
		cnb.FulName = share.FulName
		cnb.EnName = share.EnName
		cnb.CnName = share.CnName
		cnb.Market = share.Market
		cnb.ExChange = share.ExChange
		cnb.Status = share.Status
		cnb.ListDate = share.ListDate
		cnb.DelistDate = share.DelistDate
		cnb.IsHs = share.IsHs
		cnList.CnBasicList[off] = cnb
		off++
	}
	tsf := tsFile{}
	return tsf.write("cnbasic.dat", cnList)
}

func SaveStfList(status string, day string, msg proto.Message) error {
	tsf := tsFile{}
	name := fmt.Sprintf("%s_%s_stf.dat", day, status)
	return tsf.write(name, msg)
}

func GetStfList(status string, day string, msg proto.Message) error {
	tsf := tsFile{}
	name := fmt.Sprintf("%s_%s_stf.dat", day, status)
	err := tsf.read(name, msg)
	return err
}

func SaveStfRecord(msg proto.Message) error {
	tsf := tsFile{}
	name := fmt.Sprintf("normal_S_stf.dat")
	return tsf.write(name, msg)
}

func GetStfRecord(msg proto.Message) error {
	tsf := tsFile{}
	name := fmt.Sprintf("normal_S_stf.dat")
	return tsf.read(name, msg)
}

func GetCnBasic(cnList *tstock.CnBasicList) error {
	tsf := tsFile{}
	err := tsf.read("cnbasic.dat", cnList)
	return err
}
