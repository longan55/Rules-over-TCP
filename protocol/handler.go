package protocol

import (
	"encoding/binary"
	"fmt"
)

//业务处理函数

func HandlerTest(data []byte) error {
	return nil
}

// 功能码,长度应该更加广泛,但暂时使用1字节
type FunctionCode byte

type Handler func(data []byte) error

type HandlerConfig struct {
	handlerMap map[FunctionCode]Handler
}

func (hc *HandlerConfig) AddHandler(fc FunctionCode, h Handler) {
	hc.handlerMap[fc] = h
}

func NewHandlerConfig() *HandlerConfig {
	return &HandlerConfig{
		handlerMap: make(map[FunctionCode]Handler, 32),
	}
}

type FucntionHandler struct {
	length     int
	fc         FunctionCode
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
func (fh *FucntionHandler) Parse(data []byte) (map[string]any, error) {
	if len(data) != fh.length {
		return nil, fmt.Errorf("data length %d is not equal to function handler length %d", len(data), fh.length)
	}
	//HEX格式：每两个字节之间用空格隔开，前面添加0X，不足2位用0填充
	fmt.Printf("%d bytes,SOURCE: [%# X]\n", len(data), data)
	result := make(map[string]any, len(fh.m))
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

// FunctionHandler Parse() 解析函数，解码所有字段得到总的数据
// Decoder/DecoderImpl     解码函数，
// BIN/BCD/ASCII/CP56TIME2A
// BINInteger/BCDInteger...

type ParsedData struct {
	Bytes     []byte
	Origin    any
	Explained any
}
