package infra

import (
	"container/list"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/huandu/skiplist"
	"taiyigo.com/common"
	"taiyigo.com/facade/tsdb"
)

type TsdbAppender struct {
	id      string
	dir     string
	metaLen int64
	idxLen  int64
	meta    *tsdbFMMap
	widx    *tsdbFMMap
	wdat    *tsdbFile
	curMeta *TsMetaData
}

type TsdbQuery struct {
	id      string
	dir     string
	metaLen int64
	idxLen  int64
	offset  int32
	qlist   *skiplist.SkipList
	meta    *tsdbFMMap
	widx    *tsdbFMMap
	wdat    *tsdbFile
}

type Tsdb struct {
	tblMap map[string]*TsdbAppender
	mu     sync.Mutex
}

var gtsdb *Tsdb
var gtsdbOnce sync.Once

// funcitons
func Gettsdb() *Tsdb {
	gtsdbOnce.Do(func() {
		gtsdb = &Tsdb{tblMap: make(map[string]*TsdbAppender)}
	})
	return gtsdb
}

// Tsdb
func (tsdb *Tsdb) OpenAppender(id string) *TsdbAppender {
	tsdb.mu.Lock()
	defer tsdb.mu.Unlock()
	tstble, ok := tsdb.tblMap[id]
	if ok {
		return tstble
	}

	tstble = &TsdbAppender{id: id}
	tstble.open()
	tsdb.tblMap[id] = tstble
	return tstble
}

func (tsdb *Tsdb) OpenQuery(id string) *TsdbQuery {
	tblQuery := &TsdbQuery{id: id}
	tblQuery.open()
	return tblQuery
}

func (tsdb *Tsdb) Close() {
	for _, v := range tsdb.tblMap {
		v.Close()
	}

	tsdb.tblMap = make(map[string]*TsdbAppender)
}

func (tsdb *Tsdb) remove(id string) {
	tsdb.mu.Lock()
	defer tsdb.mu.Unlock()
	delete(tsdb.tblMap, id)
}

// TsdbAppender
func (tbl *TsdbAppender) open() {
	tbl.dir = fmt.Sprintf("%s/tsdb/%s", common.Conf.Infra.FsDir, tbl.id)
	os.MkdirAll(tbl.dir, 0755)
	// 获取meta 和 idx 大小
	// 这里有个坑:数值类型的如果赋值为0,序列化的时候不会被序列化，导致长度为0
	tbl.metaLen = int64(gTMD_LEN)
	tbl.idxLen = int64(gTID_LEN)
	common.Logger.Infof("metaLen:%d, idxLen:%d", tbl.metaLen, tbl.idxLen)
}

func (tbl *TsdbAppender) Close() {
	if tbl.wdat != nil {
		//关闭wdat
		tbl.wdat.close()
		tbl.wdat = nil
	}
	if tbl.widx != nil {
		//关闭widx
		tbl.widx.close()
		tbl.widx = nil
	}

	if tbl.curMeta != nil {
		tbl.meta.writeAt(int64(tbl.curMeta.Addr), tbl.curMeta)
		tbl.curMeta = nil
	}

	if tbl.meta != nil {
		// 关闭meta
		tbl.meta.close()
	}

	Gettsdb().remove(tbl.id)
}

func (tbl *TsdbAppender) Append(data *tsdb.TsdbData) error {
	err := tbl.getLastMeta()
	if isError(err, tsdbEEmpty{}) {
		return err
	}

	if (tbl.curMeta != nil) && (data.Timestamp < tbl.curMeta.End) {
		return errors.New("data is old")
	}

	// 初始化
	if tbl.curMeta == nil {
		tbl.curMeta = &TsMetaData{Start: data.Timestamp, End: data.Timestamp + 1, Addr: 0, Refblock: 0, Refitems: 0}
	}
	//
	err = tbl.getWrite()
	if err != nil {
		return err
	}

	ref, err := tbl.wdat.append(data)
	if err != nil {
		return err
	}
	//填充时间
	ref.Timestamp = data.Timestamp
	// 保存index
	err = tbl.saveIndex(ref)
	return err
}

