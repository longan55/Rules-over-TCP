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
	switch c.Encode {
	case EncodingBIN:
		switch c.Datatype {
		case DataTypeInt:
			uint64, err := BIN2Uint64(nil, c.Order)
			if err != nil {
				return nil
			}
			return int64(uint64)
		case DataTypeFloat:
			float64, err := Bin2Float64(c.Order, nil, 6)
			if err != nil {
				return nil
			}
			return float64
		case DataTypeString:
			panic("EncodingType BIN is not support DataTypeString")
		default:
			return nil
		}
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
