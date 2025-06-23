package main

import (
	"encoding/binary"
	"fmt"

	"github.com/longan55/proto/protocol"
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
	var data = []byte{0x68, 0x01, 0x97}
	//var adu = Adu{}
	//
	////起始值 --保存起始值，判断起始值是否正确
	//var StartCode = &Starter{}
	//StartCode.SetLen(1)
	//StartCode.SetDefaultValue([]byte{0x68})
	//
	////数据长度 --该长度为哪里到哪里的长度
	//var DataLen = &DataLen{}
	//DataLen.SetLen(1)
	//
	////数据域 --应用数据解析
	//var Data = &Datar{}
	//
	//adu.AddField(StartCode)
	//adu.AddField(DataLen)
	//adu.AddField(Data)
	//adu.Info()
	//adu.Debug(nil, data)
	builder := protocol.NewProtoBuilder()
	builder.SetStart(1, 0x68, binary.BigEndian).
		SetDataLength(1, binary.LittleEndian).
		SetData(1, binary.LittleEndian)
	//SetVerify(2, binary.LittleEndian, nil)
	proto := builder.Build()
	m, err := proto.UnWrap(data)
	if err != nil {
		fmt.Printf("unwrap error:%v\n", err)
		return
	}
	fmt.Printf("result:%v\n", m)
}
