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
	"taiyigo.com/facade/tstock"
	"taiyigo.com/indicators"
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

type tup struct {
	l int
	r int
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

func testBsd() {
	candles := make([]*tstock.Candle, 7)
	candles[0] = &tstock.Candle{Open: 3.0}
	candles[1] = &tstock.Candle{Open: 6.0}
	candles[2] = &tstock.Candle{Open: 9.0}
	candles[3] = &tstock.Candle{Open: 11.0}
	candles[4] = &tstock.Candle{Open: 14.0}
	candles[5] = &tstock.Candle{Open: 15.0}
	candles[6] = &tstock.Candle{Open: 37.0}
	ts := indicators.NewTimeSeries(candles)
	id := indicators.NewSimpleMovingAverage(ts, indicators.GetOpen, 5)
	for off := 0; off <= ts.LastIndex(); off++ {
		log.Printf("Off:%d, sma:%f", off, id.Calculate(off).Float())
	}
	log.Printf("\n")
	id = indicators.NewSimpleMovingAverage2(ts, indicators.GetOpen, 5)
	for off := 0; off <= ts.LastIndex(); off++ {
		log.Printf("Off:%d, sma:%f", off, id.Calculate(off).Float())
	}

	id = indicators.NewEMAIndicator(ts, indicators.GetOpen, 5)
	log.Printf("\n")
	for off := 0; off <= ts.LastIndex(); off++ {
		log.Printf("Off:%d, ema:%f", off, id.Calculate(off).Float())
	}

	id = indicators.NewEMAIndicator2(ts, indicators.GetOpen, 5)
	log.Printf("\n")
	for off := 0; off <= ts.LastIndex(); off++ {
		log.Printf("Off:%d, ema:%f", off, id.Calculate(off).Float())
	}

	log.Printf("Year:%s", common.GetYear(time.Now()))
}

func testSingle() {
	conf := "../config/tao.yaml"
	common.BaseInit(conf)
	tsd := infra.Gettsdb()
	first, _ := common.ToDay(common.YYYYMMDD, "20230627")
	second, _ := common.ToDay(common.YYYYMMDD, "20230630")
	// start := int(gNow)
	// rangeNum := 0
	// if first < start {
	// 	rangeNum = (second - start + 1)
	// } else {
	// 	rangeNum = (second - first + 1)
	// }
	tql := tsd.OpenQuery("000150.SZ")
	datList, err := tql.GetRange(uint64(first.UnixMilli()), uint64(second.UnixMilli()), 0)
	if err != nil {
		log.Printf("get error:%s", err)
	} else {
		log.Printf("get len:%d", datList.Len())
		for f := datList.Front(); f != nil; f = f.Next() {
			pdat := f.Value.(*tsdb.TsdbData)
			if pdat.Timestamp < uint64(first.UnixMilli()) || pdat.Timestamp > uint64(second.UnixMilli()) {
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
		first := rand.Int63n(yiyi)
		if first < now {
			first = now
		}
		second := first + int64(rand.Intn(4000)) + 10
		number := int(second - first)
		log.Printf("Find Point: %d, number:%d", second, number)
		timeStart := time.Now()
		datList, err := tql.GetPointN(uint64(second), number)
		tsd.CloseQuery(tql)
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
		tsd.CloseQuery(tql)
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
	dlist, err := infra.QueryCnShareDailyRange("000001.SZ", "20180716", "20180718")
	if err != nil {
		log.Printf("It is error:%s", err)
	} else {
		log.Printf("len:%d", len(dlist))
		for _, v := range dlist {
			log.Printf("%+v", v)
		}
	}

	// out, err := infra.QueryCnShareBasic("", "L")
	// if err != nil {
	// 	log.Printf("It is error:%s", err)
	// } else {
	// 	log.Printf("len:%d", out.Len())
	// 	front := out.Front().Value.(*infra.CnSharesBasic)
	// 	log.Printf("%+v", front)
	// }
	// out, err := infra.GetBasicFromTj()
	// if err != nil {
	// 	log.Printf("It is error:%s", err)
	// } else {
	// 	log.Printf("len:%d", out.Len())
	// 	front := out.Front().Value.(*infra.TjCnBasicInfo)
	// 	log.Printf("%+v", front)
	// }

	// dailyList, err := infra.GetDailyFromTj("000007.SZ", "20230715", "20230719")
	// if err != nil {
	// 	log.Printf("It is error:%s", err)
	// } else {
	// 	log.Printf("len:%d", len(dailyList))
	// 	for _, v := range dailyList {
	// 		log.Printf("%+v", v)
	// 	}
	// }

}

func testGetReal() {
	stockList := []string{"000800.SZ", "600426.SH", "300474.SZ"}
	priceList, err := infra.BatchGetRealPrice(stockList)
	if err != nil {
		log.Printf("erros:%s", err)
		return
	}
	for _, p := range priceList {
		log.Printf("%+v", p)
	}

	price, err := infra.GetRealDaily("300474.SZ")
	if err != nil {
		log.Printf("erros:%s", err)
		return
	}
	log.Printf("%+v", price)
	datas, err := infra.GetCnKData("sh000001", "", 240, 10)
	if err != nil {
		log.Printf("erros:%s", err)
		return
	}
	for _, d := range datas {
		log.Printf("%+v", d)
	}
}

func testHeap() {
	lcp := func(f any, s any) int {
		fv := f.(int)
		sv := s.(int)
		return (fv - sv)
	}
	lskp := common.NewLp(8, lcp)
	values := [12]int{12, 10, 1, 19, 96, 97, 17, 13, 98, 99, 5, 100}
	for off := 0; off < 12; off++ {
		lskp.Add(values[off])
	}

	log.Printf("hpl:%d\n", lskp.Len())
	for {
		v := lskp.Top()
		if v == nil {
			break
		}
		log.Printf("%d\t", v.(int))
	}
	log.Printf("\n")
}

func main() {
	c := 'H'
	testBsd()
	tid := common.GetTid("abcd")
	log.Printf("Tid:%s: %v", tid, common.VerifyTid("abcd", tid))
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
	case 'H':
		testHeap()
	case 'R':
		testGetReal()
	}
}
