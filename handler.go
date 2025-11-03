package rot

import (
	"encoding/binary"
	"fmt"
)

//业务处理函数

var HandlerTest = &FunctionHandler{
	length: 1,
}

// 功能码,长度应该更加广泛,但暂时使用1字节
type FunctionCode byte

// type Handler func(fh *FucntionHandler, data []byte) error
type Handler func(parsed map[string]ParsedData) error

type HandlerConfig struct {
	handlerMap map[FunctionCode]*FunctionHandler
}

func (hc *HandlerConfig) AddHandler(fc FunctionCode, h *FunctionHandler) {
	hc.handlerMap[fc] = h
}

func NewHandlerConfig() *HandlerConfig {
	return &HandlerConfig{
		handlerMap: make(map[FunctionCode]*FunctionHandler, 32),
	}
}

type FunctionHandler struct {
	length     int
	fc         FunctionCode
	handler    Handler
	fieldNames []string
	decoders   []*DecoderImpl
	encoders   []*EncoderImpl

	fccs []*FieldCodecConfig
}

func (fh *FunctionHandler) AddField(fcc *FieldCodecConfig) {
	fh.fccs = append(fh.fccs, fcc)
}

func (fh *FunctionHandler) NewDecoder(fieldName string, order binary.ByteOrder) *DecoderImpl {
	decoder := &DecoderImpl{fh: fh, order: order}
	fh.fieldNames = append(fh.fieldNames, fieldName)
	fh.decoders = append(fh.decoders, decoder)
	return decoder
}

// TODO: 预防数组越界
func (fh *FunctionHandler) Parse(data []byte) (map[string]ParsedData, error) {
	if len(data) != fh.length {
		return nil, fmt.Errorf("data length %d is not equal to function handler length %d", len(data), fh.length)
	}
	//HEX格式：每两个字节之间用空格隔开，前面添加0X，不足2位用0填充
	result := make(map[string]ParsedData, len(fh.decoders))
	offset := 0
	for i, impl := range fh.decoders {
		fieldName := fh.fieldNames[i]
		length := impl.decoder.GetByteLength()
		fmt.Printf("No.%d_FIELD_(%s)_len(%d): [%# X] ", i+1, fieldName, length, data[offset:offset+length])
		input := data[offset : offset+length]
		value, err := impl.Decode(input)
		if err != nil {
			return nil, err
		}
		offset += length
		pd := ParsedData{
			Bytes:     input,
			Origin:    value,
			Explained: impl.decoder.ExplainedValue(value),
		}
		result[fieldName] = pd
		fmt.Printf("DecodeValue:%v\n", value)
	}
	return result, nil
}

func (fh *FunctionHandler) SetHandle(h Handler) error {
	fh.handler = h
	return nil
}

func (fh *FunctionHandler) Handle(data []byte) error {
	if fh.handler == nil {
		return fmt.Errorf("handler is nil")
	}
	parsed, err := fh.Parse(data)
	if err != nil {
		return err
	}
	return fh.handler(parsed)
}

// FunctionHandler Parse() 解析函数，解码所有字段得到总的数据
// Decoder/DecoderImpl     解码函数，
// BIN/BCD/ASCII/CP56TIME2A
// BINInteger/BCDInteger...

type ParsedData struct {
	Bytes     []byte
	Origin    any
	Explained any
}

func (fh *FunctionHandler) NewEncoder(fieldName string, order binary.ByteOrder) *EncoderImpl {
	encoder := &EncoderImpl{fh: fh, order: order}
	fh.fieldNames = append(fh.fieldNames, fieldName)
	fh.encoders = append(fh.encoders, encoder)
	return encoder
}
