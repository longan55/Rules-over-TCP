package rot

import (
	"encoding/binary"
	"errors"
	"fmt"
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
			return parsed, nil
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

// WithLength 设置字段长度选项
func WithLength(length int) CodecOption {
	return &lengthOption{length}
}

// WithMode 设置编解码模式选项
func WithMode(mode CodecMode) CodecOption {
	return &modeOption{mode}
}

func WithEncode() CodecOption {
	return &modeOption{mode: ModeEncode}
}

func WithDecode() CodecOption {
	return &modeOption{mode: ModeDecode}
}

// WithCodec 设置编解码器选项
func WithCodec(codec Codec) CodecOption {
	return &codecOption{codec}
}

func WithBin() CodecOption {
	return &codecOption{codec: &CodecBIN{
		order: DefaultOrder(),
	}}
}

func WithBinWithOrder(order binary.ByteOrder) CodecOption {
	return &codecOption{codec: &CodecBIN{
		order: order,
	}}
}

func WithBcd() CodecOption {
	return &codecOption{codec: &CodecBCD{
		order: DefaultOrder(),
	}}
}

func WithBcdWithOrder(order binary.ByteOrder) CodecOption {
	return &codecOption{codec: &CodecBCD{
		order: order,
	}}
}

func WithBcdInteger(moflag bool, multiple int, offset int) CodecOption {
	return &bcdInteger{moflag: moflag, multiple: multiple, offset: offset}
}
func WithBcdFloat(moflag bool, multiple float64, offset float64) CodecOption {
	return &bcdFloat{moflag: moflag, multiple: multiple, offset: offset}
}

func WithAscii(order binary.ByteOrder) CodecOption {
	return &codecOption{codec: &CodecASCII{}}
}

// WithDataTyper 设置数据类型解释器选项
func WithDataTyper(dataTyper DataTyper) CodecOption {
	return &dataTyperOption{dataTyper}
}

// WithMultiple 设置倍数选项
func WithMultiple(multiple float64) CodecOption {
	return &multipleOption{multiple}
}

// WithOffset 设置偏移量选项
func WithOffset(offset float64) CodecOption {
	return &offsetOption{offset}
}

// WithEnum 设置枚举映射选项
func WithEnum(other string, enum map[int]any) CodecOption {
	return &enumOption{other, enum}
}
func WithBitmap(bitmap map[int]any) CodecOption {
	return &bitmapOption{bitmap}
}

type lengthOption struct {
	length int
}

func (o *lengthOption) Apply(config *FieldCodecConfig) {
	config.length = o.length
}

type modeOption struct {
	mode CodecMode
}

func (o *modeOption) Apply(config *FieldCodecConfig) {
	config.mode = o.mode
}

type codecOption struct {
	codec Codec
}

func (o *codecOption) Apply(config *FieldCodecConfig) {
	o.codec.Configure() // 初始化编解码器
	config.codec = o.codec
}

type dataTyperOption struct {
	dataTyper DataTyper
}

func (o *dataTyperOption) Apply(config *FieldCodecConfig) {
	config.dataTyper = o.dataTyper
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

type offsetOption struct {
	offset float64
}

func (o *offsetOption) Apply(config *FieldCodecConfig) {
	if config.explainConfig == nil {
		config.explainConfig = &ExplainConfig{}
	}
	config.explainConfig.offset = o.offset
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

type bitmapOption struct {
	bitmap map[int]any
}

func (o *bitmapOption) Apply(config *FieldCodecConfig) {
	if config.explainConfig == nil {
		config.explainConfig = &ExplainConfig{}
	}
	config.explainConfig.bitmap = o.bitmap
}
