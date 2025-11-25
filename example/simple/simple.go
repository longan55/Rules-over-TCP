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
package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	rot "github.com/longan55/Rules-over-TCP"
	fake "github.com/longan55/Rules-over-TCP/fake"
)

func main() {
	//处理器构建器
	builder := rot.NewProtocolBuilder()
	//default is BigEndian, if is BigEndian,this can be not called
	builder.SetDefaultOrder(binary.BigEndian)
	//构建处理器
	builder.AddElement(rot.NewStarter([]byte{0x68})).
		AddElement(rot.NewDataLen(1)).
		AddElement(rot.NewCyptoFlag()).
		AddElement(rot.NewFuncCode()).
		AddElement(rot.NewPayload()).
		AddElement(rot.NewCheckSum(0, 2))
	dataHander, err := builder.Build()
	if err != nil {
		fmt.Println("构建处理器失败:", err)
		return
	}

	//添加加密配置
	cryptConfig := rot.NewCryptConfig()
	builder.AddCryptConfig(cryptConfig)

	//配置处理器，来数据时自动处理
	setHandlerConfig(builder)

	// 创建FakeConn并设置测试数据
	fakeConn := fake.NewFakeConn()
	fakeConn.SetData([]byte{0x68, 0x0F, 0x00, 0x01, 0x7F, 0xFF, 0xFF, 0xFF, 0x80, 0x00, 0x00, 0x00, 0x12, 0x34, 0x12, 0x34, 0x01, 0x1a, 0x40})
	fakeConn.SetData([]byte{0x68, 0x0E, 0x00, 0x02, 0x12, 0x34, 0x56, 0x78, 0x12, 0x34, 0x56, 0x78, 0x12, 0x34, 0x56, 0x78, 0xbd, 0x09})
	fakeConn.SetData([]byte{0x68, 0x06, 0x00, 0x03, 0x30, 0x31, 0x32, 0x33, 0x4f, 0xa1})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 使用fake conn运行Handle方法
	go dataHander.Serve(ctx, fakeConn)

	// 给Handle方法一些时间处理数据
	time.Sleep(100 * time.Millisecond)
}

func setHandlerConfig(builder *rot.ProtocolBuilder) {
	// 原始方法 - 使用HandlerConfig
	handlerConfig := rot.NewHandlerConfig()

	// 1. BIN编码 - 使用新的API接口和链式调用
	handler1 := rot.NewFunctionHandler().
		AddField("a", rot.WithBin(), rot.WithLength(4), rot.WithInteger(true, 1, 0)).
		AddField("b", rot.WithBin(), rot.WithLength(4), rot.WithInteger(true, 1, 0)).
		AddField("c", rot.WithBin(), rot.WithLength(2), rot.WithInteger(true, 2, 0)).
		AddField("d", rot.WithBin(), rot.WithLength(2), rot.WithFloat(true, 0.01, 0)).
		AddField("e", rot.WithBin(), rot.WithLength(1), rot.WithInteger(true, 1, 0), rot.WithEnum("Other", map[int]any{0: "A", 1: "B", 2: "C"})).
		SetHandler(func(parsedData map[string]rot.ParsedData) error {
			fmt.Println("parsedData:", parsedData)
			return nil
		})

	// 2. BCD编码 - 使用链式调用
	handler2 := rot.NewFunctionHandler().
		AddField("code", rot.WithBcd(), rot.WithLength(4), rot.WithString()).
		AddField("price", rot.WithBcd(), rot.WithLength(4), rot.WithFloat(true, 0.0001, 0)).
		AddField("intPrice", rot.WithBcd(), rot.WithLength(4), rot.WithInteger(true, 1, 0)).
		SetHandler(func(parsedData map[string]rot.ParsedData) error {
			fmt.Println("parsedData:", parsedData)
			return nil
		})

	// 3. ASCII编码 - 使用链式调用
	handler3 := rot.NewFunctionHandler().
		AddField("ascii", rot.WithAscii(), rot.WithLength(4), rot.WithString()).
		SetHandler(func(parsedData map[string]rot.ParsedData) error {
			fmt.Println("parsedData:", parsedData)
			return nil
		})

	// 注册处理器
	handlerConfig.AddHandler(rot.FunctionCode(0x01), handler1)
	handlerConfig.AddHandler(rot.FunctionCode(0x02), handler2)
	handlerConfig.AddHandler(rot.FunctionCode(0x03), handler3)
	builder.AddHandlerConfig(handlerConfig)

	// 新方法1: 使用HandleFunc - 类似http处理函数注册
	builder.HandleFunc(rot.FunctionCode(0x04), func(parsedData map[string]rot.ParsedData) error {
		fmt.Println("使用HandleFunc注册的处理器:", parsedData)
		return nil
	})

	// 新方法2: 使用HandleFuncWithFields - 带字段定义的处理函数注册
	// 定义一个字段配置函数
	addBinField := func(name string, length int) func(*rot.FunctionHandler) {
		return func(fh *rot.FunctionHandler) {
			fh.AddField(name, rot.WithBin(), rot.WithLength(length), rot.WithInteger(true, 1, 0))
		}
	}

	// 直接在注册时定义字段结构
	builder.HandleFuncWithFields(
		rot.FunctionCode(0x05),
		func(parsedData map[string]rot.ParsedData) error {
			fmt.Println("使用HandleFuncWithFields注册的处理器:", parsedData)
			return nil
		},
		addBinField("field1", 2),
		addBinField("field2", 4),
		func(fh *rot.FunctionHandler) {
			fh.AddField("text", rot.WithAscii(), rot.WithLength(10), rot.WithString())
		},
	)

	fmt.Println("Handler注册完成，可使用NewMessageEncoder进行消息编码发送")
	fmt.Println("新增功能: 可使用HandleFunc和HandleFuncWithFields类似http方式注册处理函数")
}
