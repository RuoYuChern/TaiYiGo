package infra

import (
	"encoding/binary"
	"errors"
)

var (
	gTHD_LEN = int64(8)
	gTMD_LEN = int64(40)
	gTID_LEN = int64(24)
)

type TsData interface {
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(data []byte) error
}

type TsHeaderData struct {
	TsData
	Items   uint32
	Version uint32
}

type TsMetaData struct {
	TsData
	Start    uint64
	End      uint64
	Addr     uint64
	RefAddr  uint64
	Refblock uint32
	Refitems uint32
}

type TsIndexData struct {
	TsData
	Timestamp uint64
	Offset    uint64
	Block     uint32
	Len       uint32
}

func (tsh *TsHeaderData) MarshalBinary() ([]byte, error) {
	buf := make([]byte, gTHD_LEN)
	lwd := binary.LittleEndian
	lwd.PutUint32(buf, tsh.Items)
	lwd.PutUint32(buf[4:], tsh.Version)
	return buf, nil
}

func (tsh *TsHeaderData) UnmarshalBinary(data []byte) error {
	if len(data) < int(gTHD_LEN) {
		return errors.New("out of band")
	}
	lwd := binary.LittleEndian
	tsh.Items = lwd.Uint32(data)
	tsh.Version = lwd.Uint32(data[4:])
	return nil
}

func (tmd *TsMetaData) MarshalBinary() ([]byte, error) {
	buf := make([]byte, gTMD_LEN)
	lwd := binary.LittleEndian
	lwd.PutUint64(buf, tmd.Start)
	lwd.PutUint64(buf[8:], tmd.End)
	lwd.PutUint64(buf[16:], tmd.Addr)
	lwd.PutUint64(buf[24:], tmd.RefAddr)
	lwd.PutUint32(buf[32:], tmd.Refblock)
	lwd.PutUint32(buf[36:], tmd.Refitems)
	return buf, nil
}

func (tmd *TsMetaData) UnmarshalBinary(data []byte) error {
	if len(data) < int(gTMD_LEN) {
		return errors.New("out of band")
	}

	lwd := binary.LittleEndian
	tmd.Start = lwd.Uint64(data)
	tmd.End = lwd.Uint64(data[8:])
	tmd.Addr = lwd.Uint64(data[16:])
	tmd.RefAddr = lwd.Uint64(data[24:])
	tmd.Refblock = lwd.Uint32(data[32:])
	tmd.Refitems = lwd.Uint32(data[36:])
	return nil
}

func (tsi *TsIndexData) MarshalBinary() ([]byte, error) {
	buf := make([]byte, gTID_LEN)
	lwd := binary.LittleEndian
	lwd.PutUint64(buf, tsi.Timestamp)
	lwd.PutUint64(buf[8:], tsi.Offset)
	lwd.PutUint32(buf[16:], tsi.Block)
	lwd.PutUint32(buf[20:], tsi.Len)
	return buf, nil
}

func (tsi *TsIndexData) UnmarshalBinary(data []byte) error {
	if len(data) < int(gTID_LEN) {
		return errors.New("out of band")
	}
	lwd := binary.LittleEndian
	tsi.Timestamp = lwd.Uint64(data)
	tsi.Offset = lwd.Uint64(data[8:])
	tsi.Block = lwd.Uint32(data[16:])
	tsi.Len = lwd.Uint32(data[20:])
	return nil
}

func PutIntToB(data []byte, u uint32) {
	lwd := binary.LittleEndian
	lwd.PutUint32(data, u)
}

func GetIntFromB(data []byte) uint32 {
	lwd := binary.LittleEndian
	return lwd.Uint32(data)
}
