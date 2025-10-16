package protocol

import (
	"encoding/binary"
)

type EncodeBuilder struct {
	order binary.ByteOrder
}

func NewEncoder(order binary.ByteOrder) *EncodeBuilder {
	return &EncodeBuilder{order: order}
}

func (builder *EncodeBuilder) BIN() *EncodeBIN {
	return &EncodeBIN{order: builder.order}
}

type EncodeBIN struct {
	order binary.ByteOrder
}

func (bin EncodeBIN) Integer() DataTypeInteger {
	return DataTypeInteger{}
}

type DataTypeInteger struct{}

func (i DataTypeInteger) Integer() int {
	return 0
}

type BCD struct {
	order binary.ByteOrder
}

type ASCII struct {
	order binary.ByteOrder
}

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
	switch c.Encode {
	case EncodingBIN:
		value, err := decode(c.Order, c.Datatype)
		if err != nil {
			return nil
		}
		return value
	case EncodingBCD:
		switch c.Datatype {
		case DataTypeInt:
			uint64, err := BCD2Int(nil)
			if err != nil {
				return nil
			}
			return int64(uint64)
		default:
			return nil
		}
	case EncodingASCII:
		return c.Enum
	default:
		return nil
	}
}

func decode(order binary.ByteOrder, datatype DataType) (any, error) {
	switch datatype {
	case DataTypeInt:
		uint64, err := BIN2Uint64(nil, order)
		if err != nil {
			return nil, err
		}
		return int64(uint64), nil
	case DataTypeFloat:
		float64, err := Bin2Float64(order, nil, 6)
		if err != nil {
			return nil, err
		}
		return float64, nil
	case DataTypeString:
		panic("EncodingType BIN is not support DataTypeString")
	default:
		return nil, nil
	}
}
