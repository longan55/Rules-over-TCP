package rot

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
)

// BIN码 可以解释为整数、浮点数
// BCD码 可以解释为字符串、整数（max: 18446744073709551615）、浮点数
// ASCII 只能解释为字符串
// CP56TIME2A 可以解释为字符串

// CodecOption 配置选项接口
type CodecOption interface {
	Apply(config *FieldCodecConfig)
}

// CodecMode 编解码模式
type CodecMode int

const (
	ModeDecode CodecMode = iota // 解码模式（反序列化）
	ModeEncode                  // 编码模式（序列化）
)

// FieldCodecConfig 字段编解码配置
type FieldCodecConfig struct {
	name          string
	length        int
	mode          CodecMode
	codec         Codec
	dataTyper     DataTyper
	ndt           NewDataTyper
	explainConfig *ExplainConfig
}

// ExplainConfig 数据解释配置
type ExplainConfig struct {
	moflag   bool
	multiple float64 // 倍数
	offset   float64 // 偏移量

	other string      // 其他解释
	enum  map[int]any // 枚举映射

	bitmap map[int]any // 位图映射
}

// NewFieldCodecConfig 创建新的字段编解码配置
func NewFieldCodecConfig(name string, options ...CodecOption) *FieldCodecConfig {
	config := &FieldCodecConfig{
		name:          name,
		mode:          ModeDecode,
		explainConfig: &ExplainConfig{},
	}

	// 应用所有选项
	for _, option := range options {
		option.Apply(config)
	}

	// 如果没有指定编解码器，默认使用BIN编解码器（大端序）
	if config.codec == nil {
		config.codec = NewCodecBIN(binary.BigEndian)
	}

	// 根据编解码器类型设置默认的数据类型解释器
	// if config.dataTyper == nil {
	// 	switch config.codec.(type) {
	// 	case *CodecBIN:
	// 		config.dataTyper = &BINInteger{}
	// 	case *CodecBCD:
	// 		config.dataTyper = &BCDString{}
	// 	case *CodecASCII:
	// 		config.dataTyper = &ASCIIString{}
	// 	}
	// }

	return config
}

// WithLength 设置字段长度选项
func WithLength(length int) CodecOption {
	return &lengthOption{length}
}

type lengthOption struct {
	length int
}

func (o *lengthOption) Apply(config *FieldCodecConfig) {
	config.length = o.length
}

// WithMode 设置编解码模式选项
func WithMode(mode CodecMode) CodecOption {
	return &modeOption{mode}
}

type modeOption struct {
	mode CodecMode
}

func (o *modeOption) Apply(config *FieldCodecConfig) {
	config.mode = o.mode
}

// WithCodec 设置编解码器选项
func WithCodec(codec Codec) CodecOption {
	return &codecOption{codec}
}

type codecOption struct {
	codec Codec
}

func (o *codecOption) Apply(config *FieldCodecConfig) {
	o.codec.Configure() // 初始化编解码器
	config.codec = o.codec
}

// WithDataTyper 设置数据类型解释器选项
func WithDataTyper(dataTyper DataTyper) CodecOption {
	return &dataTyperOption{dataTyper}
}

type dataTyperOption struct {
	dataTyper DataTyper
}

func (o *dataTyperOption) Apply(config *FieldCodecConfig) {
	config.dataTyper = o.dataTyper
}

// WithMultiple 设置倍数选项
func WithMultiple(multiple float64) CodecOption {
	return &multipleOption{multiple}
}

type multipleOption struct {
	multiple float64
}

func (o *multipleOption) Apply(config *FieldCodecConfig) {
	if config.explainConfig == nil {
		config.explainConfig = &ExplainConfig{}
	}
	config.explainConfig.multiple = o.multiple
	if config.explainConfig.offset != 0 {
		config.explainConfig.moflag = true
	}
}

// WithOffset 设置偏移量选项
func WithOffset(offset float64) CodecOption {
	return &offsetOption{offset}
}

type offsetOption struct {
	offset float64
}

func (o *offsetOption) Apply(config *FieldCodecConfig) {
	if config.explainConfig == nil {
		config.explainConfig = &ExplainConfig{}
	}
	config.explainConfig.offset = o.offset
}

// WithEnum 设置枚举映射选项
func WithEnum(other string, enum map[int]any) CodecOption {
	return &enumOption{other, enum}
}

type enumOption struct {
	other string
	enum  map[int]any
}

func (o *enumOption) Apply(config *FieldCodecConfig) {
	if config.explainConfig == nil {
		config.explainConfig = &ExplainConfig{}
	}
	config.explainConfig.enum = o.enum
}

func WithBitmap(bitmap map[int]any) CodecOption {
	return &bitmapOption{bitmap}
}

type bitmapOption struct {
	bitmap map[int]any
}

func (o *bitmapOption) Apply(config *FieldCodecConfig) {
	if config.explainConfig == nil {
		config.explainConfig = &ExplainConfig{}
	}
	config.explainConfig.bitmap = o.bitmap
}

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

// // Decode 解码方法
func (config *FieldCodecConfig) Decode(data []byte) (*ParsedData, error) {
	if config.mode != ModeDecode {
		return nil, errors.New("config is not in decode mode")
	}
	// 使用编解码器解码
	rawValue, err := config.codec.Decode(data)
	if err != nil {
		return nil, err
	}
	explainedValue := config.ndt.Explain(rawValue)

	parsed := &ParsedData{
		Bytes:     data,
		Origin:    explainedValue,
		Explained: explainedValue,
	}

	if config.explainConfig != nil && config.explainConfig.enum != nil {
		i, ok := explainedValue.(int)
		if !ok {
			return nil, fmt.Errorf("explained value is not int: %v", explainedValue)
		}
		if enumValue, ok := config.explainConfig.enum[i]; ok {
			parsed.Explained = enumValue
			return parsed, nil
		} else {
			parsed.Explained = config.explainConfig.other
			return parsed, fmt.Errorf("explained value %v not found in enum", i)
		}
	}

	return parsed, nil
}

// Encode 编码方法
func (config *FieldCodecConfig) Encode(data any) ([]byte, error) {
	if config.mode != ModeEncode {
		return nil, errors.New("config is not in encode mode")
	}
	if config.ndt != nil {
		data = config.ndt.UnExplain(data)
	}
	// 使用编解码器编码
	return config.codec.Encode(data, config.length)
}
