package infra

import (
	"container/list"
	"errors"
	"fmt"
	"os"
	"sync"

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
	ttmd    *tsdbCursor
	tidx    *tsdbCursor
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

func (tsdb *Tsdb) CloseAppender(append *TsdbAppender) {
	append.close()
}

func (tsdb *Tsdb) OpenQuery(id string) *TsdbQuery {
	tblQuery := &TsdbQuery{id: id}
	tblQuery.open()
	return tblQuery
}

func (tsdb *Tsdb) CloseQuery(query *TsdbQuery) {
	query.close()
}

func (tsdb *Tsdb) Close() {
	for _, v := range tsdb.tblMap {
		v.close()
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

func (tbl *TsdbAppender) close() {
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
	if isError(err, gIsEmpty) {
		return err
	}

	if (tbl.curMeta != nil) && (data.Timestamp < tbl.curMeta.End) {
		return errors.New("data is old")
	}

	// 初始化
	if tbl.curMeta == nil {
		tbl.curMeta = &TsMetaData{Start: data.Timestamp, End: data.Timestamp + 1, Addr: 0, RefAddr: 0,
			Refblock: 0, Refitems: 0}
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

	//判断是否已经写满一块
	if curMeta.Refitems >= gIDX_BLOCK_SIZE {
		err := tbl.meta.writeAt(int64(curMeta.Addr), curMeta)
		if err != nil {
			common.Logger.Infof("Save ref, write meta failed:%s", err.Error())
			return err
		}
		addr := curMeta.Addr + uint64(tbl.metaLen)
		refAddr := curMeta.RefAddr + uint64(tbl.idxLen*int64(curMeta.Refitems))
		tbl.curMeta = &TsMetaData{Start: tidx.Timestamp, End: tidx.Timestamp + 1, Addr: addr, RefAddr: refAddr, Refblock: curMeta.Refblock, Refitems: 0}
		curMeta = tbl.curMeta
	}

	err := tbl.widx.append(tidx)
	if isError(err, gIsFull) {
		common.Logger.Infof("Append idx_%d failed:%s", curMeta.Refblock, err.Error())
		return err
	}

	// 之前的idx文件已经写满了，换个新的
	if isTargetError(err, gIsFull) {
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
		//再次写进去
		err = tbl.widx.append(tidx)
		if err != nil {
			common.Logger.Infof("append %s failed:%s", name, err.Error())
			return err
		}
		// 获取新的metaCure
		addr := curMeta.Addr + uint64(tbl.metaLen)
		refBlock := curMeta.Refblock + 1
		tbl.curMeta = &TsMetaData{Start: tidx.Timestamp, End: tidx.Timestamp + 1, Addr: addr, RefAddr: 0, Refblock: refBlock, Refitems: 1}
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
	tsq.ttmd = &tsdbCursor{}
	tsq.tidx = &tsdbCursor{}
	common.Logger.Infof("metaLen:%d, idxLen:%d", tsq.metaLen, tsq.idxLen)
}

func (tsq *TsdbQuery) close() {
	if tsq.tidx != nil && tsq.tidx.tsfMap != nil {
		tsq.tidx.tsfMap.close()
	}
	if tsq.ttmd != nil && tsq.ttmd.tsfMap != nil {
		tsq.ttmd.tsfMap.close()
	}
}

// 获取区间
func (tsq *TsdbQuery) GetRange(start uint64, end uint64, offset int) (*list.List, error) {
	if tsq.tidx.qOffset > offset {
		common.Logger.Infof("offset = %d < lastOffset =%d", offset, tsq.tidx.qOffset)
		return nil, errors.New("offset error")
	}

	//打开文件
	err := tsq.openMeta()
	if err != nil {
		return nil, err
	}

	// 加载满足条件的
	err = tsq.findTmdOff(start, end)
	if err != nil {
		return nil, err
	}

	err = tsq.loadIndex(start, end, offset)
	if err != nil {
		return nil, err
	}

	return tsq.loadData(start, end, offset)
}

func (tsq *TsdbQuery) loadData(start uint64, end uint64, offset int) (*list.List, error) {
	if tsq.tidx.itemList.Len() == 0 {
		return nil, nil
	}
	datList := list.New()
	lessNum := 0
	bigNum := 0
	total := 0
	tr := &tsdbBaReader{dir: tsq.dir}
	tr.init()
	for tsq.tidx.itemList.Len() > 0 {
		if total >= gLimit {
			break
		}
		front := tsq.tidx.itemList.Front()
		tsq.tidx.itemList.Remove(front)
		pIdx := front.Value.(*TsIndexData)
		if pIdx.Timestamp < start {
			lessNum += 1
			continue
		}
		if pIdx.Timestamp > end {
			bigNum++
			break
		}
		if tsq.tidx.qOffset < offset {
			tsq.tidx.qOffset++
			continue
		}
		total++
		if !tr.checkAndPut(pIdx) {
			err := tr.readAndRest(datList)
			if err != nil {
				return nil, err
			}
			tr.checkAndPut(pIdx)
		}
	}
	if lessNum > 0 {
		common.Logger.Infof("alg cost: lessNum %d", lessNum)
	}
	if bigNum > 0 {
		common.Logger.Infof("to tail: bigNum %d", bigNum)
		tsq.tidx.isEof = true
		tsq.tidx.itemList = list.New()
	}
	err := tr.readAndRest(datList)
	if err != nil {
		return nil, err
	}
	return datList, nil
}

func (tsq *TsdbQuery) loadIndex(start uint64, end uint64, offset int) error {
	if tsq.tidx.itemList == nil {
		tsq.tidx.isEof = false
		tsq.tidx.itemList = list.New()

	}
	if (tsq.tidx.itemList.Len() + tsq.tidx.qOffset) >= (gLimit + offset) {
		return nil
	}
	if tsq.tidx.isEof {
		return nil
	}

	for tsq.ttmd.itemList.Len() > 0 {
		front := tsq.ttmd.itemList.Front()
		tsq.ttmd.itemList.Remove(front)
		ptmd := front.Value.(*TsMetaData)
		if (ptmd.End < start) || (ptmd.Start > end) {
			common.Logger.Infof("Block %d start:%d,end:%d is error:[%d,%d]", ptmd.Refblock,
				ptmd.Start, ptmd.End, start, end)
			continue
		}
		name := fmt.Sprintf(gIDX_FILE_TPL, tsq.dir, ptmd.Refblock)
		tsq.tidx.tsfMap = &tsdbFMMap{maxSize: gINDEX_FILE_SIZE, name: name}
		err := tsq.tidx.tsfMap.open(os.O_RDONLY, 0755)
		if err != nil {
			common.Logger.Infof("open %s failed:%s", name, err.Error())
			return err
		}
		defer tsq.tidx.tsfMap.close()
		err = loadIdx(ptmd, tsq.tidx, start, end)
		tsq.tidx.tsfMap = nil
		if err != nil {
			if isTargetError(err, gIsEof) {
				common.Logger.Infof("loadIndex: %d", tsq.tidx.itemList.Len())
				tsq.tidx.isEof = true
				tsq.ttmd.isEof = true
				break
			}
			common.Logger.Infof("loadIdx %s failed:%s", name, err.Error())
			return err
		}

		qOffset := tsq.tidx.itemList.Len() + tsq.tidx.qOffset
		if qOffset >= (gLimit + offset) {
			break
		}
	}
	return nil
}

func (tsq *TsdbQuery) findTmdOff(start uint64, end uint64) error {
	items := tsq.ttmd.tsfMap.size / tsq.metaLen
	if items == 0 {
		return gIsEmpty
	}

	// 文件已经读完
	if tsq.ttmd.isEof {
		return nil
	}

	if tsq.ttmd.qOffset >= gLimit {
		return nil
	}

	if tsq.ttmd.itemList == nil {
		// 找到最近位置
		lastMdOff, err := findTmdBest(tsq.ttmd.tsfMap, start, end)
		if err != nil {
			common.Logger.Infof("findTmdBest error:%s", err)
			return err
		}
		if lastMdOff < 0 {
			common.Logger.Infof("Find none such range:[%d,%d]", start, end)
			return gIsEmpty
		}
		//设置文件开始读取位置
		common.Logger.Infof("start: %d, off:%d", start, lastMdOff)
		if err := tsq.ttmd.tsfMap.lseek(lastMdOff); err != nil {
			return err
		}
		tsq.ttmd.itemList = list.New()
	}

	// 加载目标文件
	tsdbMen := make([]byte, (gIo_Size/tsq.metaLen)*tsq.metaLen)
	for tsq.ttmd.qOffset <= (gLimit * 2) {
		if tsq.ttmd.isEof {
			break
		}

		n, err := tsq.ttmd.tsfMap.batchRead(tsdbMen)
		if err != nil {
			common.Logger.Infof("batchRead failed:%s", err.Error())
			return err
		}
		if n == 0 {
			break
		}
		off := 0
		itemBuf := make([]byte, gTMD_LEN)
		for off < n {
			tcopy(itemBuf, tsdbMen, off)
			off += int(gTMD_LEN)
			ptmd := &TsMetaData{}
			if err := ptmd.UnmarshalBinary(itemBuf); err != nil {
				common.Logger.Infof("UnmarshalBinary failed:%s", err.Error())
				return err
			}

			if ptmd.Start > end {
				// 设置读完
				common.Logger.Infof("tmd is over: %d > %d", ptmd.Start, end)
				tsq.ttmd.isEof = true
				break
			}
			if ptmd.End < start {
				continue
			}
			tsq.ttmd.itemList.PushBack(ptmd)
		}
	}
	return nil
}

func (tsq *TsdbQuery) openMeta() error {
	if tsq.ttmd.tsfMap != nil {
		return nil
	}
	name := fmt.Sprintf("%s/tsdb/%s/super_meta", common.Conf.Infra.FsDir, tsq.id)
	tsq.ttmd.tsfMap = &tsdbFMMap{maxSize: gMETA_FILE_SIZE, name: name}
	err := tsq.ttmd.tsfMap.open(os.O_RDONLY, 0755)
	if err != nil {
		common.Logger.Infof("open %s failed:%s", name, err.Error())
		return err
	}
	return nil
}
