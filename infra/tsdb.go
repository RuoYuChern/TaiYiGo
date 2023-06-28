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
	tsfMap  *tsdbFMMap
	ttmd    *tsdbCursor
	tidx    *tsdbCursor
	left    *tsdbLeftCursor
}

type Tsdb struct {
	tblMap map[string]*TsdbAppender
	mu     sync.Mutex
}

var gtsdb *Tsdb
var gtsdbOnce sync.Once

type tsdbGuid struct {
	common.TItemLife
}

func (tsg *tsdbGuid) Close() {
	if gtsdb != nil {
		gtsdb.Close()
	}
}

// funcitons
func Gettsdb() *Tsdb {
	gtsdbOnce.Do(func() {
		gtsdb = &Tsdb{tblMap: make(map[string]*TsdbAppender)}
		common.TaddLife(&tsdbGuid{})
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
	common.Logger.Debugf("metaLen:%d, idxLen:%d", tbl.metaLen, tbl.idxLen)
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
	if tbl.widx == nil {
		//打开IDX
		name := fmt.Sprintf(gIDX_FILE_TPL, tbl.dir, tbl.curMeta.Refblock)
		tbl.widx = newMapper(name, gINDEX_FILE_SIZE)
		err := tbl.widx.open(os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			common.Logger.Infof("open %s failed:%s", name, err.Error())
			return err
		}
	}

	if err := tbl.checkAndNewIdx(tidx.Timestamp); err != nil {
		common.Logger.Infof("openNewIdx failed:%s", err.Error())
		return err
	}

	if err := tbl.checkBlockIsFull(tidx.Timestamp); err != nil {
		common.Logger.Infof("checkBlockIsFull failed:%s", err.Error())
		return err
	}

	err := tbl.widx.append(tidx)
	if err != nil {
		common.Logger.Infof("Append idx_%d failed:%s", tbl.curMeta.Refblock, err.Error())
		return err
	}
	tbl.curMeta.End = tidx.Timestamp + 1
	tbl.curMeta.Refitems = tbl.curMeta.Refitems + 1
	return nil
}

func (tbl *TsdbAppender) checkBlockIsFull(timestamp uint64) error {
	curMeta := tbl.curMeta
	//判断是否已经写满一块
	if curMeta.Refitems >= gIDX_BLOCK_SIZE {
		err := tbl.meta.writeAt(int64(curMeta.Addr), curMeta)
		if err != nil {
			common.Logger.Infof("Save ref, write meta failed:%s", err.Error())
			return err
		}
		addr := curMeta.Addr + uint64(tbl.metaLen)
		refAddr := curMeta.RefAddr + uint64(tbl.idxLen*int64(curMeta.Refitems))
		nextBlock := curMeta.Refblock
		if (refAddr + uint64(tbl.idxLen)) >= uint64(tbl.widx.maxSize) {
			common.Logger.Infof("this unexceptions: block:%d, refAddr=%d, maxSize=%d", nextBlock, refAddr, tbl.widx.maxSize)
		} else {
			tbl.curMeta = newTmd(timestamp, addr, refAddr, nextBlock)
		}
	}
	return nil
}

func (tbl *TsdbAppender) checkAndNewIdx(timestamp uint64) error {
	curMeta := tbl.curMeta
	//判断当前是否已经满
	if (tbl.idxLen + tbl.widx.size) >= tbl.widx.maxSize {
		//将当前保存
		common.Logger.Infof("%s is full", tbl.widx.name)
		err := tbl.meta.writeAt(int64(curMeta.Addr), curMeta)
		if err != nil {
			common.Logger.Infof("Save ref, write meta failed:%s", err.Error())
			return err
		}
		// 换新的idx
		tbl.widx.close()
		nextBlock := curMeta.Refblock + 1
		name := fmt.Sprintf(gIDX_FILE_TPL, tbl.dir, nextBlock)
		tbl.widx = newMapper(name, gINDEX_FILE_SIZE)
		err = tbl.widx.open(os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			common.Logger.Infof("open %s failed:%s", name, err.Error())
			return err
		}
		//重新初始化
		addr := curMeta.Addr + uint64(tbl.metaLen)
		tbl.curMeta = newTmd(timestamp, addr, 0, nextBlock)
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
	tsq.left = &tsdbLeftCursor{}
	common.Logger.Infof("metaLen:%d, idxLen:%d", tsq.metaLen, tsq.idxLen)
}

func (tsq *TsdbQuery) close() {
	if tsq.tsfMap != nil {
		tsq.tsfMap.close()
	}
}

// 获取当前时间倒推N个点
func (tsq *TsdbQuery) GetPointN(value uint64, num int) (*list.List, error) {

	err := tsq.openMeta()
	if err != nil {
		return nil, err
	}

	err = tsq.findTidOff(value, num)
	if err != nil {
		return nil, err
	}

	return tsq.loadNData(value, num)
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

func (tsq *TsdbQuery) findTidOff(value uint64, number int) error {
	if tsq.ttmd.isEof {
		return nil
	}
	tsq.ttmd.isEof = true
	lastMdOff, err := findTmdBest(tsq.tsfMap, value, value)
	if err != nil {
		common.Logger.Infof("findTmdBest error:%s", err)
		return err
	}
	if lastMdOff < 0 {
		common.Logger.Infof("Find none such range:[%d,%d)", value, value+1)
		return gIsEmpty
	}
	common.Logger.Infof("start: %d, Find tmd off:%d", value, lastMdOff)
	if err := tsq.tsfMap.lseek(lastMdOff); err != nil {
		return err
	}
	// 只需读一个
	tsdbMen := make([]byte, (gIo_Size/tsq.metaLen)*tsq.metaLen)
	var rightPtmd *TsMetaData
	end := value
	for rightPtmd == nil {
		n, err := tsq.tsfMap.batchRead(tsdbMen)
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
			if (ptmd.End <= value) || (ptmd.Refitems == 0) {
				// [) 左闭合右开
				continue
			}
			if end < ptmd.Start {
				common.Logger.Infof("tmd is over: %d > %d", ptmd.Start, end)
				break
			}
			common.Logger.Infof("start = %d, find ptmd = [%d, %d, Refblock:%d, Refitems:%d]", value, ptmd.Start, ptmd.End,
				ptmd.Refblock, ptmd.Refitems)
			rightPtmd = ptmd
			break
		}
	}
	//找到ptmd,要找到idx的位置
	if rightPtmd == nil {
		return gIsEmpty
	}
	//找value的相对偏移位置
	offset, err := findIdxOff(tsq.dir, rightPtmd, value)
	if err != nil {
		return err
	}
	tsq.left, err = findLeft(tsq.dir, offset, number, rightPtmd)
	if err != nil {
		/****/
		return err
	}
	return nil
}

func (tsq *TsdbQuery) loadNData(value uint64, number int) (*list.List, error) {
	if tsq.left.leftNum <= 0 {
		return nil, nil
	}
	outList := list.New()
	readOff := 0
	for (readOff < gLimit) && (tsq.left.leftNum > 0) {
		name := fmt.Sprintf(gIDX_FILE_TPL, tsq.dir, tsq.left.block)
		tsfMap := &tsdbFMMap{maxSize: gINDEX_FILE_SIZE, name: name}
		err := tsfMap.open(os.O_RDONLY, 0755)
		if err != nil {
			common.Logger.Infof("open %s failed:%s", name, err.Error())
			return nil, err
		}
		defer tsfMap.close()
		offset, err := loadNData(tsq.dir, tsfMap, tsq.left, value, outList)
		if err != nil {
			if isTargetError(err, gIsEof) {
				common.Logger.Infof("Read block=%d, from off=%d end", tsq.left.block, tsq.left.offset)
				tsq.left.leftNum = 0
				continue
			}
			common.Logger.Infof("Read block=%d,at off=%d, failed:%s", tsq.left.block, tsq.left.offset, err)
			return nil, err
		}
		tsq.left.offset = offset
		if tsq.left.offset >= tsfMap.size {
			tsq.left.block = tsq.left.block + 1
			tsq.left.offset = 0
		}
		readOff = outList.Len()
		tsq.left.leftNum -= readOff
	}
	return outList, nil
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
			common.Logger.Infof("Block %d start:%d,end:%d is error:[%d,%d]", ptmd.Refblock, ptmd.Start, ptmd.End, start, end)
			continue
		}
		if ptmd.Refitems == 0 {
			continue
		}
		name := fmt.Sprintf(gIDX_FILE_TPL, tsq.dir, ptmd.Refblock)
		err := loadIdx(name, ptmd, tsq.tidx, start, end)
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
	items := tsq.tsfMap.size / tsq.metaLen
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
		lastMdOff, err := findTmdBest(tsq.tsfMap, start, end)
		if err != nil {
			common.Logger.Infof("findTmdBest error:%s", err)
			return err
		}
		if lastMdOff < 0 {
			common.Logger.Infof("Find none such range:[%d,%d]", start, end)
			return gIsEmpty
		}
		//设置文件开始读取位置
		common.Logger.Infof("start: %d, Find tmd off:%d", start, lastMdOff)
		if err := tsq.tsfMap.lseek(lastMdOff); err != nil {
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
		n, err := tsq.tsfMap.batchRead(tsdbMen)
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
			if (ptmd.End <= start) || (ptmd.Refitems == 0) {
				// [) 左闭合右开
				continue
			}
			if end < ptmd.Start {
				// 设置读完
				common.Logger.Infof("tmd is over: %d > %d", ptmd.Start, end)
				tsq.ttmd.isEof = true
				break
			}
			common.Logger.Infof("start = %d, put ptmd = [%d, %d]", start, ptmd.Start, ptmd.End)
			tsq.ttmd.itemList.PushBack(ptmd)
		}
	}
	return nil
}

func (tsq *TsdbQuery) openMeta() error {
	if tsq.tsfMap != nil {
		return nil
	}
	name := fmt.Sprintf("%s/tsdb/%s/super_meta", common.Conf.Infra.FsDir, tsq.id)
	tsq.tsfMap = &tsdbFMMap{maxSize: gMETA_FILE_SIZE, name: name}
	err := tsq.tsfMap.open(os.O_RDONLY, 0755)
	if err != nil {
		common.Logger.Infof("open %s failed:%s", name, err.Error())
		return err
	}
	return nil
}
