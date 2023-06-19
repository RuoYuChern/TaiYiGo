package infra

import (
	"errors"
	"fmt"
	"os"

	"github.com/edsrzf/mmap-go"
	"google.golang.org/protobuf/proto"
	"taiyigo.com/common"
	"taiyigo.com/facade/tsdb"
)

const DATA_FILE_SIZE = (1 << 30)
const INDEX_FILE_SIZE = (200 << 20)
const META_FILE_SIZE = (5 << 20)

type tsdbEfull struct {
}

type TsdbFile struct {
	maxSize int64
	name    string
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
	metaLen    int64
	idxLen     int64
	headLen    int64
	meta       *TsdbFMMap
	widx       *TsdbFMMap
	wdat       *TsdbFile
	curMeta    *tsdb.TsdbMeta
	metaHeader *tsdb.TsdbHeader
}

// funcitons
func getItemLen(item proto.Message) (int64, error) {
	out, err := proto.Marshal(item)
	if err != nil {
		common.Logger.Warnf("marshal item failed:%s", err.Error())
		return -1, err
	}

	return int64(len(out)), nil
}

// TsdbTable
func (tstble *TsdbTable) Open() {
	dir := fmt.Sprintf("%s/tsdb/%s", common.Conf.Infra.FsDir, tstble.id)
	os.MkdirAll(dir, 0755)
	// 获取meta 和 idx 大小
	mt := &tsdb.TsdbMeta{Start: 0, End: 0, Block: 0, Offset: 0, Items: 0}
	tstble.metaLen, _ = getItemLen(mt)

	idx := &tsdb.TsdbIndex{Timestamp: 0, Block: 0, Offset: 0, Len: 0}
	tstble.idxLen, _ = getItemLen(idx)

	hd := &tsdb.TsdbHeader{Items: 0, Version: 0}
	tstble.headLen, _ = getItemLen(hd)
}

func (tstble *TsdbTable) Append(data *tsdb.TsdbData) error {
	meta := tstble.getCurMeta()
	if data.Timestamp <= meta.End {
		return errors.New("data is old")
	}
	return nil
}

func (tstble *TsdbTable) getCurMeta() *tsdb.TsdbMeta {
	if tstble.curMeta != nil {
		return tstble.curMeta
	}
	//
	name := fmt.Sprintf("%s/tsdb/%s/meta.dat", common.Conf.Infra.FsDir, tstble.id)
	tstble.meta = &TsdbFMMap{maxSize: META_FILE_SIZE, name: name}
	err := tstble.meta.Open()
	if err != nil {
		common.Logger.Infof("fmap open %s failed:%s", name, err.Error())
		return nil
	}
	//
	tstble.metaHeader = &tsdb.TsdbHeader{Items: 0, Version: 0}
	if tstble.meta.size >= tstble.headLen {
		err = tstble.meta.ReadAt(0, tstble.headLen, tstble.metaHeader)
	} else {
		//写入meta
		err = tstble.meta.WriteAt(0, tstble.metaHeader)
	}

	if err != nil {
		common.Logger.Infof("meta header opt %s failed:%s", name, err.Error())
		return nil
	}

	return tstble.curMeta
}

// TsdbFMMap
func (tsfm *TsdbFMMap) Open() error {
	f, err := os.OpenFile(tsfm.name, os.O_RDWR|os.O_CREATE, 0755)
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
		common.Logger.Info("Read out of bound:%d >= %d", (offset + len), tsfm.size)
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
		return &tsdbEfull{}
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

func (tsf *TsdbFile) Append(msg proto.Message) error {
	out, err := proto.Marshal(msg)
	if err != nil {
		common.Logger.Infof("Append %s marshal failed:%s", tsf.name, err.Error())
		return err
	}

	dataLen := int64(len(out))
	if tsf.isFull(dataLen) {
		common.Logger.Infof("Append %s failed: is full", tsf.name)
		return &tsdbEfull{}
	}

	n, err := tsf.file.Write(out)
	if err != nil {
		common.Logger.Infof("Append %s failed:%s", tsf.name, err.Error())
		return err
	}

	tsf.size += int64(n)
	return nil

}

// tsdbFullError
func (tsfe *tsdbEfull) Error() string {
	return "Is full"
}
