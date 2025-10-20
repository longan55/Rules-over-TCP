package protocol

import "encoding/binary"

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
	fc         FunctionCode
	fieldNames []string
	m          []*DecoderImpl
}

func (fh *FucntionHandler) NewDecoder(fieldName string, order binary.ByteOrder) *DecoderImpl {
	decoder := &DecoderImpl{order: order}
	fh.fieldNames = append(fh.fieldNames, fieldName)
	fh.m = append(fh.m, decoder)
	return decoder
}

func (fh *FucntionHandler) Parse(data []byte) (map[string]any, error) {
	result := make(map[string]any)
	for i, decoder := range fh.m {
		value, err := decoder.Decode(data)
		if err != nil {
			return nil, err
		}
		result[fh.fieldNames[i]] = value
	}
	return result, nil
}
