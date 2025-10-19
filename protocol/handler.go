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
	fc FunctionCode
	h  Handler
}

func (fh *FucntionHandler) NewDecoder(order binary.ByteOrder) *DecodeBuilder {
	return &DecodeBuilder{order: order}
}
