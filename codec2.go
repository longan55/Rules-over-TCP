package rot

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
)

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
	explainConfig *ExplainConfig
}

// ExplainConfig 数据解释配置
type ExplainConfig struct {
	decimal  int         // 小数位数
	multiple float64     // 倍数
	offset   float64     // 偏移量
	enum     map[int]any // 枚举映射
	bitmap   map[int]any // 位图映射
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
	if config.dataTyper == nil {
		switch config.codec.(type) {
		case *CodecBIN:
			config.dataTyper = &BINInteger{}
		case *CodecBCD:
			config.dataTyper = &BCDString{}
		case *CodecASCII:
			config.dataTyper = &ASCIIString{}
		}
	}

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

// WithDecimal 设置小数位数选项
func WithDecimal(decimal int) CodecOption {
	return &decimalOption{decimal}
}

type decimalOption struct {
	decimal int
}

func (o *decimalOption) Apply(config *FieldCodecConfig) {
	config.explainConfig.decimal = o.decimal
	// 如果设置了小数位数，自动使用浮点数解释器
	switch config.codec.(type) {
	case *CodecBIN:
		config.dataTyper = &BINFloat1{
			decimal: o.decimal,
		}
	case *CodecBCD:
		config.dataTyper = &BCDFloat{
			decimal: o.decimal,
		}
	}
}

// WithMultiple 设置倍数选项
func WithMultiple(multiple float64) CodecOption {
	return &multipleOption{multiple}
}

type multipleOption struct {
	multiple float64
}

func (o *multipleOption) Apply(config *FieldCodecConfig) {
	config.explainConfig.multiple = o.multiple
}

// WithOffset 设置偏移量选项
func WithOffset(offset float64) CodecOption {
	return &offsetOption{offset}
}

type offsetOption struct {
	offset float64
}

func (o *offsetOption) Apply(config *FieldCodecConfig) {
	config.explainConfig.offset = o.offset
}

// WithEnum 设置枚举映射选项
func WithEnum(enum map[int]any) CodecOption {
	return &enumOption{enum}
}

type enumOption struct {
	enum map[int]any
}

func (o *enumOption) Apply(config *FieldCodecConfig) {
	config.explainConfig.enum = o.enum
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
	switch v := data.(type) {
	case int:
		return Int2Bin(v, byte(byteLength), c.order), nil
	case float64:
		// 浮点数需要特殊处理，这里简化处理
		return Int2Bin(int(v), byte(byteLength), c.order), nil
	default:
		return nil, fmt.Errorf("unsupported data type for BIN encoding: %T", data)
	}
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
	switch v := data.(type) {
	case string:
		bcdBytes, err := hex.DecodeString(v)
		if err != nil {
			return nil, err
		}
		return bcdBytes, nil
	case float64:
		// 浮点数需要转换为字符串后再编码
		format := "%0" + strconv.Itoa(byteLength*2) + ".0f"
		strValue := fmt.Sprintf(format, v*math.Pow10(c.getDecimalFromContext()))
		bcdBytes, err := hex.DecodeString(strValue)
		if err != nil {
			return nil, err
		}
		return bcdBytes, nil
	default:
		return nil, fmt.Errorf("unsupported data type for BCD encoding: %T", data)
	}
}

func (c *CodecBCD) getDecimalFromContext() int {
	// 这里简化处理，实际应该从上下文中获取
	return 0
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
	switch v := data.(type) {
	case string:
		// 确保字符串长度不超过指定长度
		if len(v) > byteLength {
			v = v[:byteLength]
		}
		return []byte(v), nil
	default:
		return nil, fmt.Errorf("unsupported data type for ASCII encoding: %T", data)
	}
}

func (c *CodecASCII) Decode(data []byte) (any, error) {
	return string(data), nil
}

// // DataTyper 数据类型解释器接口
// type DataTyper interface {
// 	Value(data any) any
// 	ExplainedValue(data any, config *ExplainConfig) any
// }

// // BINInteger BIN整数解释器
// type BINInteger struct{}

// func (t *BINInteger) Value(data any) any {
// 	// 这里假设data已经是int类型
// 	return data
// }

// func (t *BINInteger) ExplainedValue(data any, config *ExplainConfig) any {
// 	srcInt := data.(int)

// 	// 应用倍数和偏移量
// 	result := srcInt
// 	if config.multiple != 0 {
// 		result = int(float64(srcInt) * config.multiple)
// 	}
// 	if config.offset != 0 {
// 		result += int(config.offset)
// 	}

// 	// 如果有枚举映射，使用枚举值
// 	if config.enum != nil {
// 		if enumValue, ok := config.enum[result]; ok {
// 			return enumValue
// 		}
// 	}

// 	return result
// }

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

// // BCDString BCD字符串解释器
// type BCDString struct{}

// func (t *BCDString) Value(data any) any {
// 	return data
// }

// func (t *BCDString) ExplainedValue(data any, config *ExplainConfig) any {
// 	return data
// }

// // BCDFloat BCD浮点数解释器
// type BCDFloat struct {
// 	decimal int
// }

// func (t *BCDFloat) Value(data any) any {
// 	return data
// }

// func (t *BCDFloat) ExplainedValue(data any, config *ExplainConfig) any {
// 	srcStr := data.(string)
// 	decimal := t.decimal
// 	if config.decimal != 0 {
// 		decimal = config.decimal
// 	}

// 	// 转换为浮点数
// 	value, err := strconv.ParseFloat(srcStr, 64)
// 	if err != nil {
// 		return srcStr
// 	}

// 	// 应用小数位数
// 	result := value / math.Pow10(decimal)

// 	// 应用倍数和偏移量
// 	if config.multiple != 0 {
// 		result *= config.multiple
// 	}
// 	if config.offset != 0 {
// 		result += config.offset
// 	}

// 	return result
// }

// // ASCIIString ASCII字符串解释器
// type ASCIIString struct{}

// func (t *ASCIIString) Value(data any) any {
// 	return data
// }

// func (t *ASCIIString) ExplainedValue(data any, config *ExplainConfig) any {
// 	return data
// }

// // Decode 解码方法
func (config *FieldCodecConfig) Decode(data []byte) (any, error) {
	if config.mode != ModeDecode {
		return nil, errors.New("config is not in decode mode")
	}

	// 使用编解码器解码
	rawValue, err := config.codec.Decode(data)
	if err != nil {
		return nil, err
	}

	// 使用数据类型解释器解释
	if config.dataTyper != nil {
		return config.dataTyper.ExplainedValue(rawValue), nil
	}

	return rawValue, nil
}

// Encode 编码方法
func (config *FieldCodecConfig) Encode(data any) ([]byte, error) {
	if config.mode != ModeEncode {
		return nil, errors.New("config is not in encode mode")
	}

	// 使用编解码器编码
	return config.codec.Encode(data, config.length)
}
