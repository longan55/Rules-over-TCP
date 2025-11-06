package rot

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
)

type DataTyper interface {
	Value(src any) any
	ExplainedValue(src any) any
}

// BIN码 可以解释为整数、浮点数
// BCD码 可以解释为字符串、整数（max: 18446744073709551615）、浮点数
// ASCII 只能解释为字符串
// CP56TIME2A 可以解释为字符串

// Codec 编解码器接口
type Codec interface {
	Configure() // 初始化编解码器
	Encode(data any, byteLength int) ([]byte, error)
	Decode(data []byte) (any, error)
}

// CodecBIN BIN编解码器
type CodecBIN struct {
	order binary.ByteOrder
}

func NewCodecBIN(order binary.ByteOrder) *CodecBIN {
	return &CodecBIN{order: order}
}

func (c *CodecBIN) Configure() {
	// 初始化操作（如果需要）
}

func (c *CodecBIN) Encode(data any, byteLength int) ([]byte, error) {
	// 根据数据类型处理编码
	v, ok := data.(int)
	if !ok {
		return nil, fmt.Errorf("unsupported data type for BIN encoding: %T", data)
	}
	return Int2Bin(v, byte(byteLength), c.order), nil
}

func (c *CodecBIN) Decode(data []byte) (any, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data for decoding")
	}
	src := Bin2Int(data, c.order)
	return src, nil
}

// CodecBCD BCD编解码器
type CodecBCD struct {
	order binary.ByteOrder
}

func NewCodecBCD(order binary.ByteOrder) *CodecBCD {
	return &CodecBCD{order: order}
}

func (c *CodecBCD) Configure() {
	// 初始化操作（如果需要）
}

func (c *CodecBCD) Encode(data any, byteLength int) ([]byte, error) {
	// 根据数据类型处理编码
	v, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("unsupported data type for BCD encoding: %T", data)
	}
	bcdBytes, err := hex.DecodeString(v)
	if err != nil {
		return nil, err
	}
	return bcdBytes, nil
}

func (c *CodecBCD) Decode(data []byte) (any, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data for decoding")
	}
	// 根据字节序处理数据
	if c.order == binary.LittleEndian && len(data) > 1 {
		// 小端序需要反转字节顺序
		reversed := make([]byte, len(data))
		for i := range len(reversed) {
			reversed[i] = data[len(data)-1-i]
		}
		return hex.EncodeToString(reversed), nil
	} else {
		// 大端序（默认）直接使用原数据
		return hex.EncodeToString(data), nil
	}
}

// CodecASCII ASCII编解码器
type CodecASCII struct {
	// ASCII编解码器配置
}

func NewCodecASCII() *CodecASCII {
	return &CodecASCII{}
}

func (c *CodecASCII) Configure() {
	// 初始化操作（如果需要）
}

func (c *CodecASCII) Encode(data any, byteLength int) ([]byte, error) {
	// 根据数据类型处理编码
	v, ok := data.(string)
	if !ok {
		return nil, fmt.Errorf("unsupported data type for ASCII encoding: %T", data)
	}
	// 确保字符串长度不超过指定长度
	if len(v) > byteLength {
		v = v[:byteLength]
	}
	return []byte(v), nil
}

func (c *CodecASCII) Decode(data []byte) (any, error) {
	return string(data), nil
}

// BINFloat BIN浮点数解释器
type BINFloat1 struct {
	decimal int
}

var _ DataTyper = (*BINFloat1)(nil)

func (t *BINFloat1) Value(data any) any {
	return data
}

func (t *BINFloat1) ExplainedValue(data any) any {
	srcInt := data.(int)
	decimal := t.decimal
	// 计算浮点数
	result := float64(srcInt) / math.Pow10(decimal)

	// 应用倍数和偏移量

	return result
}

// BIN   - INT     explain(int -> int)
// BIN   - FLOAT   explain(int -> float)
// BCD   - STRING  explain(string -> string)
// BCD   - FLOAT   explain(string -> float)
// BCD   - INT     explain(string -> int)
// ASCII - STRING  explain(string -> string)

type NewDataTyper interface {
	Explain(data any) any
	UnExplain(data any) any
}

func WithBinInteger(moflag bool, multiple int, offset int) CodecOption {
	return &binInteger{moflag: moflag, multiple: multiple, offset: offset}
}

