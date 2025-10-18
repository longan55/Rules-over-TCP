package protocol

import (
	"encoding/binary"
	"encoding/hex"
	"strconv"
)

type Decoder interface {
	// 如果希望 value 是传地址/指针，入参类型应为指针类型，使用 any 表示任意类型的指针
	encode(data []byte, value any) error
}

type DecodeBuilder struct {
	order binary.ByteOrder
}

func NewDecoder(order binary.ByteOrder) *DecodeBuilder {
	return &DecodeBuilder{order: order}
}

func (builder *DecodeBuilder) BIN() *DecodeBIN {
	return &DecodeBIN{order: builder.order}
}

func (builder *DecodeBuilder) BCD() *BCD {
	return &BCD{order: builder.order}
}

func (builder *DecodeBuilder) ASCII() *ASCII {
	return &ASCII{order: builder.order}
}

func (builder *DecodeBuilder) CP56TIME2A() *CP56TIME2A {
	return &CP56TIME2A{order: builder.order}
}

type DecodeBIN struct {
	byteLength int
	order      binary.ByteOrder
}

func (bin *DecodeBIN) SetByteLength(byteLength int) *DecodeBIN {
	bin.byteLength = byteLength
	return bin
}

func (bin *DecodeBIN) encode(data []byte, value any) error {
	*value.(*int) = Bin2Int(data, bin.order)
	return nil
}

func (bin *DecodeBIN) Integer() *BINInteger {
	return &BINInteger{
		encoder: bin,
		order:   bin.order,
		mul:     1,
	}
}

func (bin *DecodeBIN) Float1() *BINFloat {
	return &BINFloat{
		encoder: bin,
		order:   bin.order,
		mul:     1,
	}
}

func (bin *DecodeBIN) String1() *BINString {
	return &BINString{
		encoder: bin,
		order:   bin.order,
	}
}

type BINInteger struct {
	encoder  Decoder
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
func (i *BINInteger) Multiple(mul float64) *BINInteger {
	i.mul = mul
	if !i.oflag {
		i.mflag = true
	}
	return i
}

// Offset: 设置偏移量，注意和Multiple的先后顺序，将影响计算顺序
func (i *BINInteger) Offset(offset float64) *BINInteger {
	i.offset = offset
	if i.mflag {
		i.oflag = true
	}
	return i
}

func (i *BINInteger) SetEnum(enum map[int]any) *BINInteger {
	i.enum = enum
	return i
}

func (i *BINInteger) SetBitMap(bitmap map[int]any) *BINInteger {
	i.bitmap = bitmap
	return i
}

func (i *BINInteger) SourceValue(data []byte) int {
	i.encoder.encode(data, &i.srcValue)
	if i.mflag {
		return i.srcValue*int(i.mul) + int(i.offset)
	} else if i.oflag {
		return (i.srcValue + int(i.offset)) * int(i.mul)
	} else {
		return i.srcValue
	}
}

func (i *BINInteger) ExplainedValue(data []byte) any {
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

type BINFloat struct {
	encoder  Decoder
	order    binary.ByteOrder
	mflag    bool
	oflag    bool
	mul      float64
	offset   float64
	srcValue float64
}

func (f *BINFloat) Multiple(mul float64) *BINFloat {
	f.mul = mul
	if !f.oflag {
		f.mflag = true
	}
	return f
}

func (f *BINFloat) Offset(offset float64) *BINFloat {
	f.offset = offset
	if f.mflag {
		f.oflag = true
	}
	return f
}

func (f *BINFloat) SourceValue(data []byte) float64 {
	temp := 0
	f.encoder.encode(data, &temp)
	if f.mflag {
		f.srcValue = float64(temp)*f.mul + f.offset
		return f.srcValue
	} else if f.oflag {
		f.srcValue = (float64(temp) + f.offset) * f.mul
		return f.srcValue
	} else {
		f.srcValue = float64(temp)
		return f.srcValue
	}
}

func (f *BINFloat) ExplainedValue(data []byte) float64 {
	return f.SourceValue(data)
}

type BINString struct {
	encoder  Decoder
	order    binary.ByteOrder
	srcValue string
}

func (s *BINString) SourceValue(data []byte) string {
	temp := 0
	s.encoder.encode(data, &temp)
	return s.srcValue
}

type BCD struct {
	order binary.ByteOrder
}

func (bcd *BCD) encode(data []byte, value any) error {
	// 根据字节序处理数据
	if bcd.order == binary.LittleEndian && len(data) > 1 {
		// 小端序需要反转字节顺序
		reversed := make([]byte, len(data))
		for i := range len(reversed) {
			reversed[i] = data[len(data)-1-i]
		}
		*value.(*string) = hex.EncodeToString(reversed)
	} else {
		// 大端序（默认）直接使用原数据
		*value.(*string) = hex.EncodeToString(data)
	}
	return nil
}

func (bcd *BCD) Integer() *BCDInteger {
	return &BCDInteger{
		encoder: bcd,
		// order:   bcd.order,
	}
}

type BCDInteger struct {
	encoder Decoder
	// order   binary.ByteOrder
}

func (bcdi *BCDInteger) SourceValue(data []byte) int {
	temp := ""
	bcdi.encoder.encode(data, &temp)
	source, err := strconv.Atoi(temp)
	if err != nil {
		return 0
	}
	return source
}

// type BCDFloat struct {
// 	encoder Encoder
// 	order   binary.ByteOrder
// }

// func (bcdf *BCDFloat) SourceValue(data []byte) float64 {
// 	temp := ""
// 	bcdf.encoder.encode(data, &temp)
// 	source, err := strconv.ParseFloat(temp, 64)
// 	if err != nil {
// 		return 0
// 	}
// 	return source
// }

func (bcd *BCD) String() *BCDString {
	return &BCDString{
		encoder: bcd,
	}
}

type BCDString struct {
	encoder Decoder
}

func (bcds *BCDString) SourceValue(data []byte) (str string) {
	bcds.encoder.encode(data, &str)
	return
}

type ASCII struct {
	order binary.ByteOrder
}

func (ascii *ASCII) encode(data []byte, value any) error {
	*value.(*string) = string(data)
	return nil
}

func (ascii *ASCII) String() *ASCIIString {
	return &ASCIIString{
		encoder: ascii,
	}
}

type ASCIIString struct {
	encoder Decoder
}

func (ascii *ASCIIString) SourceValue(data []byte) (str string) {
	ascii.encoder.encode(data, &str)
	return
}

type CP56TIME2A struct {
	order binary.ByteOrder
}

func (cp56 *CP56TIME2A) encode(data []byte, value any) error {
	*value.(*string) = ParseCP56time2a(data)
	return nil
}

func (cp56 *CP56TIME2A) String() *CP56TIME2AString {
	return &CP56TIME2AString{
		encoder: cp56,
	}
}

type CP56TIME2AString struct {
	encoder Decoder
}

func (cp56 *CP56TIME2AString) SourceValue(data []byte) (str string) {
	cp56.encoder.encode(data, &str)
	return
}
