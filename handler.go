package rot

import (
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

	fccs []*FieldCodecConfig
}

func (fh *FunctionHandler) AddField(fieldName string, options ...CodecOption) *FieldCodecConfig {
	fh.fieldNames = append(fh.fieldNames, fieldName)
	fcc := NewFieldCodecConfig(fieldName, options...)
	fh.fccs = append(fh.fccs, fcc)
	return fcc
}

func (fh *FunctionHandler) Parse2(data []byte) (map[string]ParsedData, error) {
	if len(data) != fh.length {
		return nil, fmt.Errorf("data length %d is not equal to function handler length %d", len(data), fh.length)
	}
	result := make(map[string]ParsedData, len(fh.fccs))
	offset := 0
	for i, impl := range fh.fccs {
		fieldName := fh.fieldNames[i]
		length := impl.length
		input := data[offset : offset+length]
		value, err := impl.Decode(input)
		if err != nil {
			return nil, err
		}
		result[fieldName] = *value
	}
	return result, nil
}

func (fh *FunctionHandler) SetHandle(h Handler) error {
	for _, fcc := range fh.fccs {
		fh.length += fcc.length
	}
	fh.handler = h
	return nil
}

func (fh *FunctionHandler) Handle(data []byte) error {
	if fh.handler == nil {
		return fmt.Errorf("handler is nil")
	}
	parsed, err := fh.Parse2(data)
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
