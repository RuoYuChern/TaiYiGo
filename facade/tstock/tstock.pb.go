// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v4.23.3
// source: tstock.proto

package tstock

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Candle struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Period   uint64  `protobuf:"varint,1,opt,name=period,proto3" json:"period,omitempty"`
	Pcg      float64 `protobuf:"fixed64,2,opt,name=pcg,proto3" json:"pcg,omitempty"`
	Pcgp     float64 `protobuf:"fixed64,3,opt,name=pcgp,proto3" json:"pcgp,omitempty"`
	Open     float64 `protobuf:"fixed64,4,opt,name=open,proto3" json:"open,omitempty"`
	Close    float64 `protobuf:"fixed64,5,opt,name=close,proto3" json:"close,omitempty"`
	High     float64 `protobuf:"fixed64,6,opt,name=high,proto3" json:"high,omitempty"`
	Low      float64 `protobuf:"fixed64,7,opt,name=low,proto3" json:"low,omitempty"`
	Amount   float64 `protobuf:"fixed64,8,opt,name=amount,proto3" json:"amount,omitempty"`
	PreClose float64 `protobuf:"fixed64,9,opt,name=preClose,proto3" json:"preClose,omitempty"`
	Volume   uint32  `protobuf:"varint,10,opt,name=volume,proto3" json:"volume,omitempty"`
}

func (x *Candle) Reset() {
	*x = Candle{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tstock_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Candle) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Candle) ProtoMessage() {}

func (x *Candle) ProtoReflect() protoreflect.Message {
	mi := &file_tstock_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Candle.ProtoReflect.Descriptor instead.
func (*Candle) Descriptor() ([]byte, []int) {
	return file_tstock_proto_rawDescGZIP(), []int{0}
}

func (x *Candle) GetPeriod() uint64 {
	if x != nil {
		return x.Period
	}
	return 0
}

func (x *Candle) GetPcg() float64 {
	if x != nil {
		return x.Pcg
	}
	return 0
}

func (x *Candle) GetPcgp() float64 {
	if x != nil {
		return x.Pcgp
	}
	return 0
}

func (x *Candle) GetOpen() float64 {
	if x != nil {
		return x.Open
	}
	return 0
}

func (x *Candle) GetClose() float64 {
	if x != nil {
		return x.Close
	}
	return 0
}

func (x *Candle) GetHigh() float64 {
	if x != nil {
		return x.High
	}
	return 0
}

func (x *Candle) GetLow() float64 {
	if x != nil {
		return x.Low
	}
	return 0
}

func (x *Candle) GetAmount() float64 {
	if x != nil {
		return x.Amount
	}
	return 0
}

func (x *Candle) GetPreClose() float64 {
	if x != nil {
		return x.PreClose
	}
	return 0
}

func (x *Candle) GetVolume() uint32 {
	if x != nil {
		return x.Volume
	}
	return 0
}

type Crypto struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Period    uint64  `protobuf:"varint,1,opt,name=period,proto3" json:"period,omitempty"`
	Pcg       float64 `protobuf:"fixed64,2,opt,name=pcg,proto3" json:"pcg,omitempty"`
	Pcgp      float64 `protobuf:"fixed64,3,opt,name=pcgp,proto3" json:"pcgp,omitempty"`
	Open      float64 `protobuf:"fixed64,4,opt,name=open,proto3" json:"open,omitempty"`
	Close     float64 `protobuf:"fixed64,5,opt,name=close,proto3" json:"close,omitempty"`
	High      float64 `protobuf:"fixed64,6,opt,name=high,proto3" json:"high,omitempty"`
	Low       float64 `protobuf:"fixed64,7,opt,name=low,proto3" json:"low,omitempty"`
	Weight    float64 `protobuf:"fixed64,8,opt,name=weight,proto3" json:"weight,omitempty"`
	Vol       float64 `protobuf:"fixed64,9,opt,name=vol,proto3" json:"vol,omitempty"`
	Quotal    float64 `protobuf:"fixed64,10,opt,name=quotal,proto3" json:"quotal,omitempty"`
	EventTime uint64  `protobuf:"varint,11,opt,name=eventTime,proto3" json:"eventTime,omitempty"`
}

func (x *Crypto) Reset() {
	*x = Crypto{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tstock_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Crypto) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Crypto) ProtoMessage() {}

