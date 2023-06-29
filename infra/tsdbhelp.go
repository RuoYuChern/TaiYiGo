package infra

import (
	"container/list"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"google.golang.org/protobuf/proto"
	"taiyigo.com/common"
	"taiyigo.com/facade/tsdb"
)

var (
	gDATA_FILE_SIZE  = int64(1 << 30)
	gINDEX_FILE_SIZE = int64(300 << 20)
	gMETA_FILE_SIZE  = int64(5 << 20)
	gDAT_FILE_TPL    = "%s/dat_b_%d"
	gIDX_FILE_TPL    = "%s/idx_b_%d"
	gDAT_PREFIX      = "dat"
	gIo_Size         = int64(2 << 20)
	gFlush_Size      = int64(200 << 20)
	gIDX_BLOCK_SIZE  = (uint32)((3 << 20) / gTID_LEN)
	gLimit           = 5000
)

type tsdbIoBacth struct {
	buf    []byte
	offset int
	file   *os.File
}

type tsdbFile struct {
	maxSize  int64
	dir      string
	name     string
	block    int32
	size     int64
	cacheLen int64
	wcache   *tsdbIoBacth
	file     *os.File
}

type tsFile struct {
	file *os.File
}

type tsdbFMMap struct {
	maxSize int64
	name    string
	size    int64
	wcache  *tsdbIoBacth
	file    *os.File
}

type tsdbCursor struct {
	qOffset  int
	isEof    bool
	itemList *list.List
}

type tsdbLeftCursor struct {
	block   uint32
	offset  int64
	leftNum int
}

type tsdbBaReader struct {
	dir     string
	block   uint32
	offset  uint64
	bufLen  uint32
	idxList *list.List
}

// private functions
func newMapper(name string, maxSize int64) *tsdbFMMap {
	return &tsdbFMMap{maxSize: maxSize, name: name}
}

func newIoBatch(file *os.File) *tsdbIoBacth {
	iob := &tsdbIoBacth{buf: make([]byte, gIo_Size), offset: 0, file: file}
	return iob
}

func newTmd(t uint64, addr uint64, refAddr uint64, refBlock uint32) *TsMetaData {
	return &TsMetaData{Start: t, End: t + 1, Addr: addr, RefAddr: refAddr, Refblock: refBlock, Refitems: 0}
}

func tcopy(dst []byte, src []byte, offset int) {
	l := len(dst)
	for i := 0; i < l; i++ {
		dst[i] = src[offset+i]
	}
}

func topTailTmd(ioBuf []byte, itemBuf []byte, rlen int) (*TsMetaData, *TsMetaData) {
	tmdLow := &TsMetaData{}
	tcopy(itemBuf, ioBuf, 0)
	err := tmdLow.UnmarshalBinary(itemBuf)
	if err != nil {
		common.Logger.Infof("topTailTmd failed:%s", err.Error())
		return nil, nil
	}
	tcopy(itemBuf, ioBuf, rlen-int(gTMD_LEN))
	tmdHigh := &TsMetaData{}
	err = tmdHigh.UnmarshalBinary(itemBuf)
	if err != nil {
		common.Logger.Infof("topTailTmd failed:%s", err.Error())
		return nil, nil
	}
	return tmdLow, tmdHigh
}

func getTargetBlockNo(dir string, prexfix string) (int, error) {
	curBlock := -1
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		//过滤前缀
		if !strings.HasPrefix(d.Name(), prexfix) {
			return nil
		}
		parts := strings.SplitN(d.Name(), "_", 3)
		block, err := strconv.Atoi(parts[2])
		if err != nil {
			common.Logger.Warnf("Atoi %s failed:%s", d.Name(), err.Error())
			return nil
		}
		if block > curBlock {
			curBlock = block
		}
		return nil
	})
	if err != nil {
		return -1, nil
	}
	if curBlock < 0 {
		curBlock = 0
	}

	return curBlock, nil
}