type binInteger struct {
	moflag   bool
	multiple int
	offset   int
}

var (
	_ CodecOption  = (*binInteger)(nil)
	_ NewDataTyper = (*binInteger)(nil)
)

func (t *binInteger) Explain(data any) any {
	srcInt := data.(int)
	// 应用倍数和偏移量
	result := srcInt
	if t.moflag {
		result = srcInt*t.multiple + t.offset
	} else {
		result = (srcInt + t.offset) * t.multiple
	}
	return result
}
func (t *binInteger) UnExplain(data any) any {
	srcFloat := data.(int)
	result := 0
	if t.moflag {
		result = (srcFloat - t.offset) / t.multiple
	} else {
		result = (srcFloat / t.multiple) - t.offset
	}
	return result
}
func (t *binInteger) Apply(config *FieldCodecConfig) {
	config.ndt = t
}
func WithBinFloat(moflag bool, multiple float64, offset float64) CodecOption {
	return &binFloat{moflag: moflag, multiple: multiple, offset: offset}
}

type binFloat struct {
	moflag   bool
	multiple float64
	offset   float64
}

var (
	_ CodecOption  = (*binFloat)(nil)
	_ NewDataTyper = (*binFloat)(nil)
)

func (t *binFloat) Explain(data any) any {
	srcInt := data.(int)
	result := 0.0
	if t.moflag {
		result = float64(srcInt)*t.multiple + t.offset
	} else {
		result = (float64(srcInt) + t.offset) * t.multiple
	}
	return result
}

func (t *binFloat) UnExplain(data any) any {
	srcFloat := data.(float64)
	result := 0
	if t.moflag {
		result = int((srcFloat - t.offset) / t.multiple)
	} else {
		result = int((srcFloat / t.multiple) - t.offset)
	}
	return result
}

func (t *binFloat) Apply(config *FieldCodecConfig) {
	config.ndt = t
}

type bcdInteger struct {
	moflag   bool
	multiple int
	offset   int
}

var (
	_ CodecOption  = (*bcdInteger)(nil)
	_ NewDataTyper = (*bcdInteger)(nil)
)

func (t *bcdInteger) Explain(data any) any {
	str := data.(string)
	// 应用倍数和偏移量
	srcInt, err := strconv.Atoi(str)
	if err != nil {
		return nil
	}
	result := srcInt
	if t.moflag {
		result = srcInt*t.multiple + t.offset
	} else {
		result = (srcInt + t.offset) * t.multiple
	}
	return result
}
func (t *bcdInteger) UnExplain(data any) any {
	srcFloat := data.(int)
	result := 0
	if t.moflag {
		result = (srcFloat - t.offset) / t.multiple
	} else {
		result = (srcFloat / t.multiple) - t.offset
	}
	return strconv.Itoa(result)
}
func (t *bcdInteger) Apply(config *FieldCodecConfig) {
	config.ndt = t
}

type bcdFloat struct {
	moflag   bool
	multiple float64
	offset   float64
}

var (
	_ CodecOption  = (*bcdFloat)(nil)
	_ NewDataTyper = (*bcdFloat)(nil)
)

func (t *bcdFloat) Explain(data any) any {
	str := data.(string)
	// 应用倍数和偏移量
	srcFloat, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return nil
	}
	result := srcFloat
	if t.moflag {
		result = srcFloat*t.multiple + t.offset
	} else {
		result = (srcFloat + t.offset) * t.multiple
	}
	return result
}
func (t *bcdFloat) UnExplain(data any) any {
	srcFloat := data.(float64)
	result := 0.0
	if t.moflag {
		result = (srcFloat - t.offset) / t.multiple
	} else {
		result = (srcFloat / t.multiple) - t.offset
	}
	return strconv.FormatFloat(result, 'f', 6, 64)
}
func (t *bcdFloat) Apply(config *FieldCodecConfig) {
	config.ndt = t
}
func WithBcdString() CodecOption {
	return &bcdString{}
}

type bcdString struct {
}

func (t *bcdString) Apply(config *FieldCodecConfig) {
	config.ndt = t
}
func (t *bcdString) Explain(data any) any {
	return data
}
func (t *bcdString) UnExplain(data any) any {
	return data
}
