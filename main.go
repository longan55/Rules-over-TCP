package main

import "github.com/longan55/proto/protocol"

// 起始符 数据长度    数据域  校验
// 0x68  1      N      crc16
//
// 数据单元格式 -》 按功能码解析
//
// 两种方式
// 1.统一的Field结构体，统一的方法
// 2.不同的结构体，同一个接口

func main() {
	//原始数据
	var sourceData = []byte{0x68, 0x01, 0x01, 0x97}
	//处理器构建器
	builder := protocol.NewBuilder()
	//构建处理器
	builder.AddFielder(protocol.NewStarter([]byte{0x68})).
		AddFielder(protocol.NewDataLen(1))
	dataHander := builder.Build()
	//添加功能
	dataHander.AddFunction(protocol.FunctionCode(0x01), &protocol.FuctionTest{})
	//解析
	dataHander.Parse(sourceData)
}
