package main

import (
	"fmt"

	rot "github.com/longan55/Rules-over-TCP"
)

// 起始符 数据长度    数据域  校验
// 0x68  1      N      crc16
//
// 数据单元格式 -》 按功能码解析
//
// 两种方式
// 1.统一的Field结构体，统一的方法
// 2.不同的结构体，同一个接口

func main() {
	TestSrc()
}

func TestSrc() {
	//原始数据
	var sourceData = [][]byte{{0x68}, {0x01}, {0x00}, {0x01}, {0x97}}
	//处理器构建器
	builder := rot.NewBuilder()

	//添加加密配置
	cryptConfig := rot.NewCryptConfig()
	cryptConfig.AddCrypt(0, rot.CryptNone)
	builder.AddCryptConfig(cryptConfig)

	//配置处理器，来数据时自动处理
	handlerConfig := rot.NewHandlerConfig()
	handlerConfig.AddHandler(0x01, rot.HandlerTest)
	builder.AddHandlerConfig(handlerConfig)

	//TODO:
	//配置编码器，用于主动发送数据或被处理器调用

	//构建处理器
	builder.AddFielder(rot.NewStarter([]byte{0x68})).
		AddFielder(rot.NewDataLen(1)).
		AddFielder(rot.NewCyptoFlag()).
		AddFielder(rot.NewFuncCode()).
		AddFielder(rot.NewDataZone())
	dataHander, err := builder.Build()
	if err != nil {
		fmt.Println("构建处理器失败:", err)
		return
	}

	//解析
	if fun, err := dataHander.Parse(sourceData); err != nil {
		fmt.Println("解析失败:", err)
		return
	} else {
		fmt.Println("解析成功:", fun)
	}
}