func (x *Crypto) ProtoReflect() protoreflect.Message {
	mi := &file_tstock_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Crypto.ProtoReflect.Descriptor instead.
func (*Crypto) Descriptor() ([]byte, []int) {
	return file_tstock_proto_rawDescGZIP(), []int{1}
}

func (x *Crypto) GetPeriod() uint64 {
	if x != nil {
		return x.Period
	}
	return 0
}

func (x *Crypto) GetPcg() float64 {
	if x != nil {
		return x.Pcg
	}
	return 0
}

func (x *Crypto) GetPcgp() float64 {
	if x != nil {
		return x.Pcgp
	}
	return 0
}

func (x *Crypto) GetOpen() float64 {
	if x != nil {
		return x.Open
	}
	return 0
}

func (x *Crypto) GetClose() float64 {
	if x != nil {
		return x.Close
	}
	return 0
}

func (x *Crypto) GetHigh() float64 {
	if x != nil {
		return x.High
	}
	return 0
}

func (x *Crypto) GetLow() float64 {
	if x != nil {
		return x.Low
	}
	return 0
}

func (x *Crypto) GetWeight() float64 {
	if x != nil {
		return x.Weight
	}
	return 0
}

func (x *Crypto) GetVol() float64 {
	if x != nil {
		return x.Vol
	}
	return 0
}

func (x *Crypto) GetQuotal() float64 {
	if x != nil {
		return x.Quotal
	}
	return 0
}

func (x *Crypto) GetEventTime() uint64 {
	if x != nil {
		return x.EventTime
	}
	return 0
}

type CnBasic struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Symbol     string `protobuf:"bytes,1,opt,name=symbol,proto3" json:"symbol,omitempty"`
	Name       string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Area       string `protobuf:"bytes,3,opt,name=area,proto3" json:"area,omitempty"`
	Industry   string `protobuf:"bytes,4,opt,name=industry,proto3" json:"industry,omitempty"`
	FulName    string `protobuf:"bytes,5,opt,name=fulName,proto3" json:"fulName,omitempty"`
	EnName     string `protobuf:"bytes,6,opt,name=enName,proto3" json:"enName,omitempty"`
	CnName     string `protobuf:"bytes,7,opt,name=cnName,proto3" json:"cnName,omitempty"`
	Market     string `protobuf:"bytes,8,opt,name=market,proto3" json:"market,omitempty"`
	ExChange   string `protobuf:"bytes,9,opt,name=exChange,proto3" json:"exChange,omitempty"`
	Status     string `protobuf:"bytes,10,opt,name=status,proto3" json:"status,omitempty"`
	ListDate   string `protobuf:"bytes,11,opt,name=listDate,proto3" json:"listDate,omitempty"`
	DelistDate string `protobuf:"bytes,12,opt,name=delistDate,proto3" json:"delistDate,omitempty"`
	IsHs       string `protobuf:"bytes,13,opt,name=isHs,proto3" json:"isHs,omitempty"`
}

func (x *CnBasic) Reset() {
	*x = CnBasic{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tstock_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CnBasic) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CnBasic) ProtoMessage() {}

func (x *CnBasic) ProtoReflect() protoreflect.Message {
	mi := &file_tstock_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CnBasic.ProtoReflect.Descriptor instead.
func (*CnBasic) Descriptor() ([]byte, []int) {
	return file_tstock_proto_rawDescGZIP(), []int{2}
}

func (x *CnBasic) GetSymbol() string {
	if x != nil {
		return x.Symbol
	}
	return ""
}

func (x *CnBasic) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *CnBasic) GetArea() string {
	if x != nil {
		return x.Area
	}
	return ""
}

func (x *CnBasic) GetIndustry() string {
	if x != nil {
		return x.Industry
	}
	return ""
}

func (x *CnBasic) GetFulName() string {
	if x != nil {
		return x.FulName
	}
	return ""
}

func (x *CnBasic) GetEnName() string {
	if x != nil {
		return x.EnName
	}
	return ""
}

func (x *CnBasic) GetCnName() string {
	if x != nil {
		return x.CnName
	}
	return ""
}

func (x *CnBasic) GetMarket() string {
	if x != nil {
		return x.Market
	}
	return ""
}

func (x *CnBasic) GetExChange() string {
	if x != nil {
		return x.ExChange
	}
	return ""
}

func (x *CnBasic) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *CnBasic) GetListDate() string {
	if x != nil {
		return x.ListDate
	}
	return ""
}

func (x *CnBasic) GetDelistDate() string {
	if x != nil {
		return x.DelistDate
	}
	return ""
}

func (x *CnBasic) GetIsHs() string {
	if x != nil {
		return x.IsHs
	}
	return ""
}

type StfInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Symbol string `protobuf:"bytes,1,opt,name=symbol,proto3" json:"symbol,omitempty"`
	Status string `protobuf:"bytes,2,opt,name=status,proto3" json:"status,omitempty"`
	Name   string `protobuf:"bytes,3,opt,name=name,proto3" json:"name,omitempty"`
	Opt    string `protobuf:"bytes,4,opt,name=opt,proto3" json:"opt,omitempty"`
	Day    uint64 `protobuf:"varint,5,opt,name=day,proto3" json:"day,omitempty"`
}

func (x *StfInfo) Reset() {
	*x = StfInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tstock_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StfInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StfInfo) ProtoMessage() {}

func (x *StfInfo) ProtoReflect() protoreflect.Message {
	mi := &file_tstock_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StfInfo.ProtoReflect.Descriptor instead.
func (*StfInfo) Descriptor() ([]byte, []int) {
	return file_tstock_proto_rawDescGZIP(), []int{3}
}

func (x *StfInfo) GetSymbol() string {
	if x != nil {
		return x.Symbol
	}
	return ""
}

func (x *StfInfo) GetStatus() string {
	if x != nil {
		return x.Status
	}
	return ""
}

func (x *StfInfo) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *StfInfo) GetOpt() string {
	if x != nil {
		return x.Opt
	}
	return ""
}

func (x *StfInfo) GetDay() uint64 {
	if x != nil {
		return x.Day
	}
	return 0
}

type CnBasicList struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Numbers     int32      `protobuf:"varint,1,opt,name=numbers,proto3" json:"numbers,omitempty"`
	CnBasicList []*CnBasic `protobuf:"bytes,2,rep,name=cnBasicList,proto3" json:"cnBasicList,omitempty"`
}

