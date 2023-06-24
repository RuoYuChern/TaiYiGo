package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"os"

	"google.golang.org/protobuf/proto"
	"taiyigo.com/common"
	pb "taiyigo.com/facade/tsdb"
	"taiyigo.com/infra"
)

func testMarshal() error {
	mt := &pb.TsdbMeta{Start: 1, End: 2, Addr: 3, Refblock: 4, Refitems: 5}
	out, err := proto.Marshal(mt)
	if err != nil {
		log.Printf("marshal error:%s", err.Error())
		return err
	}

	log.Printf("marshal len:%d", len(out))
	omt := &pb.TsdbMeta{}
	err = proto.Unmarshal(out, omt)
	if err != nil {
		log.Printf("Unmarshal error:%s", err.Error())
		return err
	}
	log.Printf("%+v", omt)

	tmt := &infra.TsMetaData{Start: 1000, End: 2789564123, Addr: 8000000, RefAddr: 10, Refblock: 107896543, Refitems: 1230098743}
	var network bytes.Buffer
	enc := gob.NewEncoder(&network)
	err = enc.Encode(tmt)
	if err != nil {
		log.Printf("Encode error:%s", err.Error())
		return err
	}

	log.Printf("encode len:%d", network.Len())

	dec := gob.NewDecoder(&network)
	var dtmt infra.TsMetaData
	err = dec.Decode(&dtmt)
	if err != nil {
		log.Printf("Decode error:%s", err.Error())
		return err
	}
	log.Printf("%+v", dtmt)
	return nil
}

func testWr() {
	conf := "../config/tao.yaml"
	common.BaseInit(conf)
	tsd := infra.Gettsdb()
	tbl := tsd.OpenAppender("btc_usd")
	data := []byte("Hello world")
	now := uint64(1000000)
	msg := &pb.TsdbData{Timestamp: now, Data: data}
	err := tbl.Append(msg)
	if err != nil {
		log.Printf("fuzz append failed:%s", err.Error())
		os.Exit(-1)
	}

	var times uint64 = 1
	var yiyi uint64 = 100000000
	for times <= yiyi {
		msg.Timestamp = now + times
		err = tbl.Append(msg)
		if err != nil {
			log.Printf("fuzz append failed:%s, times:%d", err.Error(), times)
			break
		}
		times += 1
	}
	tsd.CloseAppender(tbl)
	tsd.Close()
}

func fbsd(v int, isa []int) {
	low := 0
	higth := len(isa)
	mid := 0
	for low < higth {
		mid = (low + higth) / 2
		if v == isa[mid] {
			return
		}
		if v > isa[mid] {
			low = mid + 1
		} else {
			higth = mid
		}
	}
	if low < len(isa) {
		log.Printf("v:%d, low:%d, high:%d, mid:%d, v < low=%v", v, low, higth,
			mid, (v < isa[low]))
	} else {
		log.Printf("v:%d is out of range", v)
	}
}

func tbsd(v int, itup []tup) {
	low := 0
	higth := len(itup)
	mid := 0
	for low < higth {
		mid = (low + higth) / 2
		if (v >= itup[mid].l) && (v <= itup[mid].r) {
			log.Printf("v:%d best off:%d", v, mid)
			return
		}
		if v < itup[mid].l {
			higth = mid
		} else {
			low = mid + 1
		}
	}
	log.Printf("v:%d best off:%d, mid:%d, high:%d", v, low, mid, higth)
}

type tup struct {
	l int
	r int
}

func testBsd() {
	isa := []int{1, 3, 5, 7, 9, 11, 17, 19}
	bsa := []int{0, 2, 4, 6, 8, 10, 12, 14, 16, 20}
	for _, v := range bsa {
		fbsd(v, isa)
	}
	log.Printf("--------------------------------------\n")
	itup := []tup{{1, 3}, {5, 7}, {9, 11}, {17, 19}}
	for _, v := range bsa {
		tbsd(v, itup)
	}
}

func main() {
	testBsd()

	//if testMarshal() == nil {
	//	testWr()
	//}
}
