package rot

import (
	"encoding/binary"
)

// CodecOption 配置选项接口
type CodecOption interface {
	Apply(config *FieldCodecConfig)
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

func WithAscii() CodecOption {
	return &codecOption{codec: &CodecASCII{}}
}

// WithLength 设置字段长度选项
func WithLength(length int) CodecOption {
	return &lengthOption{length}
}

// WithDataTyper 设置数据类型解释器选项
func WithDataTyper(dataTyper DataTyper) CodecOption {
	return &dataTyperOption{dataTyper}
}

func WithInteger(moflag bool, multiple int, offset int) CodecOption {
	return &dtInteger{moflag: moflag, multiple: multiple, offset: offset}
}

func WithFloat(moflag bool, multiple float64, offset float64) CodecOption {
	return &dtFloat{moflag: moflag, multiple: multiple, offset: offset}
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
