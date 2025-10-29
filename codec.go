package rot

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
)

type DecodeType interface {
	decode(data []byte, value any) error
	Decode(data []byte) any
	GetByteLength() int
	// 如果希望 src 是传地址/指针，入参类型应为指针类型，使用 any 表示任意类型的指针
	ExplainedValue(src any) any
}

type DataTyper interface {
	Value(src any) any
	ExplainedValue(src any) any
}

type DecoderImpl struct {
	fh      *FucntionHandler
	decoder DecodeType
	order   binary.ByteOrder
}

func (impl *DecoderImpl) AddLength(length int) {
	impl.fh.length += length
}

func (impl *DecoderImpl) Decode(data []byte) (any, error) {
	sv := impl.decoder.Decode(data)
	return sv, nil
}

func (impl *DecoderImpl) BIN() *BIN {
	bin := BIN{decoder: impl, order: impl.order}
	impl.decoder = &bin
	return &bin
}

func (impl *DecoderImpl) BCD() *BCD {
	bcd := BCD{decoder: impl, order: impl.order}
	impl.decoder = &bcd
	return &bcd
}

func (impl *DecoderImpl) ASCII() *ASCII {
	ascii := ASCII{decoder: impl, order: impl.order}
	impl.decoder = &ascii
	return &ascii
}

func (impl *DecoderImpl) CP56TIME2A() *CP56TIME2A {
	cp56time2a := CP56TIME2A{decoder: impl, order: impl.order}
	impl.decoder = &cp56time2a
	return &cp56time2a
}

type BIN struct {
	decoder    *DecoderImpl
	dataTyper  DataTyper
	byteLength int
	order      binary.ByteOrder
}

var _ DecodeType = (*BIN)(nil)

func (bin *BIN) GetByteLength() int {
	return bin.byteLength
}

func (bin *BIN) SetByteLength(byteLength int) *BIN {
	bin.byteLength = byteLength
	bin.decoder.AddLength(byteLength)
	return bin
}

func (bin *BIN) decode(data []byte, value any) error {
	*value.(*int) = Bin2Int(data, bin.order)
	return nil
}

func (bin *BIN) Decode(data []byte) any {
	var src int
	err := bin.decode(data, &src)
	if err != nil {
		return nil
	}
	sv := bin.dataTyper.Value(src)
	return sv
}

func (bin *BIN) ExplainedValue(src any) any {
	return bin.dataTyper.ExplainedValue(src)
}

func (bin *BIN) Integer() *BINInteger {
	temp := &BINInteger{
		encoder: bin,
		order:   bin.order,
		mul:     1,
	}
	bin.dataTyper = temp
	return temp
}

func (bin *BIN) Float1() *BINFloat {
	tmp := &BINFloat{
		encoder: bin,
		order:   bin.order,
		mul:     1,
	}
	bin.dataTyper = tmp
	return tmp
}

func (bin *BIN) String1() *BINString {
	return &BINString{
		encoder: bin,
		order:   bin.order,
	}
}

type BINInteger struct {
	encoder DecodeType
	order   binary.ByteOrder
	mflag   bool
	oflag   bool
	mul     float64
	offset  float64
	enum    map[int]any
	bitmap  map[int]any
}

var _ DataTyper = (*BINInteger)(nil)

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

func (i *BINInteger) Value(src any) any {
	srcInt := src.(int)
	if i.mflag {
		return srcInt*int(i.mul) + int(i.offset)
	} else if i.oflag {
		return (srcInt + int(i.offset)) * int(i.mul)
	} else {
		return srcInt
	}
}

func (i *BINInteger) ExplainedValue(src any) any {
	srcInt := src.(int)
	if i.enum != nil {
		return i.enum[srcInt]
	}
	if i.bitmap != nil {
		return i.bitmap[srcInt]
	}
	return srcInt
}

type BINFloat struct {
	encoder DecodeType
	order   binary.ByteOrder
	mflag   bool
	oflag   bool
	mul     float64
	offset  float64
	// srcValue float64
}

var _ DataTyper = (*BINFloat)(nil)

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

func (f *BINFloat) Value(src any) any {
	srcFloat := src.(int)
	if f.mflag {
		return float64(srcFloat)*f.mul + f.offset
	} else if f.oflag {
		return (float64(srcFloat) + f.offset) * f.mul
	} else {
		return float64(srcFloat)
	}
}

func (f *BINFloat) ExplainedValue(src any) any {
	return src
}

func (s *BINString) ExplainedValue(src any) any {
	srcStr := src.(string)
	return srcStr
}

type BINString struct {
	encoder DecodeType
	order   binary.ByteOrder
}

var _ DataTyper = (*BINString)(nil)

func (s *BINString) Value(src any) any {
	srcStr := src.(int)
	return strconv.Itoa(srcStr)
}

// ////////
// /////////
// /////////
// ////////
// ////////
// ////////
// /////////
// /////////
// ////////
// ////////
// ////////
// /////////
// /////////
// ////////
// ////////
// ////////
// /////////
// /////////
// ////////
// ////////
// ////////
// /////////
// /////////
// ////////
// ////////
type BCD struct {
	decoder    *DecoderImpl
	order      binary.ByteOrder
	byteLength int
	dataTyper  DataTyper
}

