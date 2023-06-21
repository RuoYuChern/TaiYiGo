package infra

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/huandu/skiplist"
	"google.golang.org/protobuf/proto"
	"taiyigo.com/common"
)

const gDATA_FILE_SIZE = (1 << 30)
const gINDEX_FILE_SIZE = (200 << 20)
const gMETA_FILE_SIZE = (5 << 20)
const gDAT_FILE_TPL = "%s/dat_b_%d"
const gIDX_FILE_TPL = "%s/idx_b_%d"
const gDAT_PREFIX = "dat"
const gIo_Size = (2 << 20)
const gFlush_Size = (200 << 20)
const gQuery_Size = (2 << 20)

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

type tsdbQIdx struct {
	idx *TsIndexData
	off int
}

// private functions
func newMapper(name string, maxSize int64) *tsdbFMMap {
	return &tsdbFMMap{maxSize: maxSize, name: name}
}

func newIoBatch(file *os.File) *tsdbIoBacth {
	iob := &tsdbIoBacth{buf: make([]byte, gIo_Size), offset: 0, file: file}
	return iob
}

func newSkipList() *skiplist.SkipList {
	list := skiplist.New(skiplist.GreaterThanFunc(func(lhs, rhs interface{}) int {
		s1 := lhs.(*tsdbQIdx)
		s2 := rhs.(*tsdbQIdx)
		if s1.off > s2.off {
			return 1
		} else if s1.off < s2.off {
			return -1
		}
		return 0
	}))
	return list
}

func calMetaIoSize(size int64, itemLen int64) int64 {
	var twoMB int64 = (2 << 20)
	if size <= twoMB {
		return size
	}

	items := (size / itemLen)
	bufSize := twoMB
	low := int64(0)
	high := items
	var oneMB int64 = (1 << 20)
	var threeMb int64 = (3 << 20)
	for low < high {
		mid := (low + high)
		midSize := (mid * itemLen)
		bufSize = midSize
		if midSize <= oneMB {
			low = mid
		} else if midSize >= threeMb {
			high = mid
		}
		break
	}

	return bufSize
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

// tsdbFullError
func (tsfe tsdbEfull) Error() string {
	return "Is full"
}

// tsdbEEmpty
func (tsfee tsdbEEmpty) Error() string {
	return "Is Empty"
}

// tsdbIoBacth
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

func (tsfm *tsdbFMMap) batchRdMeta(buf []byte, rdItems int64) {
	return
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
