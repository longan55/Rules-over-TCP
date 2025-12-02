/*
* Copyright 2025-2026 longan55 or authors.
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*      https://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */
package rot

import (
	"fmt"
	"net"
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
	length  int
	handler Handler

	fccs []*FieldCodecConfig
}

// NewFunctionHandler 创建一个新的FunctionHandler实例
func NewFunctionHandler() *FunctionHandler {
	return &FunctionHandler{
		fccs: make([]*FieldCodecConfig, 0),
	}
}

// AddField 添加字段配置，并返回自身以支持链式调用
func (fh *FunctionHandler) AddField(fieldName string, options ...CodecOption) *FunctionHandler {
	fcc := NewFieldCodecConfig(fieldName, options...)
	fh.fccs = append(fh.fccs, fcc)
	fh.length += fcc.length
	return fh
}

// SetHandler 设置处理函数，并计算总长度
func (fh *FunctionHandler) SetHandler(h Handler) *FunctionHandler {
	fh.handler = h
	return fh
}

func (fh *FunctionHandler) Parse(data []byte) (map[string]ParsedData, error) {
	if len(data) != fh.length {
		return nil, fmt.Errorf("data length %d is not equal to function handler length %d", len(data), fh.length)
	}
	result := make(map[string]ParsedData, len(fh.fccs))
	offset := 0
	for _, impl := range fh.fccs {
		length := impl.length

		// 检查偏移量是否越界
		if offset+length > len(data) {
			return nil, fmt.Errorf("field %s exceeds data bounds", impl.name)
		}

		input := data[offset : offset+length]
		value, err := impl.Decode(input)
		if err != nil {
			return nil, fmt.Errorf("failed to decode field %s: %w", impl.name, err)
		}
		result[impl.name] = *value
		offset += length
	}
	return result, nil
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

// Encode 将字段数据编码为二进制
func (fh *FunctionHandler) Encode(data map[string]any) ([]byte, error) {
	var result []byte
	for _, fcc := range fh.fccs {
		// 切换到编码模式
		originalMode := fcc.mode
		fcc.mode = ModeEncode

		// 确保字段存在
		value, exists := data[fcc.name]
		if !exists {
			fcc.mode = originalMode // 恢复原始模式
			return nil, fmt.Errorf("field %s not found in data", fcc.name)
		}

		// 编码字段
		encoded, err := fcc.Encode(value)
		if err != nil {
			fcc.mode = originalMode // 恢复原始模式
			return nil, fmt.Errorf("failed to encode field %s: %w", fcc.name, err)
		}

		result = append(result, encoded...)
		fcc.mode = originalMode // 恢复原始模式
	}
	return result, nil
}

type ParsedData struct {
	Bytes     []byte
	Origin    any
	Explained any
}

// MessageEncoder 消息编码器接口，用于发送数据时的编码处理
type MessageEncoder struct {
	handler  *FunctionHandler
	funcCode FunctionCode
}

// NewMessageEncoder 创建一个新的消息编码器
func NewMessageEncoder(funcCode FunctionCode) *MessageEncoder {
	return &MessageEncoder{
		funcCode: funcCode,
	}
}

// SetHandler 设置用于编码的FunctionHandler
func (e *MessageEncoder) SetHandler(handler *FunctionHandler) *MessageEncoder {
	e.handler = handler
	return e
}

// Encode 仅编码数据，不发送
func (e *MessageEncoder) Encode(data map[string]any) ([]byte, error) {
	if e.handler == nil {
		return nil, fmt.Errorf("handler not set")
	}

	// 编码负载数据
	payload, err := e.handler.Encode(data)
	if err != nil {
		return nil, err
	}

	// 这里应该实现完整的消息编码逻辑
	// 包括添加起始符、长度、功能码、校验和等
	// 为简化示例，这里仅返回编码后的负载数据
	return payload, nil
}

// EncodeAndSend 编码数据并发送
func (e *MessageEncoder) EncodeAndSend(conn net.Conn, data map[string]any) error {
	// 编码数据
	binaryData, err := e.Encode(data)
	if err != nil {
		return err
	}

	// 发送数据
	_, err = conn.Write(binaryData)
	return err
}

// 创建消息编码器的便捷方法
func CreateMessageEncoder(funcCode FunctionCode, handler *FunctionHandler) *MessageEncoder {
	return NewMessageEncoder(funcCode).SetHandler(handler)
}
