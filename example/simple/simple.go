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
	RegisterHandlerConfig(builder)

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

func RegisterHandlerConfig(builder *rot.ProtocolBuilder) {
	builder.HandleFunc(rot.FunctionCode(0x02), ParseHandle02)
	builder.HandleFunc(rot.FunctionCode(0x03), ParseHandle03)

	builder.HandleFunc(rot.FunctionCode(0x01), func(fh *rot.FunctionHandler) {
		fh.AddField("a", rot.WithBin(), rot.WithLength(4), rot.WithInteger(true, 1, 0)).
			AddField("b", rot.WithBin(), rot.WithLength(4), rot.WithInteger(true, 1, 0)).
			AddField("c", rot.WithBin(), rot.WithLength(2), rot.WithInteger(true, 2, 0)).
			AddField("d", rot.WithBin(), rot.WithLength(2), rot.WithFloat(true, 0.01, 0)).
			AddField("e", rot.WithBin(), rot.WithLength(1), rot.WithInteger(true, 1, 0), rot.WithEnum("Other", map[int]any{0: "A", 1: "B", 2: "C"})).
			SetHandler(func(parsedData map[string]rot.ParsedData) error {
				fmt.Println("parsedData:", parsedData)
				return nil
			})
	})
}

func ParseHandle02(fh *rot.FunctionHandler) {
	fh.AddField("code", rot.WithBcd(), rot.WithLength(4), rot.WithString()).
		AddField("price", rot.WithBcd(), rot.WithLength(4), rot.WithFloat(true, 0.0001, 0)).
		AddField("intPrice", rot.WithBcd(), rot.WithLength(4), rot.WithInteger(true, 1, 0)).
		SetHandler(func(parsedData map[string]rot.ParsedData) error {
			fmt.Println("parsedData:", parsedData)
			return nil
		})
}
func ParseHandle03(fh *rot.FunctionHandler) {
	fh.AddField("ascii", rot.WithAscii(), rot.WithLength(4), rot.WithString()).
		SetHandler(func(parsedData map[string]rot.ParsedData) error {
			fmt.Println("parsedData:", parsedData)
			return nil
		})
}