func (tbl *TsdbAppender) saveIndex(tidx *TsIndexData) error {
	curMeta := tbl.curMeta
	if tbl.widx == nil {
		//打开IDX
		name := fmt.Sprintf(gIDX_FILE_TPL, tbl.dir, curMeta.Refblock)
		tbl.widx = newMapper(name, gINDEX_FILE_SIZE)
		err := tbl.widx.open(os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			common.Logger.Infof("open %s failed:%s", name, err.Error())
			return err
		}
	}

	err := tbl.widx.append(tidx)
	if isError(err, tsdbEfull{}) {
		common.Logger.Infof("Append idx_%d failed:%s", curMeta.Refblock, err.Error())
		return err
	}

	// 之前的idx文件已经写满了，换个新的
	if isTargetError(err, tsdbEfull{}) {
		//将当前保存
		common.Logger.Infof("%s is full", tbl.widx.name)
		err = tbl.meta.writeAt(int64(curMeta.Addr), curMeta)
		if err != nil {
			common.Logger.Infof("Save ref, write meta failed:%s", err.Error())
			return err
		}
		// 换新的idx
		tbl.widx.close()
		name := fmt.Sprintf(gIDX_FILE_TPL, tbl.dir, curMeta.Refblock+1)
		tbl.widx = newMapper(name, gINDEX_FILE_SIZE)
		err = tbl.widx.open(os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			common.Logger.Infof("open %s failed:%s", name, err.Error())
			return err
		}
		err = tbl.widx.append(tidx)
		if err != nil {
			common.Logger.Infof("append %s failed:%s", name, err.Error())
			return err
		}
		// 获取新的metaCure
		addr := curMeta.Addr + uint64(tbl.metaLen)
		refBlock := curMeta.Refblock + 1
		tbl.curMeta = &TsMetaData{Start: tidx.Timestamp, End: tidx.Timestamp + 1, Addr: addr, Refblock: refBlock, Refitems: 1}
	} else {
		curMeta.End = tidx.Timestamp + 1
		curMeta.Refitems = curMeta.Refitems + 1
	}
	return nil
}

func (tbl *TsdbAppender) getWrite() error {
	if tbl.wdat != nil {
		return nil
	}

	block, err := getTargetBlockNo(tbl.dir, gDAT_PREFIX)
	if err != nil {
		return nil
	}

	tbl.wdat = &tsdbFile{maxSize: gDATA_FILE_SIZE, dir: tbl.dir, block: int32(block)}
	return tbl.wdat.open(os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
}

func (tbl *TsdbAppender) getLastMeta() error {
	if tbl.curMeta != nil {
		return nil
	}
	//
	name := fmt.Sprintf("%s/tsdb/%s/super_meta", common.Conf.Infra.FsDir, tbl.id)
	tbl.meta = newMapper(name, gMETA_FILE_SIZE)
	err := tbl.meta.open(os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		common.Logger.Infof("fmap open %s failed:%s", name, err.Error())
		return err
	}

	if tbl.meta.size == 0 {
		return nil
	}

	//比较大小是否合法
	if (tbl.meta.size % tbl.metaLen) != 0 {
		return errors.New("meta size is error")
	}

	tbl.curMeta = &TsMetaData{}
	// 获取最新的一个metadata
	offset := tbl.meta.size - tbl.metaLen
	err = tbl.meta.readAt(offset, tbl.metaLen, tbl.curMeta)
	if err != nil {
		common.Logger.Infof("meta read %s failed:%s", name, err.Error())
		return err
	}
	common.Logger.Infof("%s cure meta:%+v", tbl.id, tbl.curMeta)

	return nil
}

// TsdbQuery
func (tsq *TsdbQuery) open() {
	tsq.dir = fmt.Sprintf("%s/tsdb/%s", common.Conf.Infra.FsDir, tsq.id)
	// 获取meta 和 idx 大小
	// 这里有个坑:数值类型的如果赋值为0,序列化的时候不会被序列化，导致长度为0
	tsq.metaLen = int64(gTMD_LEN)
	tsq.idxLen = int64(gTID_LEN)
	common.Logger.Infof("metaLen:%d, idxLen:%d", tsq.metaLen, tsq.idxLen)
}

// 获取区间
func (tsq *TsdbQuery) GetRange(start uint64, end uint64, offset int) (*list.List, error) {
	err := tsq.openMeta()
	if err != nil {
		return nil, err
	}

	// 加载满足条件的
	err = tsq.load(start, end)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (tsq *TsdbQuery) load(start uint64, end uint64) error {
	items := tsq.meta.size / tsq.metaLen
	if items == 0 {
		return &tsdbEEmpty{}
	}

	bufSize := calMetaIoSize(tsq.meta.size, tsq.metaLen)
	if (bufSize % tsq.metaLen) != 0 {
		common.Logger.Infof("io size is error: meta size=%d, metaLen=%d, bufSize=%d", tsq.meta.size,
			tsq.metaLen, bufSize)
		return errors.New("io size is error")
	}
	rdItems := bufSize / tsq.metaLen
	rdBuf := make([]byte, bufSize)
	rlen := int64(0)
	for rlen < tsq.meta.size {
		tsq.meta.batchRdMeta(rdBuf, rdItems)
		rlen += bufSize
	}

	return nil
}

func (tsq *TsdbQuery) openMeta() error {
	if tsq.meta != nil {
		return nil
	}
	name := fmt.Sprintf("%s/tsdb/%s/super_meta", common.Conf.Infra.FsDir, tsq.id)
	tsq.meta = &tsdbFMMap{maxSize: gMETA_FILE_SIZE, name: name}
	err := tsq.meta.open(os.O_RDONLY, 0755)
	if err != nil {
		common.Logger.Infof("open %s failed:%s", name, err.Error())
		return err
	}
	return nil
}