// bsfind
func bsfindTmd(ioBuf []byte, bufLen int, start uint64, end uint64) (int, error) {
	if int64(bufLen)%gTMD_LEN != 0 {
		common.Logger.Infof("bufLen:%d is error", bufLen)
	}

	itemBuf := make([]byte, gTMD_LEN)
	high := int64(bufLen) / gTMD_LEN
	oHigh := high
	low := int64(0)
	mid := int64(0)
	for low < high {
		mid = (low + high) / 2
		offset := mid * gTMD_LEN
		tcopy(itemBuf, ioBuf, int(offset))
		tmd := &TsMetaData{}
		err := tmd.UnmarshalBinary(itemBuf)
		if err != nil {
			return -1, err
		}
		if (start >= tmd.Start) && (start <= tmd.End) {
			return int(mid * gTMD_LEN), nil
		}
		//在区间外
		if start < tmd.Start {
			//小于左区间
			high = mid
		} else {
			//大于右区间
			low = mid + 1
		}
	}
	//这种情况下,low是最佳值
	if low >= oHigh {
		common.Logger.Infof("some is unexceptions: low %d >= high %d", low, oHigh)
	}
	return int(low * gTMD_LEN), nil
}

// bsfindIdx
func bsfindIdx(ioBuf []byte, bufLen int, start uint64) (int, error) {
	if int64(bufLen)%gTID_LEN != 0 {
		common.Logger.Infof("bufLen:%d is error", bufLen)
	}
	itemBuf := make([]byte, gTID_LEN)
	high := int64(bufLen) / gTID_LEN
	oHigh := high
	low := int64(0)
	mid := int64(0)
	for low < high {
		mid = (low + high) / 2
		offset := mid * gTID_LEN
		tcopy(itemBuf, ioBuf, int(offset))
		tmd := &TsIndexData{}
		err := tmd.UnmarshalBinary(itemBuf)
		if err != nil {
			return -1, err
		}
		if tmd.Timestamp == start {
			common.Logger.Infof("bsfindIdx: start = %d, offset=%d,key=%d", start, offset, tmd.Timestamp)
			return int(offset), nil
		}
		//在区间外
		if start < tmd.Timestamp {
			//小于左区间
			high = mid
		} else {
			//大于右区间
			low = mid + 1
		}
	}
	if low >= oHigh {
		common.Logger.Infof("some is unexceptions: low %d >= high %d", low, oHigh)
		low = oHigh - 1
	}
	offset := int(low * gTID_LEN)
	tcopy(itemBuf, ioBuf, offset)
	tmd := &TsIndexData{}
	tmd.UnmarshalBinary(itemBuf)
	common.Logger.Infof("bsfindIdx: start = %d, offset=%d,key=%d", start, offset, tmd.Timestamp)
	return offset, nil
}

// midsearch
func findTmdBest(fsTmd *tsdbFMMap, start uint64, end uint64) (int64, error) {
	buffSize := (gIo_Size / gTMD_LEN) * gTMD_LEN
	ioBuf := make([]byte, buffSize)
	itemBuf := make([]byte, gTMD_LEN)
	high := fsTmd.size / buffSize
	if (fsTmd.size % buffSize) != 0 {
		high += 1
	}
	oHigh := high
	low := int64(0)
	bestOff := int64(-1)
	bestLen := 0
	mid := int64(0)
	for low < high {
		mid = (low + high) / 2
		offset := mid * buffSize
		n, err := fsTmd.batchReadAt(offset, ioBuf)
		if err != nil {
			return -1, err
		}
		if n == 0 {
			break
		}
		bestLen = n
		tmdLow, tmdHigh := topTailTmd(ioBuf, itemBuf, bestLen)
		if tmdLow == nil {
			return -1, errors.New("top tail tmd failed")
		}
		// 设置值
		if (start >= tmdLow.Start) && (start <= tmdHigh.End) {
			bestOff = offset
			break
		}
		if start < tmdLow.Start {
			high = mid
			continue
		}
		if start > tmdHigh.End {
			low = mid + 1
		}
	}
	if bestOff == -1 {
		if low >= oHigh {
			/**start在大区间之外**/
			common.Logger.Infof("This is exceptions:start=%d, low:%d, higth:%d, mid:%d", start, low, high, mid)
			return -1, nil
		}
		bestOff = (low * buffSize)
		/**当前位置是最佳**/
		return bestOff, nil
	}
	//找到区间,要确认最佳值
	off, err := bsfindTmd(ioBuf, bestLen, start, end)
	if err != nil {
		return -1, err
	}
	return (bestOff + int64(off)), nil
}

