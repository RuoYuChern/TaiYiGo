package infra

import (
	"bytes"
	"fmt"
)

type TsHeaderData struct {
	Items   uint32
	Version uint32
}

type TsMetaData struct {
	Start    uint64
	End      uint64
	Addr     int64
	Refblock int32
	Refitems int32
}

type TsIndexData struct {
	Timestamp uint64
	Offset    int64
	Block     int32
	Len       int32
}

func (tsh *TsHeaderData) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	fmt.Fprintln(&b, tsh.Items, tsh.Version)
	return b.Bytes(), nil
}

func (tsh *TsHeaderData) UnmarshalBinary(data []byte) error {
	b := bytes.NewBuffer(data)
	_, err := fmt.Fscanln(b, &tsh.Items, &tsh.Version)
	return err
}

func (tmd *TsMetaData) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	fmt.Fprintln(&b, tmd.Start, tmd.End, tmd.Addr, tmd.Refblock, tmd.Refitems)
	return b.Bytes(), nil
}

func (tmd *TsMetaData) UnmarshalBinary(data []byte) error {
	b := bytes.NewBuffer(data)
	_, err := fmt.Fscanln(b, &tmd.Start, &tmd.End, &tmd.Addr, &tmd.Refblock, &tmd.Refitems)
	return err
}

func (tsi *TsIndexData) MarshalBinary() ([]byte, error) {
	var b bytes.Buffer
	fmt.Fprintln(&b, tsi.Timestamp, tsi.Offset, tsi.Block, tsi.Len)
	return b.Bytes(), nil
}

func (tsi *TsIndexData) UnmarshalBinary(data []byte) error {
	b := bytes.NewBuffer(data)
	_, err := fmt.Fscanln(b, &tsi.Timestamp, &tsi.Offset, &tsi.Block, &tsi.Len)
	return err
}
