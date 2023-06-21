package infra

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"
	"taiyigo.com/common"
	"taiyigo.com/facade/tsdb"
)

const gDATA_FILE_SIZE = (1 << 30)
const gINDEX_FILE_SIZE = (200 << 20)
const gMETA_FILE_SIZE = (5 << 20)
const gDAT_FILE_TPL = "%s/dat_b_%d"
const gIDX_FILE_TPL = "%s/idx_b_%d"
const gDAT_PREFIX = "dat"
const gIo_Size = (2 << 20)
const gFlush_Size = (200 << 20)

type tsdbEfull struct {
}

type tsdbEEmpty struct {
}

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

type tsdbFMMap struct {
	maxSize int64
	name    string
	size    int64
	wcache  *tsdbIoBacth
	file    *os.File
}

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

// private functions
func newMapper(name string, maxSize int64) *tsdbFMMap {
	return &tsdbFMMap{maxSize: maxSize, name: name}
}

func newIoBatch(file *os.File) *tsdbIoBacth {
	iob := &tsdbIoBacth{buf: make([]byte, gIo_Size), offset: 0, file: file}
	return iob
}

func isError(err error, target error) bool {
	if err == nil {
		return false
	}
	return !errors.Is(err, target)
}

func isTargetError(err error, target error) bool {
	return errors.Is(err, target)
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

// Tsdb
func (tsdb *Tsdb) OpenAppender(id string) *TsdbAppender {
	tsdb.mu.Lock()
	defer tsdb.mu.Unlock()
	tstble, ok := tsdb.tblMap[id]
	if ok {
		return tstble
	}

	tstble = &TsdbAppender{id: id}
	tstble.Open()
	tsdb.tblMap[id] = tstble
	return tstble
}

func (tsdb *Tsdb) OpenQuery(id string) *TsdbQuery {
	tblQuery := &TsdbQuery{id: id}
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
func (tbl *TsdbAppender) Open() {
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
	err := tbl.getCurMeta()
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

func (tbl *TsdbAppender) getCurMeta() error {
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
	common.Logger.Infof("%s len:%d, cap:%d", tsfm.name, tsfm.size, tsfm.maxSize)
	return nil
}

func (tsfm *tsdbFMMap) writeAt(offset int64, msg TsData) error {
	out, err := msg.MarshalBinary()
	if err != nil {
		common.Logger.Infof("Write %s marshal failed:%s", tsfm.name, err.Error())
		return err
	}

	wl := int64(len(out))
	if (offset + wl) >= tsfm.maxSize {
		common.Logger.Infof("Write %s out of bound:%d >= %d", tsfm.name, (offset + wl), tsfm.maxSize)
		return &tsdbEfull{}
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

func (tsfm *tsdbFMMap) readAt(offset int64, len int64, msg TsData) error {
	if (offset + len) > tsfm.size {
		common.Logger.Infof("Read out of bound:%d >= %d", (offset + len), tsfm.size)
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
	if (tsfm.size + wl) >= tsfm.maxSize {
		common.Logger.Infof("Append %s out of bound:%d + %d >= %d", tsfm.name, tsfm.size, wl, tsfm.maxSize)
		return tsdbEfull{}
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
	return ((dataLen + tsf.size) >= tsf.maxSize)
}

func (tsf *tsdbFile) readAt(offset int64, len int, msg proto.Message) error {
	dat := make([]byte, len)
	n, err := tsf.file.ReadAt(dat, offset)
	if err != nil {
		common.Logger.Infof("Read %s failed:%s", tsf.name, err.Error())
		return err
	}
	if n != len {
		common.Logger.Infof("Read %s failed:%d != %d", tsf.name, n, len)
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

func (iob *tsdbIoBacth) append(dat []byte) error {
	dlen := len(dat)
	doff := 0
	for doff < dlen {
		rlen := gIo_Size - iob.offset
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

// tsdbFullError
func (tsfe tsdbEfull) Error() string {
	return "Is full"
}

// tsdbEEmpty
func (tsfee tsdbEEmpty) Error() string {
	return "Is Empty"
}