func findIdxOff(dir string, ptmd *TsMetaData, value uint64) (int64, error) {
	common.Logger.Infof("findIdxOff value=%d at block %d offset = %d", value, ptmd.Refblock, ptmd.RefAddr)
	if value == ptmd.Start {
		return 0, nil
	}
	bufSize := ptmd.Refitems * uint32(gTID_LEN)
	idxMen := make([]byte, bufSize)
	name := fmt.Sprintf(gIDX_FILE_TPL, dir, ptmd.Refblock)
	tsfm := newMapper(name, gINDEX_FILE_SIZE)
	defer tsfm.close()
	if err := tsfm.open(os.O_RDONLY, 0755); err != nil {
		common.Logger.Infof("open %s failed:%s", name, err)
		return 0, err
	}
	n, err := tsfm.batchReadAt(int64(ptmd.RefAddr), idxMen)
	if err != nil {
		common.Logger.Infof("read block %d offset = %d failed:%s", ptmd.Refblock, ptmd.RefAddr, err.Error())
		return 0, err
	}
	if n != int(bufSize) {
		common.Logger.Infof("read block %d at %d size error:%d != %d", ptmd.Refblock, ptmd.RefAddr, n, bufSize)
		return 0, errors.New("size error")
	}
	offset, err := bsfindIdx(idxMen, n, value)
	if err != nil {
		common.Logger.Infof("bsfindIdx block %d error:%s", ptmd.Refblock, err.Error())
		return 0, err
	}

	return int64(offset), nil
}

func findLeft(dir string, offset int64, num int, ptmd *TsMetaData) (*tsdbLeftCursor, error) {
	totalOffset := int64(num-1) * gTID_LEN
	leftCur := &tsdbLeftCursor{leftNum: num}
	valueOffset := int64(ptmd.RefAddr) + offset
	if totalOffset <= valueOffset {
		leftCur.block = ptmd.Refblock
		//通过相对便宜位置计算绝对位置
		leftCur.offset = (valueOffset - totalOffset)
		common.Logger.Infof("From [%d,%d] To [%d,%d]", leftCur.block, leftCur.offset, ptmd.Refblock, valueOffset)
		return leftCur, nil
	}
	totalOffset -= valueOffset
	leftBlock := ptmd.Refblock
	for leftBlock > 0 {
		name := fmt.Sprintf(gIDX_FILE_TPL, dir, (leftBlock - 1))
		tsfm := newMapper(name, gINDEX_FILE_SIZE)
		if err := tsfm.open(os.O_RDONLY, 0755); err != nil {
			common.Logger.Infof("open %s failed:%s", name, err)
			tsfm.close()
			return nil, err
		}
		if totalOffset <= tsfm.size {
			leftCur.block = (leftBlock - 1)
			leftCur.offset = (tsfm.size - totalOffset)
			common.Logger.Infof("From [%d,%d] To [%d,%d]", leftCur.block, leftCur.offset, ptmd.Refblock, valueOffset)
			tsfm.close()
			return leftCur, nil
		}
		tsfm.close()
		totalOffset -= tsfm.size
		leftBlock -= 1
	}
	common.Logger.Infof("Start refblock:%d, leftBlock:%d, totalOffset:%d", ptmd.Refblock, leftBlock, totalOffset)
	return nil, gIsEmpty
}