func (bcd *BCD) decode(data []byte, value any) error {
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

func (bcd *BCD) Decode(data []byte) any {
	var value string
	err := bcd.decode(data, &value)
	if err != nil {
		return nil
	}
	return value
}

func (bcd *BCD) ExplainedValue(src any) any {
	//FIXME:runtime error: invalid memory address or nil pointer dereference
	return bcd.dataTyper.ExplainedValue(src)
}

func (bcd *BCD) SetByteLength(byteLength int) *BCD {
	bcd.byteLength = byteLength
	bcd.decoder.AddLength(byteLength)
	return bcd
}
func (bcd *BCD) GetByteLength() int {
	return bcd.byteLength
}
func (bcd *BCD) Integer() *BCDInteger {
	return &BCDInteger{
		encoder: bcd,
		// order:   bcd.order,
	}
}

type BCDInteger struct {
	encoder DecodeType
	// order   binary.ByteOrder
}

var _ DataTyper = (*BCDInteger)(nil)

func (bcdi *BCDInteger) Value(src any) any {
	srcInt := src.(string)
	source, err := strconv.Atoi(srcInt)
	if err != nil {
		return 0
	}
	return source
}

func (bcdi *BCDInteger) ExplainedValue(src any) any {
	return bcdi.Value(src)
}

func (bcdi *BCDInteger) SourceValue(data []byte) int {
	temp := ""
	bcdi.encoder.decode(data, &temp)
	source, err := strconv.Atoi(temp)
	if err != nil {
		return 0
	}
	return source
}

func (bcd *BCD) Float() *BCDFloat {
	return &BCDFloat{
		encoder: bcd,
	}
}

type BCDFloat struct {
	encoder DecodeType
	decimal int
}

var _ DataTyper = (*BCDFloat)(nil)

func (bcdf *BCDFloat) ExplainedValue(src any) any {
	return bcdf.Value(src)
}

func (bcdf *BCDFloat) DecimalPlace(decimal int) *BCDFloat {
	bcdf.decimal = decimal
	return bcdf
}

func (bcdf *BCDFloat) Value(src any) any {
	srcFloat := src.(int)
	return float64(srcFloat) / math.Pow10(bcdf.decimal)
}

func (bcdf *BCDFloat) SourceValue(data []byte) float64 {
	temp := ""
	bcdf.encoder.decode(data, &temp)

	// 先将字符串转换为浮点数
	source, err := strconv.ParseFloat(temp, 64)
	if err != nil {
		return 0
	}

	// 计算除以10的幂次后的值
	value := source / math.Pow10(bcdf.decimal)

	// 使用FormatFloat和ParseFloat来精确控制小数位数
	// 这可以确保结果按照指定的小数位数进行四舍五入，减少浮点数精度误差
	format := "%." + strconv.Itoa(bcdf.decimal) + "f"
	formatted := fmt.Sprintf(format, value)

	// 将格式化后的字符串转回浮点数
	result, err := strconv.ParseFloat(formatted, 64)
	if err != nil {
		return value // 如果格式化失败，返回原始计算值
	}

	return result
}

func (bcd *BCD) String() *BCDString {
	return &BCDString{
		encoder: bcd,
	}
}

type BCDString struct {
	encoder DecodeType
}

func (bcds *BCDString) SourceValue(data []byte) (str string) {
	bcds.encoder.decode(data, &str)
	return
}

type ASCII struct {
	decoder    *DecoderImpl
	order      binary.ByteOrder
	byteLength int
	dataTyper  DataTyper
}

func (ascii *ASCII) decode(data []byte, value any) error {
	*value.(*string) = string(data)
	return nil
}

func (ascii *ASCII) Decode(data []byte) any {
	var value string
	ascii.decode(data, &value)
	return value
}
func (ascii *ASCII) SetByteLength(byteLength int) *ASCII {
	ascii.byteLength = byteLength
	ascii.decoder.AddLength(byteLength)
	return ascii
}
func (ascii *ASCII) GetByteLength() int {
	return ascii.byteLength
}
func (ascii *ASCII) String() *ASCIIString {
	return &ASCIIString{
		encoder: ascii,
	}
}

func (ascii *ASCII) ExplainedValue(src any) any {
	return ascii.dataTyper.ExplainedValue(src)
}

type ASCIIString struct {
	encoder DecodeType
}

func (ascii *ASCIIString) SourceValue(data []byte) (str string) {
	ascii.encoder.decode(data, &str)
	return
}

type CP56TIME2A struct {
	decoder    *DecoderImpl
	order      binary.ByteOrder
	byteLength int
	dataTyper  DataTyper
}

func (cp56 *CP56TIME2A) decode(data []byte, value any) error {
	*value.(*string) = ParseCP56time2a(data)
	return nil
}

func (cp56 *CP56TIME2A) Decode(data []byte) any {
	var value string
	cp56.decode(data, &value)
	return value
}

func (cp56 *CP56TIME2A) ExplainedValue(src any) any {
	return cp56.dataTyper.ExplainedValue(src)
}

func (cp56 *CP56TIME2A) SetByteLength(byteLength int) *CP56TIME2A {
	cp56.byteLength = byteLength
	cp56.decoder.AddLength(byteLength)
	return cp56
}
func (cp56 *CP56TIME2A) GetByteLength() int {
	return cp56.byteLength
}
func (cp56 *CP56TIME2A) String() *CP56TIME2AString {
	return &CP56TIME2AString{
		encoder: cp56,
	}
}

type CP56TIME2AString struct {
	encoder DecodeType
}

func (cp56 *CP56TIME2AString) SourceValue(data []byte) (str string) {
	cp56.encoder.decode(data, &str)
	return
}
