package rot

import (
	"encoding/binary"
	"fmt"
)

//业务处理函数

var HandlerTest = &FucntionHandler{
	length: 1,
}

// 功能码,长度应该更加广泛,但暂时使用1字节
type FunctionCode byte

type Handler func(fh *FucntionHandler, data []byte) error

type HandlerConfig struct {
	handlerMap map[FunctionCode]*FucntionHandler
}

func (hc *HandlerConfig) AddHandler(fc FunctionCode, h *FucntionHandler) {
	hc.handlerMap[fc] = h
}

func NewHandlerConfig() *HandlerConfig {
	return &HandlerConfig{
		handlerMap: make(map[FunctionCode]*FucntionHandler, 32),
	}
}

type FucntionHandler struct {
	length     int
	fc         FunctionCode
	handler    Handler
	fieldNames []string
	m          []*DecoderImpl
}

func (fh *FucntionHandler) NewDecoder(fieldName string, order binary.ByteOrder) *DecoderImpl {
	decoder := &DecoderImpl{fh: fh, order: order}
	fh.fieldNames = append(fh.fieldNames, fieldName)
	fh.m = append(fh.m, decoder)
	return decoder
}

// TODO: 预防数组越界
func (fh *FucntionHandler) Parse(data []byte) (map[string]ParsedData, error) {
	if len(data) != fh.length {
		return nil, fmt.Errorf("data length %d is not equal to function handler length %d", len(data), fh.length)
	}
	//HEX格式：每两个字节之间用空格隔开，前面添加0X，不足2位用0填充
	fmt.Printf("%d bytes,SOURCE: [%# X]\n", len(data), data)
	result := make(map[string]ParsedData, len(fh.m))
	offset := 0
	for i, impl := range fh.m {
		fieldName := fh.fieldNames[i]
		length := impl.decoder.GetByteLength()
		fmt.Printf("FIELD(%s)_%d(%d): [%# X] ", fieldName, i+1, length, data[offset:offset+length])
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

func (fh *FucntionHandler) SetHandle(h Handler) error {
	fh.handler = h
	return nil
}

func (fh *FucntionHandler) Handle(data []byte) error {
	if fh.handler == nil {
		return fmt.Errorf("handler is nil")
	}
	return fh.handler(fh, data)
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
