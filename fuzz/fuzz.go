package main

import (
	"bytes"
	"encoding/gob"
	"log"
	"math/rand"
	"os"
	"time"

	"google.golang.org/protobuf/proto"
	"taiyigo.com/common"
	"taiyigo.com/facade/tsdb"
	pb "taiyigo.com/facade/tsdb"
	"taiyigo.com/infra"
)

var gNow = uint64(1000000)

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
	now := gNow
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
	dt := "20230627"
	d, err := time.Parse("20060102", dt)
	if err != nil {
		log.Printf("err:%s", err)
	} else {
		log.Printf("err:%d", d.Unix())
	}
}

func testSingle() {
	conf := "../config/tao.yaml"
	common.BaseInit(conf)
	tsd := infra.Gettsdb()
	first := 29704768
	second := 29706150
	start := int(gNow)
	rangeNum := 0
	if first < start {
		rangeNum = (second - start + 1)
	} else {
		rangeNum = (second - first + 1)
	}
	tql := tsd.OpenQuery("btc_usd")
	datList, err := tql.GetRange(uint64(first), uint64(second), 0)
	if err != nil {
		log.Printf("get error:%d", err)
	} else {
		if rangeNum != datList.Len() {
			log.Printf("lens:%d is error", datList.Len())
		}
		for f := datList.Front(); f != nil; f = f.Next() {
			pdat := f.Value.(*tsdb.TsdbData)
			if pdat.Timestamp < uint64(first) || pdat.Timestamp > uint64(second) {
				log.Printf("Time is error:%d", pdat.Timestamp)
			}
		}
	}
	tsd.CloseQuery(tql)
	tsd.Close()
}

func testGn() {
	conf := "../config/tao.yaml"
	common.BaseInit(conf)
	tsd := infra.Gettsdb()
	tql := tsd.OpenQuery("btc_usd")
	value := uint64(87654321)
	number := 1234567
	datList, err := tql.GetPointN(value, number)
	if err != nil {
		log.Printf("errors:%s", err)
	} else {
		log.Printf("Len:%d", datList.Len())
		for f := datList.Front(); f != nil; f = f.Next() {
			pdat := f.Value.(*tsdb.TsdbData)
			diff := int(int64(pdat.Timestamp) - int64(value))
			if diff >= number {
				log.Printf("Time is error:%d, diff:%d", pdat.Timestamp, diff)
			}
		}
	}
	tsd.CloseQuery(tql)
	tsd.Close()
}

func testMgn() {
	conf := "../config/tao.yaml"
	common.BaseInit(conf)
	tsd := infra.Gettsdb()
	var yiyi int64 = 100000000
	now := int64(gNow)
	for i := 0; i < 90000; i++ {
		tql := tsd.OpenQuery("btc_usd")
		defer tsd.CloseQuery(tql)
		first := rand.Int63n(yiyi)
		if first < now {
			first = now
		}
		second := first + int64(rand.Intn(4000)) + 10
		number := int(second - first)
		log.Printf("Find Point: %d, number:%d", second, number)
		timeStart := time.Now()
		datList, err := tql.GetPointN(uint64(second), number)
		tcost := (time.Now().UnixMilli() - timeStart.UnixMilli())
		if err != nil {
			log.Printf("errors:%s", err)
			break
		} else {
			log.Printf("lens:%d, used:%d", datList.Len(), tcost)
			if datList.Len() != number {
				log.Printf("lens:%d is error", datList.Len())
				break
			}
			for f := datList.Front(); f != nil; f = f.Next() {
				pdat := f.Value.(*tsdb.TsdbData)
				if pdat.Timestamp < uint64(first) || pdat.Timestamp > uint64(second) {
					log.Printf("Time is error:%d", pdat.Timestamp)
					break
				}
			}
		}
	}
	tsd.Close()
}

func testQuery() {
	conf := "../config/tao.yaml"
	common.BaseInit(conf)
	tsd := infra.Gettsdb()
	var yiyi int64 = 100000000
	now := uint64(1000000)
	rangeNum := int64(0)
	for i := 0; i < 90000; i++ {
		tql := tsd.OpenQuery("btc_usd")
		defer tsd.CloseQuery(tql)

		first := rand.Int63n(yiyi)
		second := first + int64(rand.Intn(2000))
		if first < int64(now) {
			rangeNum = (second - int64(now) + 1)
		} else {
			rangeNum = (second - first + 1)
		}

		log.Printf("times:%d, find range[%d,%d], num:%d", i, first, second, rangeNum)
		timeStart := time.Now()
		datList, err := tql.GetRange(uint64(first), uint64(second), 0)
		diff := (time.Now().UnixMilli() - timeStart.UnixMilli())

		if err != nil {
			log.Printf("errors:%s", err)
			break
		} else {
			if datList == nil {
				if now > uint64(second) {
					log.Printf("empty")
					continue
				}
				log.Printf("overs")
				break
			}
			log.Printf("lens:%d, used:%d", datList.Len(), diff)
			if rangeNum != int64(datList.Len()) {
				log.Printf("lens:%d is error", datList.Len())
				break
			}

			for f := datList.Front(); f != nil; f = f.Next() {
				pdat := f.Value.(*tsdb.TsdbData)
				if pdat.Timestamp < uint64(first) || pdat.Timestamp > uint64(second) {
					log.Printf("Time is error:%d", pdat.Timestamp)
				}
			}
		}
	}
	tsd.Close()
}

func testCnShares() {
	conf := "../config/tao.yaml"
	common.BaseInit(conf)
	dlist, err := infra.QueryCnShareDailyRange("000001.SZ", "20180718", "20180718")
	if err != nil {
		log.Printf("It is error:%s", err)
	} else {
		log.Printf("len:%d", len(dlist))
		for _, v := range dlist {
			log.Printf("%+v", v)
		}
	}

	out, err := infra.QueryCnShareBasic("", "L")
	if err != nil {
		log.Printf("It is error:%s", err)
	} else {
		log.Printf("len:%d", out.Len())
		front := out.Front().Value.(*infra.CnSharesBasic)
		log.Printf("%+v", front)
	}

}

func main() {
	c := 'A'
	testBsd()
	switch c {
	case 'b':
		testQuery()
	case 's':
		testSingle()
	case 'w':
		testWr()
	case 'g':
		testGn()
	case 'm':
		testMgn()
	case 'A':
		testCnShares()
	}
}
