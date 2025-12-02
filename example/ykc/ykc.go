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
	"github.com/longan55/Rules-over-TCP/example/ykc/handler"
	"github.com/longan55/Rules-over-TCP/fake"
)

var builder *rot.ProtocolBuilder

func main() {
	//处理器构建器
	builder = rot.NewProtocolBuilder()
	//default is BigEndian, if is BigEndian,this can be not called
	builder.SetDefaultOrder(binary.BigEndian)
	//构建处理器
	builder.AddElement(rot.NewStarter([]byte{0x68})).
		AddElement(rot.NewDataLen(1)).
		AddElement(rot.NewSerialNumber()).
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
	handler.RegisterHandlers(builder)

	// 创建FakeConn并设置测试数据
	fakeConn := getConn()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 使用fake conn运行Handle方法
	go dataHander.Serve(ctx, fakeConn)

	// 给Handle方法一些时间处理数据
	time.Sleep(100 * time.Millisecond)
}

func getConn() *fake.FakeConn {
	conn := fake.NewFakeConn()
	conn.SetData([]byte{0x68,
		0x22,
		0x00, 0x00,
		0x00,
		0x01,
		0x55, 0x03, 0x14, 0x12, 0x78, 0x23, 0x05, 0x00, 0x02, 0x0A, 0x56, 0x34, 0x2E, 0x31, 0x2E, 0x35, 0x30, 0x00, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x04,
		0x67, 0x5A})
	return conn
}
