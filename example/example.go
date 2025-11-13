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
	go dataHander.Handle(ctx, fakeConn)

	// 给Handle方法一些时间处理数据
	time.Sleep(100 * time.Millisecond)
}

func setHandlerConfig(builder *rot.ProtocolBuilder) {
	handlerConfig := rot.NewHandlerConfig()
	//1. BIN编码,默认解释为整数，强烈建议只解释为整数或浮点数，不要解释为字符串。
	fh := new(rot.FunctionHandler)
	fh.AddField("a", rot.WithBin(), rot.WithLength(4), rot.WithInteger(true, 1, 0))
	fh.AddField("b", rot.WithBin(), rot.WithLength(4), rot.WithInteger(true, 1, 0))
	fh.AddField("c", rot.WithBin(), rot.WithLength(2), rot.WithInteger(true, 2, 0))
	fh.AddField("d", rot.WithBin(), rot.WithLength(2), rot.WithFloat(true, 0.01, 0))
	fh.AddField("e", rot.WithBin(), rot.WithLength(1), rot.WithInteger(true, 1, 0), rot.WithEnum("Other", map[int]any{0: "A", 1: "B", 2: "C"}))

	fh.SetHandle(func(parsedData map[string]rot.ParsedData) error {
		fmt.Println("parsedData:", parsedData)
		return nil
	})
	//2. BCD编码，默认解释为字符串，还可以解释为整数或浮点数，浮点数较为常见（在需要高精度传输时）
	fh1 := new(rot.FunctionHandler)
	fh1.AddField("code", rot.WithBcd(), rot.WithLength(4), rot.WithString())
	fh1.AddField("price", rot.WithBcd(), rot.WithLength(4), rot.WithFloat(true, 0.0001, 0))
	fh1.AddField("intPrice", rot.WithBcd(), rot.WithLength(4), rot.WithInteger(true, 1, 0))

	fh1.SetHandle(func(parsedData map[string]rot.ParsedData) error {
		fmt.Println("parsedData:", parsedData)
		return nil
	})
	//3. ASCII编码，仅解释为字符串
	fh2 := new(rot.FunctionHandler)
	fh2.AddField("ascii", rot.WithAscii(binary.BigEndian), rot.WithLength(4), rot.WithString())
	fh2.SetHandle(func(parsedData map[string]rot.ParsedData) error {
		fmt.Println("parsedData:", parsedData)
		return nil
	})
	//end
	handlerConfig.AddHandler(rot.FunctionCode(0x01), fh)
	handlerConfig.AddHandler(rot.FunctionCode(0x02), fh1)
	handlerConfig.AddHandler(rot.FunctionCode(0x03), fh2)
	builder.AddHandlerConfig(handlerConfig)
}
