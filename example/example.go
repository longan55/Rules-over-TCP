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

	//添加加密配置
	cryptConfig := rot.NewCryptConfig()
	cryptConfig.AddCrypt(0, rot.CryptNone)
	builder.AddCryptConfig(cryptConfig)

	//配置处理器，来数据时自动处理
	setHandlerConfig(builder)

	//TODO:
	//配置编码器，用于主动发送数据或被处理器调用

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

	// 创建FakeConn并设置测试数据
	fakeConn := fake.NewFakeConn()
	fakeConn.SetData([]byte{0x68, 0x0F, 0x00, 0x01, 0x7F, 0xFF, 0xFF, 0xFF, 0x80, 0x00, 0x00, 0x00, 0x12, 0x34, 0x12, 0x34, 0x01, 0x1a, 0x40})
	fakeConn.SetData([]byte{0x68, 0x06, 0x00, 0x02, 0x12, 0x34, 0x56, 0x78, 0x02, 0xef})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 使用fake conn运行Handle方法
	go dataHander.Handle(ctx, fakeConn)

	// 给Handle方法一些时间处理数据
	time.Sleep(100 * time.Millisecond)
}

func setHandlerConfig(builder *rot.ProtocolBuilder) {
	handlerConfig := rot.NewHandlerConfig()
	//1.
	fh := new(rot.FucntionHandler)
	fh.NewDecoder("a", binary.BigEndian).BIN().SetByteLength(4).Integer()
	fh.NewDecoder("b", binary.BigEndian).BIN().SetByteLength(4).Integer()
	fh.NewDecoder("c", binary.BigEndian).BIN().SetByteLength(2).Integer()
	fh.NewDecoder("d", binary.BigEndian).BIN().SetByteLength(2).Float1().Multiple(0.01)
	fh.NewDecoder("e", binary.BigEndian).BIN().SetByteLength(1).Integer().SetEnum(map[int]any{
		0: "A",
		1: "B",
		2: "C",
		3: "D",
	})
	fh.SetHandle(func(parsedData map[string]rot.ParsedData) error {
		fmt.Println("parsedData:", parsedData)
		return nil
	})
	//2.
	fh1 := new(rot.FucntionHandler)
	fh1.NewDecoder("code", binary.BigEndian).BCD().SetByteLength(4).String()
	fh1.SetHandle(func(parsedData map[string]rot.ParsedData) error {
		fmt.Println("parsedData:", parsedData)
		return nil
	})
	//end
	handlerConfig.AddHandler(rot.FunctionCode(0x01), fh)
	handlerConfig.AddHandler(rot.FunctionCode(0x02), fh1)
	builder.AddHandlerConfig(handlerConfig)
}
