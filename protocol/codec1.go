package protocol

import "encoding/binary"

type CodeItem struct {
	Order    binary.ByteOrder
	Encode   EncodingType
	Datatype DataType
	Multiple float64
	Offset   float64
	Enum     map[any]any
	bitmap   map[any]any
}

func NewCodecItem(order binary.ByteOrder, encode EncodingType, datatype DataType, multiple, offset float64) *CodeItem {
	return &CodeItem{
		Order:    order,
		Encode:   encode,
		Datatype: datatype,
		Multiple: multiple,
		Offset:   offset,
	}
}

func (c *CodeItem) SetEnum(enum map[any]any) {
	c.Enum = enum
}

func (c *CodeItem) Extract() any {
	return c.Enum
}
