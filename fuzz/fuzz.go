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

	tmt := &infra.TsMetaData{Start: 1000, End: 2789564123, Addr: 8000000, Refblock: 107896543, Refitems: 1230098743}
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
	tbl.Close()
	tsd.Close()
}

func main() {
	if testMarshal() == nil {
		testWr()
	}
}