func (x *CnBasicList) Reset() {
	*x = CnBasicList{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tstock_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CnBasicList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CnBasicList) ProtoMessage() {}

func (x *CnBasicList) ProtoReflect() protoreflect.Message {
	mi := &file_tstock_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CnBasicList.ProtoReflect.Descriptor instead.
func (*CnBasicList) Descriptor() ([]byte, []int) {
	return file_tstock_proto_rawDescGZIP(), []int{4}
}

func (x *CnBasicList) GetNumbers() int32 {
	if x != nil {
		return x.Numbers
	}
	return 0
}

func (x *CnBasicList) GetCnBasicList() []*CnBasic {
	if x != nil {
		return x.CnBasicList
	}
	return nil
}

type StfList struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Numbers int32      `protobuf:"varint,1,opt,name=numbers,proto3" json:"numbers,omitempty"`
	Stfs    []*StfInfo `protobuf:"bytes,2,rep,name=stfs,proto3" json:"stfs,omitempty"`
}

func (x *StfList) Reset() {
	*x = StfList{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tstock_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StfList) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StfList) ProtoMessage() {}

func (x *StfList) ProtoReflect() protoreflect.Message {
	mi := &file_tstock_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StfList.ProtoReflect.Descriptor instead.
func (*StfList) Descriptor() ([]byte, []int) {
	return file_tstock_proto_rawDescGZIP(), []int{5}
}

func (x *StfList) GetNumbers() int32 {
	if x != nil {
		return x.Numbers
	}
	return 0
}

func (x *StfList) GetStfs() []*StfInfo {
	if x != nil {
		return x.Stfs
	}
	return nil
}

var File_tstock_proto protoreflect.FileDescriptor

var file_tstock_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x74, 0x73, 0x74, 0x6f, 0x63, 0x6b, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x74, 0x73, 0x74, 0x6f, 0x63, 0x6b, 0x22, 0xe2, 0x01, 0x0a, 0x06, 0x43, 0x61, 0x6e, 0x64, 0x6c,
	0x65, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x04, 0x52, 0x06, 0x70, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x70, 0x63, 0x67,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x01, 0x52, 0x03, 0x70, 0x63, 0x67, 0x12, 0x12, 0x0a, 0x04, 0x70,
	0x63, 0x67, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x01, 0x52, 0x04, 0x70, 0x63, 0x67, 0x70, 0x12,
	0x12, 0x0a, 0x04, 0x6f, 0x70, 0x65, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x01, 0x52, 0x04, 0x6f,
	0x70, 0x65, 0x6e, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6c, 0x6f, 0x73, 0x65, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x01, 0x52, 0x05, 0x63, 0x6c, 0x6f, 0x73, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x69, 0x67,
	0x68, 0x18, 0x06, 0x20, 0x01, 0x28, 0x01, 0x52, 0x04, 0x68, 0x69, 0x67, 0x68, 0x12, 0x10, 0x0a,
	0x03, 0x6c, 0x6f, 0x77, 0x18, 0x07, 0x20, 0x01, 0x28, 0x01, 0x52, 0x03, 0x6c, 0x6f, 0x77, 0x12,
	0x16, 0x0a, 0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x08, 0x20, 0x01, 0x28, 0x01, 0x52,
	0x06, 0x61, 0x6d, 0x6f, 0x75, 0x6e, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x70, 0x72, 0x65, 0x43, 0x6c,
	0x6f, 0x73, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x01, 0x52, 0x08, 0x70, 0x72, 0x65, 0x43, 0x6c,
	0x6f, 0x73, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x18, 0x0a, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x06, 0x76, 0x6f, 0x6c, 0x75, 0x6d, 0x65, 0x22, 0xf6, 0x01, 0x0a, 0x06,
	0x43, 0x72, 0x79, 0x70, 0x74, 0x6f, 0x12, 0x16, 0x0a, 0x06, 0x70, 0x65, 0x72, 0x69, 0x6f, 0x64,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x04, 0x52, 0x06, 0x70, 0x65, 0x72, 0x69, 0x6f, 0x64, 0x12, 0x10,
	0x0a, 0x03, 0x70, 0x63, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x01, 0x52, 0x03, 0x70, 0x63, 0x67,
	0x12, 0x12, 0x0a, 0x04, 0x70, 0x63, 0x67, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x01, 0x52, 0x04,
	0x70, 0x63, 0x67, 0x70, 0x12, 0x12, 0x0a, 0x04, 0x6f, 0x70, 0x65, 0x6e, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x01, 0x52, 0x04, 0x6f, 0x70, 0x65, 0x6e, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6c, 0x6f, 0x73,
	0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x01, 0x52, 0x05, 0x63, 0x6c, 0x6f, 0x73, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x68, 0x69, 0x67, 0x68, 0x18, 0x06, 0x20, 0x01, 0x28, 0x01, 0x52, 0x04, 0x68, 0x69,
	0x67, 0x68, 0x12, 0x10, 0x0a, 0x03, 0x6c, 0x6f, 0x77, 0x18, 0x07, 0x20, 0x01, 0x28, 0x01, 0x52,
	0x03, 0x6c, 0x6f, 0x77, 0x12, 0x16, 0x0a, 0x06, 0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x08,
	0x20, 0x01, 0x28, 0x01, 0x52, 0x06, 0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x10, 0x0a, 0x03,
	0x76, 0x6f, 0x6c, 0x18, 0x09, 0x20, 0x01, 0x28, 0x01, 0x52, 0x03, 0x76, 0x6f, 0x6c, 0x12, 0x16,
	0x0a, 0x06, 0x71, 0x75, 0x6f, 0x74, 0x61, 0x6c, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x01, 0x52, 0x06,
	0x71, 0x75, 0x6f, 0x74, 0x61, 0x6c, 0x12, 0x1c, 0x0a, 0x09, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x54,
	0x69, 0x6d, 0x65, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x04, 0x52, 0x09, 0x65, 0x76, 0x65, 0x6e, 0x74,
	0x54, 0x69, 0x6d, 0x65, 0x22, 0xcb, 0x02, 0x0a, 0x07, 0x43, 0x6e, 0x42, 0x61, 0x73, 0x69, 0x63,
	0x12, 0x16, 0x0a, 0x06, 0x73, 0x79, 0x6d, 0x62, 0x6f, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x73, 0x79, 0x6d, 0x62, 0x6f, 0x6c, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04,
	0x61, 0x72, 0x65, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x61, 0x72, 0x65, 0x61,
	0x12, 0x1a, 0x0a, 0x08, 0x69, 0x6e, 0x64, 0x75, 0x73, 0x74, 0x72, 0x79, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x08, 0x69, 0x6e, 0x64, 0x75, 0x73, 0x74, 0x72, 0x79, 0x12, 0x18, 0x0a, 0x07,
	0x66, 0x75, 0x6c, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x66,
	0x75, 0x6c, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x65, 0x6e, 0x4e, 0x61, 0x6d, 0x65,
	0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x65, 0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x16,
	0x0a, 0x06, 0x63, 0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06,
	0x63, 0x6e, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x61, 0x72, 0x6b, 0x65, 0x74,
	0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x6d, 0x61, 0x72, 0x6b, 0x65, 0x74, 0x12, 0x1a,
	0x0a, 0x08, 0x65, 0x78, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x08, 0x65, 0x78, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x6c, 0x69, 0x73, 0x74, 0x44, 0x61, 0x74, 0x65, 0x18, 0x0b,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x6c, 0x69, 0x73, 0x74, 0x44, 0x61, 0x74, 0x65, 0x12, 0x1e,
	0x0a, 0x0a, 0x64, 0x65, 0x6c, 0x69, 0x73, 0x74, 0x44, 0x61, 0x74, 0x65, 0x18, 0x0c, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0a, 0x64, 0x65, 0x6c, 0x69, 0x73, 0x74, 0x44, 0x61, 0x74, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x69, 0x73, 0x48, 0x73, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x69, 0x73,
	0x48, 0x73, 0x22, 0x71, 0x0a, 0x07, 0x53, 0x74, 0x66, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x16, 0x0a,
	0x06, 0x73, 0x79, 0x6d, 0x62, 0x6f, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73,
	0x79, 0x6d, 0x62, 0x6f, 0x6c, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x10, 0x0a, 0x03, 0x6f, 0x70, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03,
	0x6f, 0x70, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x64, 0x61, 0x79, 0x18, 0x05, 0x20, 0x01, 0x28, 0x04,
	0x52, 0x03, 0x64, 0x61, 0x79, 0x22, 0x5a, 0x0a, 0x0b, 0x43, 0x6e, 0x42, 0x61, 0x73, 0x69, 0x63,
	0x4c, 0x69, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x12, 0x31,
	0x0a, 0x0b, 0x63, 0x6e, 0x42, 0x61, 0x73, 0x69, 0x63, 0x4c, 0x69, 0x73, 0x74, 0x18, 0x02, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x74, 0x73, 0x74, 0x6f, 0x63, 0x6b, 0x2e, 0x43, 0x6e, 0x42,
	0x61, 0x73, 0x69, 0x63, 0x52, 0x0b, 0x63, 0x6e, 0x42, 0x61, 0x73, 0x69, 0x63, 0x4c, 0x69, 0x73,
	0x74, 0x22, 0x48, 0x0a, 0x07, 0x53, 0x74, 0x66, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07,
	0x6e, 0x75, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07, 0x6e,
	0x75, 0x6d, 0x62, 0x65, 0x72, 0x73, 0x12, 0x23, 0x0a, 0x04, 0x73, 0x74, 0x66, 0x73, 0x18, 0x02,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x0f, 0x2e, 0x74, 0x73, 0x74, 0x6f, 0x63, 0x6b, 0x2e, 0x53, 0x74,
	0x66, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x04, 0x73, 0x74, 0x66, 0x73, 0x42, 0x0f, 0x5a, 0x0d, 0x66,
	0x61, 0x63, 0x61, 0x64, 0x65, 0x2f, 0x74, 0x73, 0x74, 0x6f, 0x63, 0x6b, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_tstock_proto_rawDescOnce sync.Once
	file_tstock_proto_rawDescData = file_tstock_proto_rawDesc
)

func file_tstock_proto_rawDescGZIP() []byte {
	file_tstock_proto_rawDescOnce.Do(func() {
		file_tstock_proto_rawDescData = protoimpl.X.CompressGZIP(file_tstock_proto_rawDescData)
	})
	return file_tstock_proto_rawDescData
}

var file_tstock_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_tstock_proto_goTypes = []interface{}{
	(*Candle)(nil),      // 0: tstock.Candle
	(*Crypto)(nil),      // 1: tstock.Crypto
	(*CnBasic)(nil),     // 2: tstock.CnBasic
	(*StfInfo)(nil),     // 3: tstock.StfInfo
	(*CnBasicList)(nil), // 4: tstock.CnBasicList
	(*StfList)(nil),     // 5: tstock.StfList
}
var file_tstock_proto_depIdxs = []int32{
	2, // 0: tstock.CnBasicList.cnBasicList:type_name -> tstock.CnBasic
	3, // 1: tstock.StfList.stfs:type_name -> tstock.StfInfo
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_tstock_proto_init() }
func file_tstock_proto_init() {
	if File_tstock_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_tstock_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Candle); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_tstock_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Crypto); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_tstock_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CnBasic); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_tstock_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StfInfo); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_tstock_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CnBasicList); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_tstock_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StfList); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_tstock_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_tstock_proto_goTypes,
		DependencyIndexes: file_tstock_proto_depIdxs,
		MessageInfos:      file_tstock_proto_msgTypes,
	}.Build()
	File_tstock_proto = out.File
	file_tstock_proto_rawDesc = nil
	file_tstock_proto_goTypes = nil
	file_tstock_proto_depIdxs = nil
}
