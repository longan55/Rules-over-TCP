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

func (fh *FunctionHandler) AddField(fieldName string, options ...CodecOption) *FieldCodecConfig {
	fcc := NewFieldCodecConfig(fieldName, options...)
	fh.fccs = append(fh.fccs, fcc)
	return fcc
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
	parsed, err := fh.Parse(data)
	if err != nil {
		return err
	}
	return fh.handler(parsed)
}

type ParsedData struct {
	Bytes     []byte
	Origin    any
	Explained any
}