func loadNData(dir string, tsfMap *tsdbFMMap, left *tsdbLeftCursor, value uint64, outList *list.List) (int64, error) {
	bufSize := (gIo_Size / gTID_LEN) * gTID_LEN
	idxMen := make([]byte, bufSize)
	itemBuf := make([]byte, gTID_LEN)
	tr := &tsdbBaReader{dir: dir}
	tr.init()
	totalOff := outList.Len()
	offset := left.offset
	leftNum := left.leftNum
	for (offset < tsfMap.size) && (leftNum > 0) && (totalOff < gLimit) {
		n, err := tsfMap.batchReadAt(offset, idxMen)
		if err != nil && !isTargetError(err, gIsEof) {
			common.Logger.Infof("Read error:%s", err)
			return 0, err
		}
		if n == 0 {
			common.Logger.Infof("Read zero len")
			break
		}
		off := 0
		for (off < n) && (leftNum > 0) && (totalOff < gLimit) {
			tcopy(itemBuf, idxMen, off)
			off += int(gTID_LEN)
			offset += gTID_LEN
			totalOff += 1
			leftNum -= 1
			pIdx := &TsIndexData{}
			if err := pIdx.UnmarshalBinary(itemBuf); err != nil {
				common.Logger.Infof("UnmarshalBinary failed:%s", err.Error())
				return 0, err
			}
			if pIdx.Timestamp > value {
				common.Logger.Infof("End offset:%d, value:%d", offset, pIdx.Timestamp)
				leftNum = 0
				break
			}
			if !tr.checkAndPut(pIdx) {
				err := tr.readAndRest(outList)
				if err != nil {
					return 0, err
				}
				tr.checkAndPut(pIdx)
			}
		}
	}
	err := tr.readAndRest(outList)
	if err != nil {
		return 0, err
	}
	return offset, nil
}

func loadIdx(name string, ptmd *TsMetaData, cur *tsdbCursor, start uint64, end uint64) error {
	tsfMap := &tsdbFMMap{maxSize: gINDEX_FILE_SIZE, name: name}
	err := tsfMap.open(os.O_RDONLY, 0755)
	if err != nil {
		common.Logger.Infof("open %s failed:%s", name, err.Error())
		return err
	}
	defer tsfMap.close()

	bufSize := ptmd.Refitems * uint32(gTID_LEN)
	idxMen := make([]byte, bufSize)
	n, err := tsfMap.batchReadAt(int64(ptmd.RefAddr), idxMen)
	if err != nil {
		common.Logger.Infof("read block %d offset = %d failed:%s", ptmd.Refblock, ptmd.RefAddr, err.Error())
		return err
	}
	if n != int(bufSize) {
		common.Logger.Infof("read block %d at %d size error:%d != %d", ptmd.Refblock, ptmd.RefAddr, n, bufSize)
		return errors.New("size error")
	}
	offset := 0
	if start > ptmd.Start {
		offset, err = bsfindIdx(idxMen, n, start)
		if err != nil {
			common.Logger.Infof("bsfindIdx block %d error:%s", ptmd.Refblock, err.Error())
			return err
		}
	}
	items := (offset / int(gTID_LEN))
	common.Logger.Debugf("Block %d range:[%d,%d]find start %d begin %d of %d", ptmd.Refblock, ptmd.Start, ptmd.End,
		start, items, ptmd.Refitems)
	itemBuf := make([]byte, gTID_LEN)
	for items < int(ptmd.Refitems) {
		items += 1
		tcopy(itemBuf, idxMen, offset)
		offset += int(gTID_LEN)
		pIdx := &TsIndexData{}
		if err := pIdx.UnmarshalBinary(itemBuf); err != nil {
			common.Logger.Infof("block %d UnmarshalBinary failed:%s", ptmd.Refblock, err.Error())
			return err
		}
		if pIdx.Timestamp > end {
			common.Logger.Debugf("block %d is eof", ptmd.Refblock)
			return gIsEof
		}

		if pIdx.Timestamp >= start && pIdx.Timestamp <= end {
			cur.itemList.PushBack(pIdx)
		} else {
			common.Logger.Infof("This is exceptions: %d < %d, block:%d, offset:%d", pIdx.Timestamp, start, ptmd.Refblock, offset)
		}

	}
	return nil
}

// tsdbIoBacth
func (iob *tsdbIoBacth) append(dat []byte) error {
	dlen := len(dat)
	doff := 0
	for doff < dlen {
		rlen := int(gIo_Size) - iob.offset
		if rlen == 0 {
			_, err := iob.file.Write(iob.buf)
			if err != nil {
				common.Logger.Infof("iobatch append io failed:%s", err.Error())
				return err
			}
			iob.offset = 0
			continue
		}
		if rlen > (dlen - doff) {
			rlen = dlen - doff
		}
		for off := 0; off < rlen; off++ {
			iob.buf[iob.offset+off] = dat[doff]
			doff++
		}
		iob.offset += rlen
	}
	return nil
}

