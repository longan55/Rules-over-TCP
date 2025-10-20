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
	fmt.Printf("function handler length:%d\n", fh.length)
	result := make(map[string]any, len(fh.m))
	offset := 0
	for i, decoder := range fh.m {
		input := data[offset : offset+decoder.bin.GetByteLength()]
		value, err := decoder.Decode(input)
		if err != nil {
			return nil, err
		}
		result[fh.fieldNames[i]] = value
		offset += decoder.bin.GetByteLength()
	}
	return result, nil
}

// FunctionHandler Parse() 解析函数，解码所有字段得到总的数据
// Decoder/DecoderImpl     解码函数，
// BIN/BCD/ASCII/CP56TIME2A
// BINInteger/BCDInteger...
