package protocol

import (
	"encoding/binary"
)

type Encoder interface {
	// 如果希望 value 是传地址/指针，入参类型应为指针类型，使用 *any 表示任意类型的指针
	encode(data []byte, value any) error
}

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
	byteLength int
	order      binary.ByteOrder
}

func (bin *EncodeBIN) SetByteLength(byteLength int) *EncodeBIN {
	bin.byteLength = byteLength
	return bin
}

func (bin *EncodeBIN) encode(data []byte, value any) error {
	*value.(*int) = Bin2Int(data, bin.order)
	return nil
}

func (bin *EncodeBIN) Integer() *DataTypeInteger {
	return &DataTypeInteger{
		encoder: bin,
		order:   bin.order,
		mul:     1,
	}
}

func (bin *EncodeBIN) Float1() *DataTypeFloat1 {
	return &DataTypeFloat1{
		encoder: bin,
		order:   bin.order,
		mul:     1,
	}
}

type DataTypeInteger struct {
	encoder  Encoder
	order    binary.ByteOrder
	mflag    bool
	oflag    bool
	mul      float64
	offset   float64
	srcValue int
	enum     map[int]any
	bitmap   map[int]any
}

// TODO:Multiple和Offset，考虑是否需要设置幂等性
func (i *DataTypeInteger) Multiple(mul float64) *DataTypeInteger {
	i.mul = mul
	if !i.oflag {
		i.mflag = true
	}
	return i
}

// Offset: 设置偏移量，注意和Multiple的先后顺序，将影响计算顺序
func (i *DataTypeInteger) Offset(offset float64) *DataTypeInteger {
	i.offset = offset
	if i.mflag {
		i.oflag = true
	}
	return i
}

func (i *DataTypeInteger) SetEnum(enum map[int]any) *DataTypeInteger {
	i.enum = enum
	return i
}

func (i *DataTypeInteger) SetBitMap(bitmap map[int]any) *DataTypeInteger {
	i.bitmap = bitmap
	return i
}

func (i *DataTypeInteger) SourceValue(data []byte) int {
	i.encoder.encode(data, &i.srcValue)
	if i.mflag {
		return i.srcValue*int(i.mul) + int(i.offset)
	} else if i.oflag {
		return (i.srcValue + int(i.offset)) * int(i.mul)
	} else {
		return i.srcValue
	}
}

func (i *DataTypeInteger) ExplainedValue(data []byte) any {
	if i.enum == nil && i.bitmap == nil {
		return i.SourceValue(data)
	}
	if i.enum != nil {
		return i.enum[i.SourceValue(data)]
	} else {
		//i.bitmap != nil
		result := make([]string, 0, len(i.bitmap))
		source := i.SourceValue(data)
		for k, v := range i.bitmap {
			if (source & (1 << k)) != 0 {
				result = append(result, v.(string))
			}
		}
		return result
	}
}

type DataTypeFloat1 struct {
	encoder  Encoder
	order    binary.ByteOrder
	mflag    bool
	oflag    bool
	mul      float64
	offset   float64
	srcValue float64
}

func (f *DataTypeFloat1) Multiple(mul float64) *DataTypeFloat1 {
	f.mul = mul
	if !f.oflag {
		f.mflag = true
	}
	return f
}

func (f *DataTypeFloat1) Offset(offset float64) *DataTypeFloat1 {
	f.offset = offset
	if f.mflag {
		f.oflag = true
	}
	return f
}

func (f *DataTypeFloat1) SourceValue(data []byte) float64 {
	f.srcValue = float64(Bin2Int(data, f.order))
	if f.mflag {
		return f.srcValue*f.mul + f.offset
	} else if f.oflag {
		return (f.srcValue + f.offset) * f.mul
	} else {
		return f.srcValue
	}
}

func (f *DataTypeFloat1) ExplainedValue(data []byte) float64 {
	return f.SourceValue(data)
}

type DataTypeString1 struct{}

type BCD struct {
	order binary.ByteOrder
}

type ASCII struct {
	order binary.ByteOrder
}