func (iob *tsdbIoBacth) flush() error {
	if iob.offset == 0 {
		return nil
	}
	_, err := iob.file.Write(iob.buf[0:iob.offset])
	if err != nil {
		common.Logger.Infof("iobatch close io failed:%s", err.Error())
		return err
	}
	iob.offset = 0
	return nil
}

// TsdbFile
func (tsf *tsdbFile) open(flag int, perm os.FileMode) error {
	tsf.name = fmt.Sprintf(gDAT_FILE_TPL, tsf.dir, tsf.block)
	file, err := os.OpenFile(tsf.name, flag, perm)
	if err != nil {
		common.Logger.Warnf("Open file %s failed:%s", tsf.name, err.Error())
		return err
	}

	info, err := os.Stat(tsf.name)
	if err != nil {
		common.Logger.Warnf("stat file %s failed:%s", tsf.name, err.Error())
		file.Close()
		return err
	}

	tsf.file = file
	tsf.size = info.Size()
	if (os.O_APPEND & flag) > 0 {
		tsf.wcache = newIoBatch(tsf.file)
	}
	return nil
}

func (tsf *tsdbFile) close() {
	if tsf.wcache != nil {
		tsf.wcache.flush()
	}

	if tsf.file != nil {
		tsf.file.Close()
		tsf.file = nil
	}
}

func (tsf *tsdbFile) isFull(dataLen int64) bool {
	return ((dataLen + tsf.size) > tsf.maxSize)
}

func (tsf *tsdbFile) bReadAt(offset int64, buf []byte, dLen int64) error {
	if (offset + dLen) > tsf.size {
		common.Logger.Infof("Read out of bound:%d > %d", (offset + dLen), tsf.size)
		return errors.New("out of bound")
	}
	n, err := tsf.file.ReadAt(buf, offset)
	if err != nil {
		common.Logger.Infof("Read %s failed:%s", tsf.name, err.Error())
		return err
	}
	if n != int(dLen) {
		common.Logger.Infof("Read %s failed:%d != %d", tsf.name, n, dLen)
		return errors.New("read len error")
	}
	return nil
}

func (tsf *tsdbFile) readAt(offset int64, dLen int64, msg proto.Message) error {
	if (offset + dLen) > tsf.size {
		common.Logger.Infof("Read out of bound:%d > %d", (offset + dLen), tsf.size)
		return errors.New("out of bound")
	}
	dat := make([]byte, dLen)
	n, err := tsf.file.ReadAt(dat, offset)
	if err != nil {
		common.Logger.Infof("Read %s failed:%s", tsf.name, err.Error())
		return err
	}
	if n != int(dLen) {
		common.Logger.Infof("Read %s failed:%d != %d", tsf.name, n, dLen)
		return errors.New("read len error")
	}

	err = proto.Unmarshal(dat, msg)
	if err != nil {
		common.Logger.Infof("Read %s Unmarshal failed:%s", tsf.name, err.Error())
		return err
	}
	return nil
}

func (tsf *tsdbFile) append(msg proto.Message) (*TsIndexData, error) {
	out, err := proto.Marshal(msg)
	if err != nil {
		common.Logger.Infof("Append %s marshal failed:%s", tsf.name, err.Error())
		return nil, err
	}

	dataLen := int64(len(out))
	if tsf.isFull(dataLen) {
		//重新打开
		err = tsf.openNew(os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			return nil, err
		}
	}

	err = tsf.wcache.append(out)
	if err != nil {
		common.Logger.Infof("Append %s failed:%s", tsf.name, err.Error())
		return nil, err
	}

	tsf.cacheLen = tsf.cacheLen + dataLen
	tsf.flush()
	ref := &TsIndexData{Block: uint32(tsf.block), Offset: uint64(tsf.size), Len: uint32(dataLen)}
	tsf.size += dataLen
	return ref, nil

}

func (tsf *tsdbFile) flush() {
	if tsf.cacheLen >= gFlush_Size {
		tsf.file.Sync()
		tsf.cacheLen = 0
	}
}

