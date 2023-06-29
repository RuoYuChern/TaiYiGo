package infra

import (
	"container/list"
	"fmt"

	"google.golang.org/protobuf/proto"
	"taiyigo.com/facade/tstock"
)

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

func GetCnBasic(cnList *tstock.CnBasicList) error {
	tsf := tsFile{}
	err := tsf.read("cnbasic.dat", cnList)
	return err
}
