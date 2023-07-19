// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v4.23.3
// source: tsorder.proto

package tsorder

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

type TOrder struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Symbol     string  `protobuf:"bytes,1,opt,name=symbol,proto3" json:"symbol,omitempty"`
	Name       string  `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	CreatDay   string  `protobuf:"bytes,3,opt,name=creatDay,proto3" json:"creatDay,omitempty"`
	BuyDay     string  `protobuf:"bytes,4,opt,name=buyDay,proto3" json:"buyDay,omitempty"`
	SellDay    string  `protobuf:"bytes,5,opt,name=sellDay,proto3" json:"sellDay,omitempty"`
	EndDay     string  `protobuf:"bytes,6,opt,name=endDay,proto3" json:"endDay,omitempty"`
	Buyer      string  `protobuf:"bytes,7,opt,name=buyer,proto3" json:"buyer,omitempty"`
	OrderId    string  `protobuf:"bytes,8,opt,name=orderId,proto3" json:"orderId,omitempty"`
	Status     int32   `protobuf:"varint,9,opt,name=status,proto3" json:"status,omitempty"`
	Vol        int32   `protobuf:"varint,10,opt,name=vol,proto3" json:"vol,omitempty"`
	OrderPrice float32 `protobuf:"fixed32,11,opt,name=orderPrice,proto3" json:"orderPrice,omitempty"`
	BuyPrice   float32 `protobuf:"fixed32,12,opt,name=buyPrice,proto3" json:"buyPrice,omitempty"`
	SellPrice  float32 `protobuf:"fixed32,13,opt,name=sellPrice,proto3" json:"sellPrice,omitempty"`
}

func (x *TOrder) Reset() {
	*x = TOrder{}
	if protoimpl.UnsafeEnabled {
		mi := &file_tsorder_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TOrder) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TOrder) ProtoMessage() {}

func (x *TOrder) ProtoReflect() protoreflect.Message {
	mi := &file_tsorder_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TOrder.ProtoReflect.Descriptor instead.
func (*TOrder) Descriptor() ([]byte, []int) {
	return file_tsorder_proto_rawDescGZIP(), []int{0}
}

func (x *TOrder) GetSymbol() string {
	if x != nil {
		return x.Symbol
	}
	return ""
}

func (x *TOrder) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *TOrder) GetCreatDay() string {
	if x != nil {
		return x.CreatDay
	}
	return ""
}

func (x *TOrder) GetBuyDay() string {
	if x != nil {
		return x.BuyDay
	}
	return ""
}

func (x *TOrder) GetSellDay() string {
	if x != nil {
		return x.SellDay
	}
	return ""
}

func (x *TOrder) GetEndDay() string {
	if x != nil {
		return x.EndDay
	}
	return ""
}

func (x *TOrder) GetBuyer() string {
	if x != nil {
		return x.Buyer
	}
	return ""
}

func (x *TOrder) GetOrderId() string {
	if x != nil {
		return x.OrderId
	}
	return ""
}

func (x *TOrder) GetStatus() int32 {
	if x != nil {
		return x.Status
	}
	return 0
}

func (x *TOrder) GetVol() int32 {
	if x != nil {
		return x.Vol
	}
	return 0
}

func (x *TOrder) GetOrderPrice() float32 {
	if x != nil {
		return x.OrderPrice
	}
	return 0
}

func (x *TOrder) GetBuyPrice() float32 {
	if x != nil {
		return x.BuyPrice
	}
	return 0
}

func (x *TOrder) GetSellPrice() float32 {
	if x != nil {
		return x.SellPrice
	}
	return 0
}

var File_tsorder_proto protoreflect.FileDescriptor

var file_tsorder_proto_rawDesc = []byte{
	0x0a, 0x0d, 0x74, 0x73, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x07, 0x74, 0x73, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x22, 0xce, 0x02, 0x0a, 0x06, 0x54, 0x4f, 0x72,
	0x64, 0x65, 0x72, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x79, 0x6d, 0x62, 0x6f, 0x6c, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x73, 0x79, 0x6d, 0x62, 0x6f, 0x6c, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12,
	0x1a, 0x0a, 0x08, 0x63, 0x72, 0x65, 0x61, 0x74, 0x44, 0x61, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x63, 0x72, 0x65, 0x61, 0x74, 0x44, 0x61, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x62,
	0x75, 0x79, 0x44, 0x61, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x62, 0x75, 0x79,
	0x44, 0x61, 0x79, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x65, 0x6c, 0x6c, 0x44, 0x61, 0x79, 0x18, 0x05,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x65, 0x6c, 0x6c, 0x44, 0x61, 0x79, 0x12, 0x16, 0x0a,
	0x06, 0x65, 0x6e, 0x64, 0x44, 0x61, 0x79, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x65,
	0x6e, 0x64, 0x44, 0x61, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x62, 0x75, 0x79, 0x65, 0x72, 0x18, 0x07,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x62, 0x75, 0x79, 0x65, 0x72, 0x12, 0x18, 0x0a, 0x07, 0x6f,
	0x72, 0x64, 0x65, 0x72, 0x49, 0x64, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6f, 0x72,
	0x64, 0x65, 0x72, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18,
	0x09, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x10, 0x0a,
	0x03, 0x76, 0x6f, 0x6c, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x05, 0x52, 0x03, 0x76, 0x6f, 0x6c, 0x12,
	0x1e, 0x0a, 0x0a, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x50, 0x72, 0x69, 0x63, 0x65, 0x18, 0x0b, 0x20,
	0x01, 0x28, 0x02, 0x52, 0x0a, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x50, 0x72, 0x69, 0x63, 0x65, 0x12,
	0x1a, 0x0a, 0x08, 0x62, 0x75, 0x79, 0x50, 0x72, 0x69, 0x63, 0x65, 0x18, 0x0c, 0x20, 0x01, 0x28,
	0x02, 0x52, 0x08, 0x62, 0x75, 0x79, 0x50, 0x72, 0x69, 0x63, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x73,
	0x65, 0x6c, 0x6c, 0x50, 0x72, 0x69, 0x63, 0x65, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x02, 0x52, 0x09,
	0x73, 0x65, 0x6c, 0x6c, 0x50, 0x72, 0x69, 0x63, 0x65, 0x42, 0x10, 0x5a, 0x0e, 0x66, 0x61, 0x63,
	0x61, 0x64, 0x65, 0x2f, 0x74, 0x73, 0x6f, 0x72, 0x64, 0x65, 0x72, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_tsorder_proto_rawDescOnce sync.Once
	file_tsorder_proto_rawDescData = file_tsorder_proto_rawDesc
)

func file_tsorder_proto_rawDescGZIP() []byte {
	file_tsorder_proto_rawDescOnce.Do(func() {
		file_tsorder_proto_rawDescData = protoimpl.X.CompressGZIP(file_tsorder_proto_rawDescData)
	})
	return file_tsorder_proto_rawDescData
}

var file_tsorder_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_tsorder_proto_goTypes = []interface{}{
	(*TOrder)(nil), // 0: tsorder.TOrder
}
var file_tsorder_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_tsorder_proto_init() }
func file_tsorder_proto_init() {
	if File_tsorder_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_tsorder_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TOrder); i {
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
			RawDescriptor: file_tsorder_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_tsorder_proto_goTypes,
		DependencyIndexes: file_tsorder_proto_depIdxs,
		MessageInfos:      file_tsorder_proto_msgTypes,
	}.Build()
	File_tsorder_proto = out.File
	file_tsorder_proto_rawDesc = nil
	file_tsorder_proto_goTypes = nil
	file_tsorder_proto_depIdxs = nil
}