func (tsf *tsdbFile) openNew(flag int, perm os.FileMode) error {
	tsf.close()
	tsf.block = tsf.block + 1
	err := tsf.open(flag, perm)
	if err != nil {
		common.Logger.Info("reopen failed")
	}
	return err
}

// TsdbFMMap
func (tsfm *tsdbFMMap) open(flag int, perm os.FileMode) error {
	f, err := os.OpenFile(tsfm.name, flag, perm)
	if err != nil {
		common.Logger.Warnf("Open file %s failed:%s", tsfm.name, err.Error())
		return err
	}

	info, err := os.Stat(tsfm.name)
	if err != nil {
		common.Logger.Warnf("open file %s stat failed:%s", tsfm.name, err.Error())
		//关闭文件
		f.Close()
		return err
	}
	tsfm.file = f
	tsfm.size = info.Size()
	if (os.O_APPEND & flag) > 0 {
		tsfm.wcache = newIoBatch(tsfm.file)
	}
	common.Logger.Debugf("%s len:%d, cap:%d", tsfm.name, tsfm.size, tsfm.maxSize)
	return nil
}

func (tsfm *tsdbFMMap) writeAt(offset int64, msg TsData) error {
	out, err := msg.MarshalBinary()
	if err != nil {
		common.Logger.Infof("Write %s marshal failed:%s", tsfm.name, err.Error())
		return err
	}

	wl := int64(len(out))
	if (offset + wl) > tsfm.maxSize {
		common.Logger.Infof("Write %s out of bound:%d > %d", tsfm.name, (offset + wl), tsfm.maxSize)
		return gIsFull
	}

	_, err = tsfm.file.WriteAt(out, offset)
	if err != nil {
		common.Logger.Infof("Write %s failed:%s", tsfm.name, err.Error())
		return err
	}

	if (offset + wl) > tsfm.size {
		tsfm.size = (offset + wl)
	}
	return nil
}

func (tsfm *tsdbFMMap) lseek(offset int64) error {
	_, err := tsfm.file.Seek(offset, 0)
	if err != nil {
		common.Logger.Infof("lseek %s to %d failed:%s", tsfm.name, offset, err.Error())
		return err
	}
	return nil
}

func (tsfm *tsdbFMMap) batchRead(buf []byte) (int, error) {
	n, err := tsfm.file.Read(buf)
	if err != nil {
		if isTargetError(err, io.EOF) {
			return 0, nil
		}
		common.Logger.Infof("read %s n:%d, size:%d, failed:%s", tsfm.name, n, tsfm.size, err.Error())
		return 0, err
	}
	if n == 0 {
		return 0, gIsEof
	}
	return n, nil
}

func (tsfm *tsdbFMMap) batchReadAt(offset int64, buf []byte) (int, error) {
	if offset > tsfm.size {
		common.Logger.Infof("Read out of bound:%d > %d", offset, tsfm.size)
		return 0, errors.New("out of bound")
	}
	n, err := tsfm.file.ReadAt(buf, offset)
	if err != nil {
		if n == (int(tsfm.size-offset)) || (isTargetError(err, io.EOF)) {
			common.Logger.Debugf("read %s tail:%s", tsfm.name, err.Error())
			return n, nil
		}
		common.Logger.Infof("read %s failed:%s, n:%d", tsfm.name, err.Error(), n)
		return 0, err
	}
	if n == 0 {
		return 0, gIsEof
	}
	return n, nil
}

func (tsfm *tsdbFMMap) readAt(offset int64, len int64, msg TsData) error {
	if (offset + len) > tsfm.size {
		common.Logger.Infof("Read out of bound:%d > %d", (offset + len), tsfm.size)
		return errors.New("out of bound")
	}

	dat := make([]byte, len)
	_, err := tsfm.file.ReadAt(dat, offset)
	if err != nil {
		common.Logger.Infof("Read %s failed:%s", tsfm.name, err.Error())
		return err
	}

	err = msg.UnmarshalBinary(dat)
	if err != nil {
		common.Logger.Infof("Read %s unmarshal failed:%s", tsfm.name, err.Error())
		return err
	}

	return nil
}

