package infra

import (
	"errors"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/edsrzf/mmap-go"
	"google.golang.org/protobuf/proto"
	"taiyigo.com/common"
	"taiyigo.com/facade/tsdb"
)

const gDATA_FILE_SIZE = (1 << 30)
const gINDEX_FILE_SIZE = (200 << 20)
const gMETA_FILE_SIZE = (5 << 20)
const gDAT_FILE_TPL = "%s/data_%d.dat"
const gIDX_FILE_TPL = "%s/idx_%d.dat"
const gDAT_PREFIX = "data"

//const gIDX_PREFIX = "idx"

type tsdbEfull struct {
}

type tsdbEEmpty struct {
}

type TsdbFile struct {
	maxSize int64
	dir     string
	name    string
	block   int32
	size    int64
	file    *os.File
}

type TsdbFMMap struct {
	maxSize int64
	name    string
	size    int64
	file    *os.File
	fmap    mmap.MMap
}

type TsdbTable struct {
	id         string
	dir        string
	metaLen    int64
	idxLen     int64
	headLen    int64
	meta       *TsdbFMMap
	widx       *TsdbFMMap
	wdat       *TsdbFile
	curMeta    *tsdb.TsdbMeta
	metaHeader *tsdb.TsdbHeader
}

type Tsdb struct {
	tblMap map[string]*TsdbTable
	mu     sync.Mutex
}

var gtsdb *Tsdb
var gtsdbOnce sync.Once

// funcitons
func Gettsdb() *Tsdb {
	gtsdbOnce.Do(func() {
		gtsdb = &Tsdb{tblMap: make(map[string]*TsdbTable)}
	})
	return gtsdb
}

