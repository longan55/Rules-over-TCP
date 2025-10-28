package main

import (
	"encoding/binary"
	"fmt"

	rot "github.com/longan55/Rules-over-TCP"
)

func main() {
	//原始数据
	var sourceData = [][]byte{{0x68},
		{0x0F},
		{0x00},
		{0x01},
		{0x7F, 0xFF, 0xFF, 0xFF, 0x80, 0x00, 0x00, 0x00, 0x12, 0x34, 0x12, 0x34, 0x01},
		{0xC0, 0xA7}}
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
		AddElement(rot.NewCheckSum(1, 2))
	dataHander, err := builder.Build()
	if err != nil {
		fmt.Println("构建处理器失败:", err)
		return
	}

	//解析
	if err := dataHander.Parse(sourceData); err != nil {
		fmt.Println("解析失败:", err)
		return
	} else {
		fmt.Println("解析成功!")
	}
}

func setHandlerConfig(builder *rot.ProtocolBuilder) {
	handlerConfig := rot.NewHandlerConfig()
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
	handlerConfig.AddHandler(rot.FunctionCode(0x01), fh)
	builder.AddHandlerConfig(handlerConfig)
}