func (tsfm *tsdbFMMap) append(msg TsData) error {
	out, err := msg.MarshalBinary()
	if err != nil {
		common.Logger.Infof("Append %s marshal failed:%s", tsfm.name, err.Error())
		return err
	}

	wl := int64(len(out))
	if (tsfm.size + wl) > tsfm.maxSize {
		common.Logger.Infof("Append %s out of bound:%d + %d > %d", tsfm.name, tsfm.size, wl, tsfm.maxSize)
		return gIsFull
	}

	err = tsfm.wcache.append(out)
	if err != nil {
		common.Logger.Infof("Append %s failed:%s", tsfm.name, err.Error())
		return err
	}

	tsfm.size += wl
	return nil
}

func (tsfm *tsdbFMMap) close() {
	if tsfm.wcache != nil {
		tsfm.wcache.flush()
		tsfm.wcache = nil
	}
	if tsfm.file != nil {
		tsfm.file.Close()
		tsfm.file = nil
	}
}

// tsdbBaReader
func (tr *tsdbBaReader) checkAndPut(pIdx *TsIndexData) bool {
	if tr.bufLen > 0 {
		if (pIdx.Block != tr.block) || (pIdx.Offset != (tr.offset + uint64(tr.bufLen))) {
			return false
		}
		if tr.bufLen+pIdx.Len > uint32(gIo_Size) {
			return false
		}
	}
	if tr.bufLen == 0 {
		tr.offset = pIdx.Offset
		tr.block = pIdx.Block
	}
	tr.bufLen += pIdx.Len
	tr.idxList.PushBack(pIdx)
	return true
}

func (tr *tsdbBaReader) init() {
	tr.bufLen = 0
	tr.offset = 0
	tr.idxList = list.New()
}

func (tr *tsdbBaReader) readAndRest(itemList *list.List) error {
	if tr.bufLen == 0 {
		return nil
	}
	dfs := &tsdbFile{maxSize: gDATA_FILE_SIZE, dir: tr.dir, block: int32(tr.block)}
	err := dfs.open(os.O_RDONLY, 0755)
	if err != nil {
		common.Logger.Infof("open block %d failed:%d", tr.block, err)
		return err
	}

	ioBuf := make([]byte, tr.bufLen)
	err = dfs.bReadAt(int64(tr.offset), ioBuf, int64(tr.bufLen))
	if err != nil {
		common.Logger.Infof("bReadAt failed:%s", err)
		dfs.close()
		tr.init()
		return err
	}
	offset := uint32(0)
	for offset < tr.bufLen {
		front := tr.idxList.Front()
		tr.idxList.Remove(front)
		pIdx := front.Value.(*TsIndexData)
		dat := make([]byte, pIdx.Len)
		tcopy(dat, ioBuf, int(offset))
		offset += pIdx.Len
		pDat := &tsdb.TsdbData{}
		err = proto.Unmarshal(dat, pDat)
		if err != nil {
			common.Logger.Infof("Unmarshal %d failed:%d", tr.block, err)
			return err
		}
		itemList.PushBack(pDat)
	}
	dfs.close()
	tr.init()
	return nil
}

func (tsf *tsFile) write(name string, msg proto.Message) error {
	if err := tsf.open(os.O_WRONLY|os.O_CREATE, name); err != nil {
		return err
	}
	defer tsf.file.Close()
	out, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	tsf.file.Write(out)
	return nil
}

func (tsf *tsFile) read(name string, msg proto.Message) error {
	if err := tsf.open(os.O_RDONLY, name); err != nil {
		return err
	}
	defer tsf.file.Close()
	info, err := tsf.file.Stat()
	if err != nil {
		return err
	}
	in := make([]byte, info.Size())
	_, err = tsf.file.Read(in)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(in, msg)
	return err
}

func (tsf *tsFile) open(flag int, name string) error {
	dir := fmt.Sprintf("%s/meta", common.Conf.Infra.FsDir)
	os.MkdirAll(dir, 0755)
	fname := fmt.Sprintf("%s/%s", dir, name)
	fout, err := os.OpenFile(fname, flag, 0755)
	if err != nil {
		return err
	}
	tsf.file = fout
	return nil
}