func getItemLen(item proto.Message) (int64, error) {
	out, err := proto.Marshal(item)
	if err != nil {
		common.Logger.Warnf("marshal item failed:%s", err.Error())
		return -1, err
	}
	return int64(len(out)), nil
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
		parts := strings.SplitAfterN(d.Name(), "_", 2)
		block, err := strconv.Atoi(parts[1])
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
func (tsdb *Tsdb) OpenTable(id string) *TsdbTable {
	tsdb.mu.Lock()
	defer tsdb.mu.Unlock()
	tstble, ok := tsdb.tblMap[id]
	if ok {
		return tstble
	}

	tstble = &TsdbTable{id: id}
	tstble.Open()
	tsdb.tblMap[id] = tstble
	return tstble
}

func (tsdb *Tsdb) Close() {
	for _, v := range tsdb.tblMap {
		v.Close()
	}

	tsdb.tblMap = make(map[string]*TsdbTable)
}

func (tsdb *Tsdb) remove(id string) {
	tsdb.mu.Lock()
	defer tsdb.mu.Unlock()
	delete(tsdb.tblMap, id)
}

// TsdbTable
func (tstble *TsdbTable) Open() {

	tstble.dir = fmt.Sprintf("%s/tsdb/%s", common.Conf.Infra.FsDir, tstble.id)
	os.MkdirAll(tstble.dir, 0755)
	// 获取meta 和 idx 大小
	// 这里有个坑:数值类型的如果赋值为0,序列化的时候不会被序列化，导致长度为0
	mt := &tsdb.TsdbMeta{Start: 1, End: 1, Addr: 1, Refblock: 1, Refitems: 1}
	tstble.metaLen, _ = getItemLen(mt)

	idx := &tsdb.TsdbIndex{Timestamp: 1, Block: 1, Offset: 1, Len: 1}
	tstble.idxLen, _ = getItemLen(idx)

	hd := &tsdb.TsdbHeader{Items: 1, Version: 1}
	tstble.headLen, _ = getItemLen(hd)
	common.Logger.Infof("metaLen:%d, idxLen:%d, headLen:%d", tstble.metaLen, tstble.idxLen, tstble.headLen)
}

func (tstble *TsdbTable) Close() {
	if tstble.wdat != nil {
		//关闭wdat
		tstble.wdat.Close()
		tstble.wdat = nil
	}
	if tstble.widx != nil {
		//关闭widx
		tstble.widx.Close()
		tstble.widx = nil
	}

	//保存元数据
	if tstble.metaHeader != nil {
		tstble.meta.WriteAt(0, tstble.metaHeader)
		tstble.metaHeader = nil
	}

	if tstble.curMeta != nil {
		tstble.meta.WriteAt(tstble.curMeta.Addr, tstble.curMeta)
		tstble.curMeta = nil
	}

	if tstble.meta != nil {
		// 关闭meta
		tstble.meta.Close()
	}

	Gettsdb().remove(tstble.id)
}

func (tstble *TsdbTable) Append(data *tsdb.TsdbData) error {
	err := tstble.getCurMeta()
	if isError(err, tsdbEEmpty{}) {
		return err
	}
	// 初始化
	if tstble.curMeta == nil {
		tstble.curMeta = &tsdb.TsdbMeta{Start: data.Timestamp, End: math.MaxInt64, Addr: tstble.headLen, Refblock: 0, Refitems: 0}
	}

	if data.Timestamp < tstble.curMeta.Start {
		return errors.New("data is old")
	}
	//
	err = tstble.getWrite()
	if err != nil {
		return err
	}

	ref, err := tstble.wdat.Append(data)
	if err != nil {
		return err
	}
	//填充时间
	ref.Timestamp = data.Timestamp
	// 保存index
	err = tstble.saveIndex(ref)
	return err
}

func (tstble *TsdbTable) saveIndex(tidx *tsdb.TsdbIndex) error {
	if tstble.widx == nil {
		//打开IDX
		name := fmt.Sprintf(gIDX_FILE_TPL, tstble.dir, tstble.curMeta.Refblock)
		tstble.widx = &TsdbFMMap{maxSize: gINDEX_FILE_SIZE, name: name}
		err := tstble.widx.Open()
		if err != nil {
			common.Logger.Infof("open %s failed:%s", name, err.Error())
			return err
		}
	}

	err := tstble.widx.Append(tidx)
	if isError(err, tsdbEfull{}) {
		common.Logger.Infof("Append idx_%d failed:%s", tstble.curMeta.Refblock, err.Error())
		return err
	}

	// 之前的idx文件已经写满了，换个新的
	if isTargetError(err, tsdbEfull{}) {
		//将当前保存
		tstble.curMeta.End = tidx.Timestamp
		err = tstble.meta.WriteAt(tstble.curMeta.Addr, tstble.curMeta)
		if err != nil {
			common.Logger.Infof("Save ref, write meta failed:%s", err.Error())
			return err
		}
		//计算当前的headlen
		items := (tstble.curMeta.Addr-tstble.headLen)/tstble.metaLen + 1
		if items > int64(tstble.metaHeader.Items) {
			tstble.metaHeader.Items = uint32(items)
			tstble.metaHeader.Version = tstble.metaHeader.Version + 1
		}
		// 换新的idx
		tstble.widx.Close()
		name := fmt.Sprintf(gIDX_FILE_TPL, tstble.dir, tstble.curMeta.Refblock+1)
		tstble.widx = &TsdbFMMap{maxSize: gINDEX_FILE_SIZE, name: name}
		err = tstble.widx.Open()
		if err != nil {
			common.Logger.Infof("open %s failed:%s", name, err.Error())
			return err
		}
		err = tstble.widx.Append(tidx)
		if err != nil {
			common.Logger.Infof("append %s failed:%s", name, err.Error())
			return err
		}
		// 获取新的metaCure
		addr := tstble.curMeta.Addr + tstble.metaLen
		refBlock := tstble.curMeta.Refblock + 1
		tstble.curMeta = &tsdb.TsdbMeta{Start: tidx.Timestamp, End: math.MaxInt64, Addr: addr, Refblock: refBlock, Refitems: 1}
		// 更新header
		tstble.metaHeader.Items = tstble.metaHeader.Items + 1
		tstble.metaHeader.Version = tstble.metaHeader.Version + 1
		tstble.meta.WriteAt(0, tstble.metaHeader)
	} else {
		tstble.curMeta.Refitems = tstble.curMeta.Refitems + 1
	}
	err = tstble.meta.WriteAt(tstble.curMeta.Addr, tstble.curMeta)
	return err
}

func (tstble *TsdbTable) getWrite() error {
	if tstble.wdat != nil {
		return nil
	}

	block, err := getTargetBlockNo(tstble.dir, gDAT_PREFIX)
	if err != nil {
		return nil
	}

	tstble.wdat = &TsdbFile{maxSize: gDATA_FILE_SIZE, dir: tstble.dir, block: int32(block)}
	return tstble.wdat.Open()
}

func (tstble *TsdbTable) getCurMeta() error {
	if tstble.curMeta != nil {
		return nil
	}
	//
	name := fmt.Sprintf("%s/tsdb/%s/meta.dat", common.Conf.Infra.FsDir, tstble.id)
	tstble.meta = &TsdbFMMap{maxSize: gMETA_FILE_SIZE, name: name}
	err := tstble.meta.Open()
	if err != nil {
		common.Logger.Infof("fmap open %s failed:%s", name, err.Error())
		return err
	}
	//
	tstble.metaHeader = &tsdb.TsdbHeader{Items: 0, Version: 0}
	if tstble.meta.size >= tstble.headLen {
		common.Logger.Infof("meta size:%d, headLen:%d", tstble.meta.size, tstble.headLen)
		err = tstble.meta.ReadAt(0, tstble.headLen, tstble.metaHeader)
	} else {
		//写入meta
		err = tstble.meta.WriteAt(0, tstble.metaHeader)
	}

	if err != nil {
		common.Logger.Infof("meta header opt %s failed:%s", name, err.Error())
		return err
	}

	if tstble.metaHeader.Items == 0 {
		return tsdbEEmpty{}
	}

	tstble.curMeta = &tsdb.TsdbMeta{}
	// 获取最新的一个metadata
	offset := tstble.headLen + (int64(tstble.metaHeader.Items-1) * tstble.metaLen)
	//tstble.curMeta.Addr = offset
	err = tstble.meta.ReadAt(offset, tstble.metaLen, tstble.curMeta)
	if err != nil {
		common.Logger.Infof("meta read %s failed:%s", name, err.Error())
		return err
	}

	return nil
}

// TsdbFMMap
func (tsfm *TsdbFMMap) Open() error {
	f, err := os.OpenFile(tsfm.name, os.O_RDWR|os.O_CREATE, 0644)
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
	if info.Size() == 0 {
		//map 要初始化文件大小
		err = f.Truncate(tsfm.maxSize)
		if err != nil {
			common.Logger.Warnf("open file %s truncate failed:%s", tsfm.name, err.Error())
			//关闭文件
			f.Close()
			return err
		}
	}

	tsfm.file = f
	tsfm.size = info.Size()
	tsfm.fmap, err = mmap.Map(tsfm.file, mmap.RDWR, 0)
	if err != nil {
		common.Logger.Warnf("open file %s map failed:%s", tsfm.name, err.Error())
		//关闭文件
		f.Close()
		return err
	}

	common.Logger.Infof("%s len:%d", tsfm.name, len(tsfm.fmap))

	return nil
}

func (tsfm *TsdbFMMap) Lock() {
	err := tsfm.fmap.Lock()
	if err != nil {
		common.Logger.Infof("lock file:%s failed:%s", tsfm.name, err.Error())
		return
	}
}

func (tsfm *TsdbFMMap) Unlock() {
	err := tsfm.fmap.Unlock()
	if err != nil {
		common.Logger.Infof("unlock file:%s failed:%s", tsfm.name, err.Error())
		return
	}
}

func (tsfm *TsdbFMMap) WriteAt(offset int64, msg proto.Message) error {
	out, err := proto.Marshal(msg)
	if err != nil {
		common.Logger.Infof("Write %s marshal failed:%s", tsfm.name, err.Error())
		return err
	}

	wl := int64(len(out))
	if (offset + wl) >= tsfm.maxSize {
		common.Logger.Info("Write out of bound:%d >= %d", (offset + wl), tsfm.size)
		return &tsdbEfull{}
	}

	copy(tsfm.fmap[offset:], out)
	if (offset + wl) > tsfm.size {
		tsfm.size = (offset + wl)
	}
	return nil
}

func (tsfm *TsdbFMMap) ReadAt(offset int64, len int64, msg proto.Message) error {
	if (offset + len) >= tsfm.size {
		common.Logger.Infof("Read out of bound:%d >= %d", (offset + len), tsfm.size)
		return errors.New("out of bound")
	}
	dat := make([]byte, len)
	copy(dat, tsfm.fmap)
	err := proto.Unmarshal(dat, msg)
	if err != nil {
		common.Logger.Infof("Unmarshal %s failed:%s", tsfm.name, err.Error())
		return err
	}

	return nil
}

func (tsfm *TsdbFMMap) Append(msg proto.Message) error {
	out, err := proto.Marshal(msg)
	if err != nil {
		common.Logger.Infof("Append %s marshal failed:%s", tsfm.name, err.Error())
		return err
	}

	wl := int64(len(out))
	if (tsfm.size + wl) >= tsfm.maxSize {
		common.Logger.Info("Write out of bound:%d >= %d", (tsfm.size + wl), tsfm.size)
		return tsdbEfull{}
	}
	copy(tsfm.fmap[tsfm.size:], out)
	tsfm.size += wl
	return nil
}

func (tsfm *TsdbFMMap) Flush() {
	tsfm.fmap.Flush()
}

func (tsfm *TsdbFMMap) Close() {
	if tsfm.fmap != nil {
		tsfm.fmap.Flush()
		tsfm.fmap.Unmap()
	}

	if tsfm.file != nil {
		tsfm.file.Close()
	}
}

// TsdbFile
func (tsf *TsdbFile) Open() error {
	tsf.name = fmt.Sprintf(gDAT_FILE_TPL, tsf.dir, tsf.block)
	file, err := os.OpenFile(tsf.name, os.O_RDWR|os.O_CREATE, 0755)
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
	return nil
}

func (tsf *TsdbFile) Close() {
	if tsf.file != nil {
		tsf.file.Close()
		tsf.file = nil
	}
}

func (tsf *TsdbFile) isFull(dataLen int64) bool {
	return ((dataLen + tsf.size) >= tsf.maxSize)
}

func (tsf *TsdbFile) ReadAt(offset int64, len int, msg proto.Message) error {
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

func (tsf *TsdbFile) Append(msg proto.Message) (*tsdb.TsdbIndex, error) {
	out, err := proto.Marshal(msg)
	if err != nil {
		common.Logger.Infof("Append %s marshal failed:%s", tsf.name, err.Error())
		return nil, err
	}

	dataLen := int64(len(out))
	if tsf.isFull(dataLen) {
		//重新打开
		err = tsf.openNew()
		if err != nil {
			return nil, err
		}
	}

	n, err := tsf.file.Write(out)
	if err != nil {
		common.Logger.Infof("Append %s failed:%s", tsf.name, err.Error())
		return nil, err
	}

	ref := &tsdb.TsdbIndex{Block: tsf.block, Offset: tsf.size, Len: int32(dataLen)}
	tsf.size += int64(n)
	return ref, nil

}

func (tsf *TsdbFile) openNew() error {
	tsf.Close()
	tsf.block = tsf.block + 1
	err := tsf.Open()
	if err != nil {
		common.Logger.Info("reopen failed")
	}
	return err
}

// tsdbFullError
func (tsfe tsdbEfull) Error() string {
	return "Is full"
}

// tsdbEEmpty
func (tsfee tsdbEEmpty) Error() string {
	return "Is Empty"
}
